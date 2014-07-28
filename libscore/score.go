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
	ReservedRAM:          0, //1048576,
	SlowOpThreshold:      300,
	LagReactionSpeed:     120,
	UptimeFactor:         130,
}

type Segment struct {
	X, Y float64
}

func Hash(input string) float64 {
	hashsum := sha1.Sum([]byte(input))
	ret := uint32(hashsum[0]) + uint32(hashsum[1])*(1<<8)
	ret = ret % 1000
	return float64(ret) / 1000.0
}

// calculate host segments
func CalculateSegments(input map[string]*Host) map[string]*Segment {
	slice := make([]string, len(input))
	Segments := make(map[string]*Segment)
	sum := 0.0
	shift := 0.0
	count := 0
	// get all legnths and summary the segment
	for name, host := range input {
		Segments[name] = &Segment{X: 0.0, Y: calculate(host, defaultProfile)}
		sum += Segments[name].Y
		slice[count] = name
		count += 1
	}

	sort.Strings(slice)

	for i := range slice {
		// let the left point of segment be the right point of previous segment
		// if it's first segment, shift will be 0
		Segments[slice[i]].X = shift
		Segments[slice[i]].Y = Segments[slice[i]].Y/sum + shift
		shift = Segments[slice[i]].Y
	}
	return Segments
}

func ChooseHost(container string, fragmentation map[string]*Segment) (host string, err error) {
	// get float value which belongs to [0;1]
	cval := Hash(container)
	for name, segment := range fragmentation {
		if cval >= segment.X && cval <= segment.Y {
			return name, nil
		}
	}
	return "", errors.New(
		fmt.Sprintf("Cannot assign any host to container name %v", container))

}

func calculate(host *Host, profile Profile) float64 {
	cpuWeight := 1.0 - minNorm(host.CpuUsage, host.CpuCapacity-profile.ReservedCPU)
	diskWeight := 1.0 - minNorm(int(float32(host.DiskCapacity)*profile.ReservedDiskCapacity), host.DiskUsage)
	ramWeight := 1 - minNorm(host.RamUsage, host.RamCapacity-host.ZfsArcMax-profile.ReservedRAM)
	uptimeFactor := 2 * math.Atan(float64(host.Uptime)/float64(profile.UptimeFactor)) / math.Pi

	speedFactor := 1 - 2*math.Atan(math.Max(0, float64(host.ControlOpTime-profile.SlowOpThreshold))/float64(profile.LagReactionSpeed))

	score := cpuWeight * diskWeight * ramWeight * speedFactor * uptimeFactor

	return score
}

// normalization by minimal argument
func minNorm(a, b int) float64 {
	var min int
	if a < b {
		min = a
	} else {
		min = b
	}
	return float64(min) / float64(b)
}
