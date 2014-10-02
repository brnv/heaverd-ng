package libscore

import (
	"errors"
	"fmt"
	"sort"
)

type HostsRange []Segment

func (s HostsRange) Len() int           { return len(s) }
func (s HostsRange) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s HostsRange) Less(i, j int) bool { return s[i].Hostname < s[j].Hostname }

func Segments(hosts map[string]Hostinfo) []Segment {
	segments := []Segment{}
	scoreSum := 0.0
	for _, host := range hosts {
		score := GetScore(host, DefaultProfile)
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
