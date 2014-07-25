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
	Uptime        int
}

// refresh method takes 1sec to complete operation, for determining current cpu usage
func (host *Host) Refresh() (err error) {
	host.Hostname, err = linux.HostName()
	if err != nil {
		return err
	}
	host.ZfsArcMax, err = zfs.ArcMax()
	if err != nil {
		return err
	}
	CpuCapacity, CpuUsage, err := linux.Cpu()
	if err != nil {
		return err
	}
	host.CpuUsage = (CpuUsage + host.CpuUsage) / 2
	host.CpuCapacity = CpuCapacity
	host.DiskCapacity, host.DiskUsage, err = linux.Disk()
	if err != nil {
		return err
	}
	host.RamCapacity, host.RamUsage, err = linux.Memory()
	if err != nil {
		return err
	}
	host.Uptime, err = linux.Uptime()
	if err != nil {
		return err
	}
	// TODO: determine control operation time
	host.ControlOpTime = 2
	host.NetAddr, err = linux.NetAddr()
	if err != nil {
		return err
	}

	return err
}
