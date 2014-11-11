package linux

import (
	"io/ioutil"
	"os/exec"

	"github.com/brnv/go-lxc"

	"net"
	"os"
	"regexp"
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

var iostatAwait = startIostatAwaitMeasure()

type IostatAwaitMeasure struct {
	value chan float64
	err   chan error
}

const (
	CpuUsageTimeRangeSec = 300
	SecondsInFiveMinutes = 300
	FiveMinutesInDay     = 288
)

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

func GetIostatAwait() (value float64, err error) {
	select {
	case value = <-iostatAwait.value:
	case err = <-iostatAwait.err:
	default:
	}
	return value, err
}

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
			usageSum := 0
			for _, usage := range usageAcc {
				usageSum += usage
			}
			ch.usage <- int(usageSum / CpuUsageTimeRangeSec)
			if index == CpuUsageTimeRangeSec-1 {
				index = 0
			} else {
				index++
			}
		}
	}()

	return ch
}

func startIostatAwaitMeasure() IostatAwaitMeasure {
	ch := IostatAwaitMeasure{}
	ch.value = make(chan float64)
	ch.err = make(chan error)

	go func() {
		ioawaitAccFiveMinutes := []float64{}
		ioawaitAccDay := []float64{}
		indexSeconds := 0
		indexFiveMinutes := 0

		for {
			pDiskStats, err := getDiskStats()
			if err != nil {
				ch.err <- err
				continue
			}

			pNrIos := getNrIos(pDiskStats)
			pRdTicks := getRdTicks(pDiskStats)
			pWrTicks := getWrTicks(pDiskStats)
			time.Sleep(time.Second)
			cDiskStats, err := getDiskStats()
			if err != nil {
				ch.err <- err
				continue
			}
			cNrIos := getNrIos(cDiskStats)
			cRdTicks := getRdTicks(cDiskStats)
			cWrTicks := getWrTicks(cDiskStats)

			indexSeconds++

			if cNrIos-pNrIos == 0 {
				ioawaitAccFiveMinutes = append(ioawaitAccFiveMinutes, 0.0)
			} else {
				ioawaitAccFiveMinutes = append(ioawaitAccFiveMinutes,
					float64(cRdTicks-pRdTicks+cWrTicks-pWrTicks)/
						float64(cNrIos-pNrIos))
			}

			if indexSeconds == SecondsInFiveMinutes {
				sum := 0.0
				for _, ioawait := range ioawaitAccFiveMinutes {
					sum += ioawait
				}
				avg := sum / SecondsInFiveMinutes
				indexFiveMinutes++
				if len(ioawaitAccDay) < FiveMinutesInDay {
					ioawaitAccDay = append(ioawaitAccDay, avg)
				} else {
					ioawaitAccDay[indexFiveMinutes-1] = avg
				}
				indexSeconds = 0
				ioawaitAccFiveMinutes = nil
			}

			if indexFiveMinutes == FiveMinutesInDay {
				indexFiveMinutes = 0
			}

			sum := 0.0
			for _, ioawait := range ioawaitAccDay {
				sum += ioawait
			}
			avg := 0.0
			if len(ioawaitAccDay) > 0 {
				avg = sum / float64(len(ioawaitAccDay))
			}

			ch.value <- avg
		}
	}()

	return ch
}

func getWrTicks(stats [][]string) int {
	wrTicks := 0
	for _, stat := range stats {
		statTicks, _ := strconv.Atoi(stat[10])
		wrTicks += statTicks
	}
	return wrTicks
}

func getRdTicks(deviceStats [][]string) int {
	rdTicks := 0
	for _, stat := range deviceStats {
		statTicks, _ := strconv.Atoi(stat[6])
		rdTicks += statTicks
	}
	return rdTicks
}

func getNrIos(stats [][]string) int {
	nrIos := 0
	for _, stat := range stats {
		statRdIos, _ := strconv.Atoi(stat[3])
		statWrIos, _ := strconv.Atoi(stat[7])
		nrIos += statRdIos + statWrIos
	}
	return nrIos
}

func getDiskStats() ([][]string, error) {
	result := make([][]string, 0)
	diskstats, err := ioutil.ReadFile("/proc/diskstats")
	if err != nil {
		return result, err
	}
	devicesIoStats := strings.Split(string(diskstats), "\n")
	for _, deviceStats := range devicesIoStats {
		statsArrayed := strings.Fields(string(deviceStats))
		if len(statsArrayed) == 0 {
			continue
		}
		if ok, _ := regexp.MatchString(`^sd\D+$`, statsArrayed[2]); !ok {
			continue
		}
		result = append(result, statsArrayed)
	}
	return result, nil
}
