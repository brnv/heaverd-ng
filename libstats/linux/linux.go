package linux

import (
	"heaverd-ng/libstats/lxc"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	memCapacityRe = regexp.MustCompile(`MemTotal:\s+(\d+)`)
	memFreeRe     = regexp.MustCompile(`MemFree:\s+(\d+)`)
)

func Memory() (capacity int, usage int, err error) {
	meminfo, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0, err
	}

	capacity, err = strconv.Atoi(memCapacityRe.FindStringSubmatch(string(meminfo))[1])
	if err != nil {
		return 0, 0, err
	}

	free, err := strconv.Atoi(memFreeRe.FindStringSubmatch(string(meminfo))[1])
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

func Uptime() (int, error) {
	var info syscall.Sysinfo_t
	err := syscall.Sysinfo(&info)
	if err != nil {
		return 0, err
	}
	return int(info.Uptime), nil
}

func HostName() (hostname string, err error) {
	hostname, err = os.Hostname()
	if err != nil {
		return "", err
	}

	return
}

func NetAddr() (netaddr []string, err error) {
	hostname, err := HostName()
	if err != nil {
		return
	}
	netaddr, err = net.LookupHost(hostname)
	if err != nil {
		return
	}
	return
}
