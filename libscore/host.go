package libscore

import (
	"heaverd-ng/libstats/linux"
	"heaverd-ng/libstats/zfs"
)

type Host struct {
	Hostname      string
	NetAddr       []string
	CpuUsage      int
	CpuCapacity   int
	DiskUsage     int
	DiskCapacity  int
	RamUsage      int
	RamCapacity   int
	ZfsArcMax     int
	ControlOpTime int
	Uptime        int64
}

func (host *Host) Refresh() error {
	hostname, err := linux.HostName()
	if err != nil {
		return err
	}
	netAddr, err = linux.NetAddr()
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

	*host = Host{
		Hostname:      hostname,
		NetAddr:       netAddr,
		CpuUsage:      cpuUsage,
		CpuCapacity:   cpuCapacity,
		DiskUsage:     diskUsage,
		DiskCapacity:  diskCapacity,
		RamUsage:      ramUsage,
		RamCapacity:   ramCapacity,
		ZfsArcMax:     zfsArcMax,
		ControlOpTime: controlOpTime,
		Uptime:        uptime,
	}

	return nil
}
