package libscore

import (
	"crypto/sha1"
	"math"
)

type Profile struct {
	ReservedCPU          int
	ReservedDiskCapacity float32
	ReservedRAM          int
	SlowOpThreshold      int
	LagReactionSpeed     int
	UptimeFactor         int
}

var DefaultProfile = Profile{
	ReservedCPU:          100,
	ReservedDiskCapacity: 0.1,
	ReservedRAM:          1 * 1024 * 1024,
	SlowOpThreshold:      300,
	LagReactionSpeed:     120,
	UptimeFactor:         130,
}

type Segment struct {
	Hostname string
	Score    float64
	X, Y     float64
}

func GetScore(host Hostinfo, profile Profile) float64 {
	uptimeFactor := 2 * math.Atan(float64(host.Uptime)/
		float64(profile.UptimeFactor)) / math.Pi

	speedFactor := 1 - 2*math.Atan(math.Max(0, float64(host.ControlOpTime-
		profile.SlowOpThreshold))/float64(profile.LagReactionSpeed))

	score := host.CpuWeight * host.DiskWeight * host.RamWeight * speedFactor * uptimeFactor

	return score
}

func CpuWeight(cpuUsage int, cpuCapacity int, profile Profile) float64 {
	return 1.0 - minNorm(cpuUsage, cpuCapacity-profile.ReservedCPU)
}

func DiskWeight(diskFree int, diskCapacity int, profile Profile) float64 {
	return minNorm(diskFree,
		int(float32(diskCapacity)*(1-profile.ReservedDiskCapacity)))
}

func RamWeight(ramFree int, ramCapacity int,
	zfsArcMax int, zfsArcCurrent int, profile Profile) float64 {
	// 22.10.2014 remove zfs arc from calculation
	//realFree := ramFree - (zfsArcMax - zfsArcCurrent)
	realFree := ramFree
	if realFree < 0 {
		return 0
	}
	// 22.10.2014 remove zfs arc from calculation
	//realCapacity := ramCapacity - zfsArcMax
	realCapacity := ramCapacity
	return minNorm(realFree, realCapacity-profile.ReservedRAM)
}

func minNorm(a, b int) float64 {
	if a < b {
		return float64(a) / float64(b)
	} else {
		return 1.0
	}
}

func hash(input string) float64 {
	hashsum := sha1.Sum([]byte(input))
	ret := uint32(hashsum[0]) + uint32(hashsum[1])*(1<<8)
	ret = ret % 1000
	return float64(ret) / 1000.0
}
