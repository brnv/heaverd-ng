package heaver

import (
	"heaverd-ng/libstats/lxc"
	"log"
	"os/exec"
	"regexp"
)

var (
	createArgs  = []string{"heaver", "-Cn", "", "-i", "virtubuntu", "--net", "auto"}
	controlArgs = []string{"heaver", "", ""}
	startArg    = "-Sn"
	stopArg     = "-Tn"
	destroyArg  = "-Dn"
	reIp        = regexp.MustCompile("(\\d{1,3}.\\d{1,3}.\\d{1,3}.\\d{1,3})")
	reStarted   = regexp.MustCompile("started")
	reStopped   = regexp.MustCompile("stopped")
	reDestroyed = regexp.MustCompile("destroyed")
)

func Create(containerName string) lxc.Container {
	createArgs[2] = containerName

	output, err := getHeaverCmd(createArgs).Output()
	if err != nil {
		log.Println("[error]", err)
		return lxc.Container{}
	}

	ip := ""
	matches := reIp.FindStringSubmatch(string(output))
	if matches != nil {
		ip = matches[1]
	}

	container := lxc.Container{
		Name: containerName,
		Ip:   ip,
	}

	return container
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

	log.Println(containerName, action)

	return true
}

func getHeaverCmd(args []string) *exec.Cmd {
	cmd := &exec.Cmd{
		Path: "/usr/bin/heaver",
		Args: args,
	}
	return cmd
}
