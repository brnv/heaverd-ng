package main

import (
	"fmt"
	"heaverd-ng/libperfomance"
	"time"
)

var a, b, c libperfomance.Host

func LocalRefreshRoutine() {
	// NOTE: MEM_MIN in kbytes
	a = libperfomance.Host{Reserved: libperfomance.Reserved{CPU_MIN: 100, DISK_MIN: 0.1, MEM_MIN: 1048576, OP_TIME_THRESHOLD: 300, SLOWNESS: 120, UPTIME_PERIOD: 130}}
	b = libperfomance.Host{Reserved: libperfomance.Reserved{CPU_MIN: 100, DISK_MIN: 0.1, MEM_MIN: 1048576, OP_TIME_THRESHOLD: 300, SLOWNESS: 120, UPTIME_PERIOD: 130}}
	c = libperfomance.Host{Reserved: libperfomance.Reserved{CPU_MIN: 100, DISK_MIN: 0.1, MEM_MIN: 1048576, OP_TIME_THRESHOLD: 300, SLOWNESS: 120, UPTIME_PERIOD: 130}}

	// testing
	for {
		go a.Refresh()
		go b.Refresh()
		go c.Refresh()
		time.Sleep(time.Second)
	}
}

func main() {
	go LocalRefreshRoutine()

	Hosts := make(map[string]*libperfomance.Host)
	Hosts["a"] = &a
	Hosts["b"] = &b
	Hosts["c"] = &c

	for {
		time.Sleep(time.Second)
		fmt.Printf("============tick===========\n")
		//for i, host := range Hosts {
		//	fmt.Printf("%v: Host %v: length: %v\n", i, host.Hostname, host.GetLength())
		//}
		Segments := libperfomance.CalculateSegments(Hosts)
		for i, seg := range Segments {
			fmt.Printf("%v: segment: %+v\n", i, seg)
		}
		fmt.Printf("\n")
	}

}
