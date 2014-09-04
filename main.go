package main

import (
	"flag"
	"heaverd-ng/tracker"
	"heaverd-ng/webserver"
	"log"
	"sync"
	"time"
)

var (
	webListen     string
	clusterListen string
)

func main() {
	flag.StringVar(&webListen, "web-listen", "8081", "")
	flag.StringVar(&clusterListen, "cluster-listen", "1444", "")
	flag.Parse()

	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	log.SetPrefix("[heaverd-ng] ")

	wg := sync.WaitGroup{}
	wg.Add(1)
	go webserver.Start(webListen, clusterListen, time.Now().UnixNano())
	go tracker.Start(clusterListen)
	wg.Wait()
}
