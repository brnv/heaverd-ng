package heaver

import (
	"encoding/json"
	"heaverd-ng/libstats/lxc"
	"os/exec"
)

func Create(containerName string) string {
	cmd := exec.Command("heaver", "-CSn", containerName, "-i", "virtubuntu")
	_, err := cmd.Output()
	if err != nil {

	}
	container := lxc.Container{
		Name: containerName,
	}

	result, err := json.Marshal(container)
	if err != nil {

	}

	return string(result)
}
