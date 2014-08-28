package main

import (
	"flag"
	"heaverd-ng/tracker"
	"heaverd-ng/webserver"
	"log"
	"sync"
	"time"
)

func main() {
	flag.Parse()

	var webListen string
	flag.StringVar(&webListen, "web-listen", "8081", "")

	var clusterListen string
	flag.StringVar(&clusterListen, "cluster-listen", "1444", "")

	log.SetFlags(log.Lshortfile)
	log.SetPrefix("[heaverd-ng] ")

	wg := sync.WaitGroup{}
	wg.Add(1)
	go webserver.Start(":"+webListen, time.Now().UnixNano())
	go tracker.Start(":" + clusterListen)
	wg.Wait()
}
