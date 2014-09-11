package main

import (
	"flag"
	"heaverd-ng/tracker"
	"heaverd-ng/webapi"
	"log"
	"sync"
	"time"
)

var (
	webAddr  string
	peerAddr string
	etcdAddr string
)

func main() {
	flag.StringVar(&webAddr, "web-addr", "8081", "")
	flag.StringVar(&peerAddr, "peer-addr", "1444", "")
	flag.StringVar(&etcdAddr, "etcd-addr", "4001", "")
	flag.Parse()

	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	log.SetPrefix("[heaverd-ng] ")

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go webapi.Start(wg, webAddr, peerAddr, time.Now().UnixNano())
	go tracker.Start(wg, peerAddr, etcdAddr)
	wg.Wait()
}
