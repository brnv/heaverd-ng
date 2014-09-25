package heaver

import (
	"errors"
	"heaverd-ng/libstats/lxc"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

var (
	createArgs  = []string{"heaver", "-Cn", "", "-i", "", "--net", "br0"}
	controlArgs = []string{"heaver", "", ""}
	startArg    = "-Sn"
	stopArg     = "-Tn"
	destroyArg  = "-Dn"
	reIp        = regexp.MustCompile("(\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3})")
	reStarted   = regexp.MustCompile("started")
	reStopped   = regexp.MustCompile("stopped")
	reDestroyed = regexp.MustCompile("destroyed")
	reList      = regexp.MustCompile(`\s*([\d\w-\.]*):\s([a-z]*).*:\s([\d\.]*)/`)
)

func Create(containerName string, image string) (lxc.Container, error) {
	createArgs[2] = containerName
	createArgs[4] = image

	cmd := getHeaverCmd(createArgs)
	output, err := cmd.Output()
	if err != nil {
		return lxc.Container{Status: "error"}, errors.New(string(output))
	}

	ip := ""
	matches := reIp.FindStringSubmatch(string(output))
	if matches != nil {
		ip = matches[1]
	}

	container := lxc.Container{
		Name:   containerName,
		Status: "created",
		Image:  image,
		Ip:     ip,
	}

	return container, nil
}

func Control(containerName string, action string) bool {
	var reControl *regexp.Regexp
	switch action {
	case "start":
		controlArgs[1] = startArg
		reControl = reStarted
	case "stop":
		controlArgs[1] = stopArg
		reControl = reStopped
	case "destroy":
		controlArgs[1] = destroyArg
		reControl = reDestroyed
	}
	controlArgs[2] = containerName

	answer, err := getHeaverCmd(controlArgs).Output()
	if err != nil {
		log.Println("[error]", err)
		return false
	}

	matches := reControl.FindStringSubmatch(string(answer))
	if matches == nil {
		return false
	}

	log.Println("container", containerName, matches[0])

	return true
}

func List(host string) (map[string]lxc.Container, error) {
	cmd := exec.Command("heaver", "-L")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	list := make(map[string]lxc.Container)
	containers := strings.Split(string(output), "\n")
	for _, container := range containers {
		parsed := reList.FindStringSubmatch(container)
		if parsed != nil {
			name := parsed[1]
			list[name] = lxc.Container{
				Name:   name,
				Host:   host,
				Status: parsed[2],
				Ip:     parsed[3],
			}
		}
	}

	return list, nil
}

func getHeaverCmd(args []string) *exec.Cmd {
	cmd := &exec.Cmd{
		Path: "/usr/bin/heaver",
		Args: args,
	}
	return cmd
}
