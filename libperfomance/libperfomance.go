package libperfomance

import (
	"crypto/sha1"
	"heaverd-ng/res/linux"
	"heaverd-ng/res/linux/zfs"
	"math"
)

// Structure of resources of host machine
type Resources struct {
	// current cpu usage in % by containers in host
	Cpu int
	// cpu capacity (# of cores * 100%)
	CpuCapacity int
	// current disk free space (containers working partition)
	Disk int
	// total disk capacity
	DiskCapacity int
	// current ram free
	Ram int
	// total ram capacity
	RamCapacity int

	// current zfs_arc_max value
	ZfsArcMax int

	// control operation time
	ControlOpTime int
	// uptime value in seconds
	Uptime int
}

// resource reservation counstants
type Reserved struct {
	// system reservation of CPU (defaul 100 (%) = 1 core for example)
	CPU_MIN int
	// system reservation of disk
	DISK_MIN float32
	// system reservation of memory
	MEM_MIN int
	// time in seconds after which the host is considered to be slow
	OP_TIME_THRESHOLD int
	// time of operation of host elimination from balancing
	SLOWNESS int
	// time of operation of host introduction in balancing
	UPTIME_PERIOD int
}

// weight & factors
type Factors struct {
	CpuWeight    float64
	DiskWeight   float64
	RamWeight    float64
	SpeedFactor  float64
	UptimeFactor float64
}

// define single host configuration structure
type Host struct {
	// Hostname string
	Hostname  string
	Resources Resources
	Reserved  Reserved
	Factors   Factors
	length    float64
}

// segment, where host-server lies
type Segment struct {
	X, Y float64
}

// Generate value that belongs segment [0;1]
func Hash(input string) float64 {
	data := []byte(input)
	hashsum := sha1.Sum(data)
	var ret uint32 = 0
	// length of hashsum is always eq 20
	for i := 0; i < 20; i++ {
		ret = ret + uint32(hashsum[19-i])*uint32(math.Pow(16, float64(i)))
	}
	ret = ret % 1000
	return float64(ret) / 1000.0
}

// normalization by minimal argument
func MinNorm(a, b int) float64 {
	var min int
	if a < b {
		min = a
	} else {
		min = b
	}
	return float64(min) / float64(b)
}

// calculate host segments
// TODO: make sorting
func CalculateSegments(input map[string]*Host) map[string]*Segment {
	Segments := make(map[string]*Segment)
	sum := 0.0
	shift := 0.0
	// get all legnths and summary the segment
	for name, host := range input {
		Segments[name] = &Segment{X: 0.0, Y: host.GetLength()}
		sum += Segments[name].Y
	}

	for name, seg := range Segments {
		// let the left point of segment be the right point of previous segment
		// if it's first segment, shift will be 0
		Segments[name] = &Segment{X: shift, Y: seg.Y/sum + shift}
		shift = Segments[name].Y
	}
	return Segments
}

// refresh method takes 1sec to complete operation, for determining current cpu usage
func (host *Host) Refresh() (err error) {
	host.Hostname, err = linux.GetHostName()
	if err != nil {
		return err
	}
	host.Resources.ZfsArcMax, err = zfs.GetArcMax()
	if err != nil {
		return err
	}

	CpuCapacity, CpuUsage, err := linux.GetCpuStats()
	if err != nil {
		return err
	}
	DiskCapacity, DiskUsage, err := linux.GetDiskStats()
	if err != nil {
		return err
	}
	RamCapacity, RamUsage, err := linux.GetRamStats()
	if err != nil {
		return err
	}

	host.Resources.Cpu = (CpuUsage + host.Resources.Cpu) / 2
	host.Resources.CpuCapacity = CpuCapacity
	host.Factors.CpuWeight = 1.0 - MinNorm(host.Resources.Cpu, host.Resources.CpuCapacity-host.Reserved.CPU_MIN)

	host.Resources.Disk = DiskUsage
	host.Resources.DiskCapacity = DiskCapacity
	host.Factors.DiskWeight = 1.0 - MinNorm(int(float32(host.Resources.DiskCapacity)*host.Reserved.DISK_MIN), host.Resources.Disk)

	host.Resources.Ram = RamUsage
	host.Resources.RamCapacity = RamCapacity
	host.Factors.RamWeight = 1 - MinNorm(host.Resources.Ram, host.Resources.RamCapacity-host.Resources.ZfsArcMax-host.Reserved.MEM_MIN)

	host.Resources.Uptime, err = linux.GetUptime()
	if err != nil {
		return err
	}

	// TODO: determine control operation time
	host.Resources.ControlOpTime = 2

	host.Factors.UptimeFactor = 2 * math.Atan(float64(host.Resources.Uptime)/float64(host.Reserved.UPTIME_PERIOD)) / math.Pi

	host.Factors.SpeedFactor = 1 - 2*math.Atan(math.Max(0, float64(host.Resources.ControlOpTime-host.Reserved.OP_TIME_THRESHOLD))/float64(host.Reserved.SLOWNESS))

	host.length = host.Factors.CpuWeight * host.Factors.DiskWeight * host.Factors.RamWeight * host.Factors.SpeedFactor * host.Factors.UptimeFactor

	return
}

func (host *Host) GetLength() float64 {
	return host.length
}
