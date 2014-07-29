package main

import (
	"fmt"
	"heaverd-ng/libscore"
	"os"
	"time"
)

func main() {
	Hosts := make([]libscore.Host, 3)
	HostA := libscore.Host{}
	HostB := libscore.Host{}
	HostC := libscore.Host{}

	go func() {
		for {
			HostA.Refresh()
			HostB.Refresh()
			HostC.Refresh()
			Hosts[0] = HostA
			Hosts[1] = HostB
			Hosts[2] = HostC
			time.Sleep(time.Second)
		}
	}()

	name := os.Args[1]
	for {
		time.Sleep(time.Second)
		fmt.Printf("============tick===========\n")

		Segments := libscore.Segments(Hosts)

		for i, seg := range Segments {
			fmt.Printf("%v: segment: %+v\n", i, seg)
		}
		host, err := libscore.ChooseHost(name, Segments)

		if err != nil {
			fmt.Printf("Error occured: %v\n", err)
		} else {
			fmt.Printf("Choosedhost: %v\n", host)
		}
	}
}
