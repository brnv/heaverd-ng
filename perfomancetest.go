package main

import (
	"fmt"
	"heaverd-ng/libperfomance"
	"os"
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
		go func() {
			err := a.Refresh()
			if err != nil {
				fmt.Printf("ERROR OCCURED: %v\n", err)
				os.Exit(1)
			}
		}()
		go func() {
			err := b.Refresh()
			if err != nil {
				fmt.Printf("ERROR OCCURED: %v\n", err)
				os.Exit(1)
			}
		}()
		go func() {
			err := c.Refresh()
			if err != nil {
				fmt.Printf("ERROR OCCURED: %v\n", err)
				os.Exit(1)
			}
		}()
		time.Sleep(time.Second)
	}
}

func main() {
	go LocalRefreshRoutine()
	name := os.Args[1]
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
		host, err := libperfomance.ChooseHost(name, Segments)
		fmt.Printf("Container %v has hash: %v\n", name, libperfomance.Hash(name))
		if err != nil {
			fmt.Printf("Error occured: %v\n", err)
		} else {
			fmt.Printf("Choosedhost: %v\n", host)
		}
	}

}
