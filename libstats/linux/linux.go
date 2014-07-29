package linux

import (
	"errors"
	"fmt"
	"heaverd-ng/libstats/lxc"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var cpuMeasure = startCpuMeasure()

type CPUMeasure struct {
	usage chan int
	err   chan error
}

func startCpuMeasure() CPUMeasure {
	ch := CPUMeasure{}
	ch.usage = make(chan int)
	ch.err = make(chan error)
	go func() {
		for {
			ticksBeforeSleep, err := lxc.CpuTicks()
			if err != nil {
				ch.err <- err
				continue
			}
			time.Sleep(time.Second)
			ticksAfterSleep, err := lxc.CpuTicks()
			if err != nil {
				ch.err <- err
				continue
			}
			usage := ticksBeforeSleep - ticksAfterSleep
			ch.usage <- usage
		}
	}()
	return ch
}

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
	if usage < 0 {
		return 0, 0, errors.New(
			fmt.Sprintf("Free memory value is bigger than total capacity"))
	}
	return capacity, usage, err
}

func Cpu() (capacity int, usage int, err error) {
	capacity = runtime.NumCPU() * 100
	select {
	case usage = <-cpuMeasure.usage:
	case err = <-cpuMeasure.err:
	default:
	}
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
