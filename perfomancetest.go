package main

import (
	"fmt"
	"heaverd-ng/libscore"
	"os"
	"time"
)

var a, b, c libscore.Host

func LocalRefreshRoutine() {
	// NOTE: MEM_MIN in kbytes
	a = libscore.Host{}
	b = libscore.Host{}
	c = libscore.Host{}

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
	Hosts := make(map[string]*libscore.Host)
	Hosts["a"] = &a
	Hosts["b"] = &b
	Hosts["c"] = &c

	for {
		time.Sleep(time.Second)
		fmt.Printf("============tick===========\n")
		//for i, host := range Hosts {
		//	fmt.Printf("%v: Host %v: length: %v\n", i, host.Hostname, host.GetLength())
		//}
		Segments := libscore.CalculateSegments(Hosts)
		for i, seg := range Segments {
			fmt.Printf("%v: segment: %+v\n", i, seg)
		}
		host, err := libscore.ChooseHost(name, Segments)
		fmt.Printf("Container %v has hash: %v\n", name, libscore.Hash(name))
		if err != nil {
			fmt.Printf("Error occured: %v\n", err)
		} else {
			fmt.Printf("Choosedhost: %v\n", host)
		}
	}

}
