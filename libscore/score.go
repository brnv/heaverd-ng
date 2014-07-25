package libscore

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"math"
	"sort"
)

const (
	cpuIdle         = 100
	diskIdle        = 0.1
	memFree         = 1048576
	opTimeThreshold = 300
	slowness        = 120
	uptime          = 130
)

type Segment struct {
	X, Y float64
}

func Hash(input string) float64 {
	hashsum := sha1.Sum([]byte(input))
	var ret uint32 = 0
	// length of hashsum is always eq 20
	for i := 0; i < 20; i++ {
		ret = ret + uint32(hashsum[19-i])*uint32(math.Pow(16, float64(i)))
	}
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
		Segments[name] = &Segment{X: 0.0, Y: calculate(host)}
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

func calculate(host *Host) float64 {
	cpuWeight := 1.0 - minNorm(host.CpuUsage, host.CpuCapacity-cpuIdle)
	diskWeight := 1.0 - minNorm(int(float32(host.DiskCapacity)*diskIdle), host.DiskUsage)
	ramWeight := 1 - minNorm(host.RamUsage, host.RamCapacity-host.ZfsArcMax-memFree)
	uptimeFactor := 2 * math.Atan(float64(host.Uptime)/float64(uptime)) / math.Pi

	speedFactor := 1 - 2*math.Atan(math.Max(0, float64(host.ControlOpTime-opTimeThreshold))/float64(slowness))

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
