package zfs

import (
	"fmt"
	"os"
)

func ArcMax() (int, error) {
	file, err := os.Open("/sys/module/zfs/parameters/zfs_arc_max")
	if err != nil {
		return 0, err
	}

	arcMaxBytes := 0
	fmt.Fscanf(file, "%d", &arcMaxBytes)
	file.Close()

	return arcMaxBytes / 1024, err
}
