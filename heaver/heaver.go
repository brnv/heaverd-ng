package heaver

import "os/exec"

func Create(containerName string) string {
	cmd := exec.Command("heaver", "-Cn", containerName, "-i", "virtubuntu")
	result, _ := cmd.Output()

	return string(result)
}
