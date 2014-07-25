package linux

import (
	"heaverd-ng/libstats/lxc"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func Memory() (capacity int, usage int, err error) {
	cmd := exec.Command("grep", "MemTotal", "/proc/meminfo")
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	capacity, err = strconv.Atoi(strings.Fields(string(out))[1])
	if err != nil {
		return 0, 0, err
	}

	cmd = exec.Command("grep", "MemFree", "/proc/meminfo")
	out, err = cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	free, err := strconv.Atoi(strings.Fields(string(out))[1])
	if err != nil {
		return 0, 0, err
	}

	usage = capacity - free

	return capacity, usage, err
}

func Cpu() (capacity int, usage int, err error) {
	capacity = runtime.NumCPU() * 100

	ticks1, err := lxc.CpuTicks()
	if err != nil {
		return 0, 0, err
	}
	time.Sleep(time.Second)
	ticks2, err := lxc.CpuTicks()
	if err != nil {
		return 0, 0, err
	}

	usage = ticks2 - ticks1

	return capacity, usage, err

}

func Disk() (capacity int, usage int, err error) {
	cmd := exec.Command("df", "-P", "/")
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	capacity, err = strconv.Atoi(strings.Fields(string(out))[8])
	if err != nil {
		return 0, 0, err

	}

	usage, err = strconv.Atoi(strings.Fields(string(out))[10])
	if err != nil {
		return 0, 0, err

	}

	return capacity, usage, err
}

func Uptime() (uptime int64, err error) {
	var info syscall.Sysinfo_t
	err = syscall.Sysinfo(&info)
	if err != nil {
		return 0, err
	}
	uptime = info.Uptime
	return uptime, err
}

func HostName() (hostname string, err error) {
	hostname, err = os.Hostname()
	if err != nil {
		return "", err
	}

	return hostname, err
}
