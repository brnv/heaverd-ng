package libscore

import (
	"fmt"
	"heaverd-ng/libstats/linux"
	"heaverd-ng/libstats/lxc"
	"heaverd-ng/libstats/zfs"
)

type Hostinfo struct {
	Hostname      string
	CpuUsage      int
	CpuCapacity   int
	DiskUsage     int
	DiskCapacity  int
	RamUsage      int
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
	diskCapacity, diskUsage, err := linux.Disk()
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

	containers, err := lxc.ContainerList()
	if err != nil {
		return err
	}

	*host = Hostinfo{
		Hostname:      hostname,
		CpuUsage:      cpuUsage,
		CpuCapacity:   cpuCapacity,
		DiskUsage:     diskUsage,
		DiskCapacity:  diskCapacity,
		RamUsage:      ramUsage,
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
