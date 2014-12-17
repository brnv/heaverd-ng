package libscore

import (
	"fmt"
	"time"

	"github.com/brnv/go-heaver"
	"github.com/brnv/go-lxc"
	"github.com/brnv/heaverd-ng/libstats/linux"
	"github.com/brnv/heaverd-ng/libstats/zfs"
)

type Hostinfo struct {
	LastUpdateTimestamp int64
	Hostname            string
	CpuUsage            int
	CpuCapacity         int
	DiskFree            int
	DiskCapacity        int
	RamFree             int
	RamCapacity         int
	ZfsArcMax           int
	ZfsArcCurrent       int
	ControlOpTime       int
	IostatAwait         float64
	Uptime              int64
	NetAddr             []string
	CpuWeight           float64
	DiskWeight          float64
	RamWeight           float64
	DiskIOWeight        float64
	Score               float64
	Pools               []string
	Containers          map[string]lxc.Container
	Images              map[string]heaver.Image
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
	cpuCapacity, cpuUsage, err := linux.GetCpu()
	if err != nil {
		return err
	}
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

	containers, err := heaver.ListContainers(hostname)
	if err != nil {
		return err
	}

	images, err := heaver.ListImages()
	if err != nil {
		//do nothing
	}

	iostatAwait, err := linux.GetIostatAwait()
	if err != nil {
		return err
	}

	*host = Hostinfo{
		LastUpdateTimestamp: time.Now().UnixNano(),
		Hostname:            hostname,
		CpuUsage:            cpuUsage,
		CpuCapacity:         cpuCapacity,
		DiskFree:            diskFree,
		DiskCapacity:        diskCapacity,
		RamFree:             ramFree,
		RamCapacity:         ramCapacity,
		ZfsArcMax:           zfsArcMax,
		ZfsArcCurrent:       zfsArcCurrent,
		ControlOpTime:       controlOpTime,
		IostatAwait:         iostatAwait,
		Uptime:              uptime,
		NetAddr:             netAddr,
		CpuWeight:           CpuWeight(cpuUsage, cpuCapacity, DefaultProfile),
		DiskWeight:          DiskWeight(diskFree, diskCapacity, DefaultProfile),
		RamWeight:           RamWeight(ramFree, ramCapacity, zfsArcMax, zfsArcCurrent, DefaultProfile),
		DiskIOWeight:        DiskIOWeight(iostatAwait, DefaultProfile),
		Containers:          containers,
		Pools:               host.Pools,
		Images:              images,
	}

	host.Score = GetScore(*host, DefaultProfile)

	return nil
}

func (host *Hostinfo) String() string {
	return fmt.Sprint(*host)
}
