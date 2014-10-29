package linux

import (
	"os/exec"

	"github.com/brnv/go-lxc"

	"net"
	"os"
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

const CpuUsageTimeRangeSec = 300

func startCpuMeasure() CPUMeasure {
	ch := CPUMeasure{}
	ch.usage = make(chan int)
	ch.err = make(chan error)

	go func() {
		usageAcc := make([]int, CpuUsageTimeRangeSec)
		index := 0
		for {
			ticksBeforeSleep, err := lxc.GetCpuTicks()
			if err != nil {
				ch.err <- err
				continue
			}
			time.Sleep(time.Second)
			ticksAfterSleep, err := lxc.GetCpuTicks()
			if err != nil {
				ch.err <- err
				continue
			}
			usageAcc[index] = ticksAfterSleep - ticksBeforeSleep
			usageAvg := 0
			for _, usage := range usageAcc {
				usageAvg += usage
			}
			ch.usage <- int(usageAvg / CpuUsageTimeRangeSec)
			if index == CpuUsageTimeRangeSec-1 {
				index = 0
			} else {
				index++
			}
		}
	}()

	return ch
}

func Memory() (capacity int, free int, err error) {
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
	free, err = strconv.Atoi(strings.Fields(string(out))[1])
	if err != nil {
		return 0, 0, err
	}
	cmd = exec.Command("grep", "Cached", "/proc/meminfo")
	out, err = cmd.Output()
	if err != nil {
		return 0, 0, err
	}
	cached, err := strconv.Atoi(strings.Fields(string(out))[1])
	if err != nil {
		return 0, 0, err
	}
	free = free + cached
	return capacity, free, nil
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

func Disk() (capacity int, free int, err error) {
	cmd := exec.Command("df", "-P", "/")
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}
	capacity, err = strconv.Atoi(strings.Fields(string(out))[8])
	if err != nil {
		return 0, 0, err
	}
	free, err = strconv.Atoi(strings.Fields(string(out))[10])
	if err != nil {
		return 0, 0, err
	}
	return capacity, free, nil
}

func Uptime() (uptime int64, err error) {
	var info syscall.Sysinfo_t
	err = syscall.Sysinfo(&info)
	if err != nil {
		return 0, err
	}
	uptime = info.Uptime
	return uptime, nil
}

func HostName() (hostname string, err error) {
	hostname, err = os.Hostname()
	if err != nil {
		return "", err
	}

	return hostname, nil
}

func NetAddr() (netaddr []string, err error) {
	hostname, err := HostName()
	if err != nil {
		return []string{}, err
	}
	// FIXME is it the right way?
	netaddr, err = net.LookupHost(hostname)
	if err != nil {
		return []string{}, err
	}
	return netaddr, nil
}
