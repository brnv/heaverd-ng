package libscore

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"math"
	"sort"
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

type HostsRange []Segment

func (s HostsRange) Len() int           { return len(s) }
func (s HostsRange) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s HostsRange) Less(i, j int) bool { return s[i].Hostname < s[j].Hostname }

func Segments(hosts map[string]Hostinfo) []Segment {
	segments := []Segment{}
	scoreSum := 0.0
	for _, host := range hosts {
		score := Score(host, DefaultProfile)
		segments = append(segments, Segment{
			Hostname: host.Hostname,
			Score:    score,
		})
		scoreSum += score
	}
	sort.Sort(HostsRange(segments))
	shift := 0.0
	for i := range segments {
		segments[i].X = shift
		segments[i].Y = segments[i].Score/scoreSum + shift
		shift = segments[i].Y
	}
	return segments
}

func ChooseHost(containerName string, segments []Segment) (host string, err error) {
	point := hash(containerName)
	for _, segment := range segments {
		if point >= segment.X && point <= segment.Y {
			return segment.Hostname, nil
		}
	}
	return "", errors.New(
		fmt.Sprintf("Cannot assign any host to container name %v", containerName))
}

func Score(host Hostinfo, profile Profile) float64 {
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
	realFree := ramFree - (zfsArcMax - zfsArcCurrent)
	if realFree < 0 {
		return 0
	}
	realCapacity := ramCapacity - zfsArcMax
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
