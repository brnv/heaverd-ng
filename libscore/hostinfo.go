package libscore

import (
	"fmt"
	"heaverd-ng/heaver"
	"heaverd-ng/libstats/linux"
	"heaverd-ng/libstats/lxc"
	"heaverd-ng/libstats/zfs"
)

type Hostinfo struct {
	Hostname      string
	CpuUsage      int
	CpuCapacity   int
	DiskFree      int
	DiskCapacity  int
	RamFree       int
	RamCapacity   int
	ZfsArcMax     int
	ControlOpTime int
	Uptime        int64
	NetAddr       []string
	Containers    map[string]lxc.Container
}

func (host *Hostinfo) Refresh() error {
	hostname, err := linux.HostName()
	if err != nil {
		return err
	}
	netAddr, err := linux.NetAddr()
	if err != nil {
		return err
	}
	zfsArcMax, err := zfs.ArcMax()
	if err != nil {
		return err
	}
	cpuCapacity, cpuUsage, err := linux.Cpu()
	if err != nil {
		return err
	}
	cpuUsage = (cpuUsage + host.CpuUsage) / 2
	diskCapacity, diskFree, err := linux.Disk()
	if err != nil {
		return err
	}
	ramCapacity, ramUsage, err := linux.Memory()
	if err != nil {
		return err
	}
	uptime, err := linux.Uptime()
	if err != nil {
		return err
	}
	// TODO: determine control operation time
	controlOpTime := 2

	containers, err := heaver.List(hostname)
	if err != nil {
		return err
	}

	*host = Hostinfo{
		Hostname:      hostname,
		CpuUsage:      cpuUsage,
		CpuCapacity:   cpuCapacity,
		DiskFree:      diskFree,
		DiskCapacity:  diskCapacity,
		RamFree:       ramUsage,
		RamCapacity:   ramCapacity,
		ZfsArcMax:     zfsArcMax,
		ControlOpTime: controlOpTime,
		Uptime:        uptime,
		NetAddr:       netAddr,
		Containers:    containers,
	}

	return nil
}

func (host *Hostinfo) String() string {
	return fmt.Sprint(*host)
}
