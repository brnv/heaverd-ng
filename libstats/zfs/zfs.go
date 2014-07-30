package zfs

import (
	"os/exec"
	"strconv"
	"strings"
)

func ArcMax() (int, error) {
	cmd := exec.Command("grep", "c_max", "/proc/spl/kstat/zfs/arcstats")
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	arcMax, err := strconv.Atoi(strings.Fields(string(out))[2])
	if err != nil {
		return 0, err
	}
	return arcMax, nil
}
