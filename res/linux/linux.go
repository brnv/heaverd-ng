package linux

import (
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// obtaining ram capacity and usage in kb
func GetRamStats() (capacity, usage int, err error) {
	cmd := exec.Command("free", "-k")
	out, err := cmd.Output()

	if err != nil {
		return 0, 0, err
	}

	result, err := strconv.Atoi(strings.Fields(string(out))[7])
	if err != nil {
		return 0, 0, err
	}
	capacity = result

	result, err = strconv.Atoi(strings.Fields(string(out))[9])
	if err != nil {
		return 0, 0, err
	}
	usage = result

	return
}

// get current cpu ticks
func GetCpuTicks() (int, error) {
	data, err := ioutil.ReadFile("/sys/fs/cgroup/cpu/lxc/cpuacct.stat")
	if err != nil {
		return 0, err
	}

	user, err := strconv.Atoi(strings.Fields(string(data))[1])
	if err != nil {
		return 0, err
	}

	system, err := strconv.Atoi(strings.Fields(string(data))[3])
	if err != nil {
		return 0, err
	}

	return user + system
}

// obtaining cpu capacity and usage in %
// NOTE: function uses "sleep" for second and returns in % how
// many ticks occurs from max.
func GetCpuStats() (capacity, usage int, err error) {
	// Nuber of cpu cores * 100%

	capacity = runtime.NumCPU() * 100

	ticks1, err := GetCpuTicks()
	if err != nil {
		return 0, 0, err
	}
	time.Sleep(time.Second)
	ticks2, err := GetCpuTicks()
	if err != nil {
		return 0, 0, err
	}

	usage = ticks2 - ticks1

	return
}

// obtaining disk capacity and usage in kb
func GetDiskStats() (capacity, usage int, err error) {
	cmd := exec.Command("df", "-P", "/")
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	result, err := strconv.Atoi(strings.Fields(string(out))[8])
	if err != nil {
		return 0, 0, err

	}
	capacity = result

	result, err = strconv.Atoi(strings.Fields(string(out))[10])
	if err != nil {
		return 0, 0, err

	}
	usage = result

	return
}

func GetUptime() (int, error) {
	var info syscall.Sysinfo_t
	err := syscall.Sysinfo(&info)
	if err != nil {
		return 0, err
	}
	return int(info.Uptime), nil
}

func GetHostName() (hostname string, err error) {
	hostname, err = os.Hostname()
	if err != nil {
		return 0, err
	}

	return
}
