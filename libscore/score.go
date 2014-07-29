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

var defaultProfile = Profile{
	ReservedCPU:          100,
	ReservedDiskCapacity: 0.1,
	ReservedRAM:          1048576,
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

func Segments(hosts []Host) []Segment {
	Segments := []Segment{}
	scoreSum := 0.0
	for _, host := range hosts {
		score := calculate(host, defaultProfile)
		Segments = append(Segments, Segment{Hostname: host.Hostname, Score: score})
		scoreSum += score
	}
	sort.Sort(HostsRange(Segments))
	shift := 0.0
	for i := range Segments {
		Segments[i].X = shift
		Segments[i].Y = Segments[i].Score/scoreSum + shift
		shift = Segments[i].Y
	}
	return Segments
}

func ChooseHost(container string, segments []Segment) (host string, err error) {
	cval := hash(container)
	for _, segment := range segments {
		if cval >= segment.X && cval <= segment.Y {
			return segment.Hostname, nil
		}
	}
	return "", errors.New(
		fmt.Sprintf("Cannot assign any host to container name %v", container))
}

func calculate(host Host, profile Profile) float64 {
	cpuWeight := 1.0 - minNorm(host.CpuUsage, host.CpuCapacity-profile.ReservedCPU)
	diskWeight := 1.0 - minNorm(int(float32(host.DiskCapacity)*profile.ReservedDiskCapacity), host.DiskUsage)
	ramWeight := 1 - minNorm(host.RamUsage, host.RamCapacity-host.ZfsArcMax-profile.ReservedRAM)
	uptimeFactor := 2 * math.Atan(float64(host.Uptime)/float64(profile.UptimeFactor)) / math.Pi
	speedFactor := 1 - 2*math.Atan(math.Max(0, float64(host.ControlOpTime-
		profile.SlowOpThreshold))/float64(profile.LagReactionSpeed))

	score := cpuWeight * diskWeight * ramWeight * speedFactor * uptimeFactor

	return score
}

func minNorm(a, b int) float64 {
	var min int
	if a < b {
		min = a
	} else {
		min = b
	}
	return float64(min) / float64(b)
}

func hash(input string) float64 {
	hashsum := sha1.Sum([]byte(input))
	ret := uint32(hashsum[0]) + uint32(hashsum[1])*(1<<8)
	ret = ret % 1000
	return float64(ret) / 1000.0
}
