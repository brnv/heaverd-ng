package lxc

import (
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Container struct {
	Name   string
	Status string
	Ip     string
}

var (
	heaverOutputPattern = regexp.MustCompile(`\s*([A-Za-z0-9_-]*):\s([a-z]*).*:\s([0-9\.]*)/`)
)

// CpuTicks возвращает метрику использования процессора контейнером,
// для пользовательского и системного времени, в "тиках"
func CpuTicks() (ticks int, err error) {
	stats, err := ioutil.ReadFile("/sys/fs/cgroup/cpu/lxc/cpuacct.stat")
	if err != nil {
		return 0, err
	}
	userTime, err := strconv.Atoi(strings.Fields(string(stats))[1])
	if err != nil {
		return 0, err
	}
	systemTime, err := strconv.Atoi(strings.Fields(string(stats))[3])
	if err != nil {
		return 0, err
	}
	ticks = userTime + systemTime
	return ticks, nil
}

func ContainerList() (map[string]Container, error) {
	cmd := exec.Command("heaver", "-L")
	heaverList, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	list := make(map[string]Container)
	chunked := strings.Split(string(heaverList), "\n")
	for _, outputChunk := range chunked {
		parsed := heaverOutputPattern.FindStringSubmatch(outputChunk)
		if parsed != nil {
			name := parsed[1]
			list[name] = Container{
				Name:   name,
				Status: parsed[2],
				Ip:     parsed[3],
			}
		}
	}

	return list, nil
}
