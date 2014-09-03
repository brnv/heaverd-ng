package lxc

import (
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Container struct {
	Name string
	Host string
	Ip   string
}

var (
	reHeaver = regexp.MustCompile(`\s*([\w-]*):\s([a-z]*).*:\s([\d\.]*)/`)
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
	containers := strings.Split(string(heaverList), "\n")
	for _, container := range containers {
		parsed := reHeaver.FindStringSubmatch(container)
		if parsed != nil {
			name := parsed[1]
			list[name] = Container{
				Name: name,
				Ip:   parsed[3],
			}
		}
	}

	return list, nil
}
