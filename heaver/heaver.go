package heaver

import (
	"heaverd-ng/libstats/lxc"
	"log"
	"os/exec"
	"regexp"
)

var (
	createArgs = []string{"heaver", "-Cn", "", "-i", "virtubuntu", "--net", "auto"}
	reIp       = regexp.MustCompile("(\\d{1,3}.\\d{1,3}.\\d{1,3}.\\d{1,3})")
)

func Create(containerName string) lxc.Container {
	createArgs[2] = containerName
	cmd := exec.Cmd{
		Path: "/usr/bin/heaver",
		Args: createArgs,
	}

	answer, err := cmd.Output()
	if err != nil {
		log.Println("[error]", err)
		return lxc.Container{}
	}

	ip := ""
	matches := reIp.FindStringSubmatch(string(answer))
	if matches != nil {
		ip = matches[1]
	}

	container := lxc.Container{
		Name: containerName,
		Ip:   ip,
	}

	return container
}
