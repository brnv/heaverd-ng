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

func Segments(hosts map[string]Info) []Segment {
	segments := []Segment{}
	scoreSum := 0.0
	for _, host := range hosts {
		score := calculateHostScore(host, DefaultProfile)
		segments = append(segments, Segment{Hostname: host.Hostname, Score: score})
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

func calculateHostScore(host Info, profile Profile) float64 {
	cpuWeight := 1.0 - minNorm(host.CpuUsage, host.CpuCapacity-profile.ReservedCPU)
	diskWeight := 1.0 - minNorm(int(float32(host.DiskCapacity)*profile.ReservedDiskCapacity), host.DiskUsage)
	ramWeight := 1 - minNorm(host.RamUsage, host.RamCapacity-(host.ZfsArcMax/1024)-profile.ReservedRAM)
	uptimeFactor := 2 * math.Atan(float64(host.Uptime)/float64(profile.UptimeFactor)) / math.Pi
	speedFactor := 1 - 2*math.Atan(math.Max(0, float64(host.ControlOpTime-
		profile.SlowOpThreshold))/float64(profile.LagReactionSpeed))

	// TODO	пока на хосте lxbox и yapa исключил этот параметр
	ramWeight = 1
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
