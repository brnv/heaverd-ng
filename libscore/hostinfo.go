package libscore

import (
	"fmt"
	"git.rn/go-heaver"
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
	ZfsArcCurrent int
	ControlOpTime int
	Uptime        int64
	NetAddr       []string
	CpuWeight     float64
	DiskWeight    float64
	RamWeight     float64
	Score         float64
	Pools         []string
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
	zfsArcMax, err := zfs.GetArcMax()
	if err != nil {
		return err
	}
	zfsArcCurrent, err := zfs.GetArcCurrent()
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
	ramCapacity, ramFree, err := linux.Memory()
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
		RamFree:       ramFree,
		RamCapacity:   ramCapacity,
		ZfsArcMax:     zfsArcMax,
		ZfsArcCurrent: zfsArcCurrent,
		ControlOpTime: controlOpTime,
		Uptime:        uptime,
		NetAddr:       netAddr,
		CpuWeight:     CpuWeight(cpuUsage, cpuCapacity, DefaultProfile),
		DiskWeight:    DiskWeight(diskFree, diskCapacity, DefaultProfile),
		RamWeight:     RamWeight(ramFree, ramCapacity, zfsArcMax, zfsArcCurrent, DefaultProfile),
		Containers:    containers,
		Pools:         host.Pools,
	}

	host.Score = GetScore(*host, DefaultProfile)

	return nil
}

func (host *Hostinfo) String() string {
	return fmt.Sprint(*host)
}
