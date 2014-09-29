package lxc

import (
	"io/ioutil"

	"strconv"
	"strings"
)

type Container struct {
	Name   string `json:"name"`
	Host   string `json:"host"`
	Status string `json:"status"`
	Image  string `json:"image"`
	Ip     string `json:"ip"`
	Key    string `json:"key"`
}

// CpuTicks возвращает метрику использования процессора контейнером,
// для пользовательского и системного времени, в "тиках"
func GetCpuTicks() (ticks int, err error) {
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

	return userTime + systemTime, nil
}
