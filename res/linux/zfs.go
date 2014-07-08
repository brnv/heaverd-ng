package zfs

import (
	"fmt"
	"os"
)

// obtaining zfs arc cache max value
func GetZfsArcMaxValue() (int, error) {
	var num int
	fi, err := os.Open("/sys/module/zfs/parameters/zfs_arc_max")
	if err != nil {
		return 0, err
	}
	fmt.Fscanf(fi, "%d", &num)
	fi.Close()
	return num
}
