package main

import (
	"bufio"
	"flag"
	"heaverd-ng/tracker"
	"heaverd-ng/webapi"
	"log"
	"os"
	"sync"
	"time"
)

var (
	webApiConfigPath  = "config/webapi.toml"
	trackerConfigPath = "config/tracker.toml"
	webAddr           string
	peerAddr          string
	etcdAddr          string
)

func main() {
	f, err := os.Open(webApiConfigPath)
	if err != nil {
		log.Fatal("[error]", err)
	}
	webapi.Config.ReadHash(bufio.NewReader(f))

	f, err = os.Open(trackerConfigPath)
	if err != nil {
		log.Fatal("[error]", err)
	}
	tracker.Config.ReadHash(bufio.NewReader(f))

	flag.StringVar(&webAddr, "web-addr", "", "")
	flag.StringVar(&peerAddr, "peer-addr", "", "")
	flag.StringVar(&etcdAddr, "etcd-addr", "", "")
	flag.Parse()

	if webAddr != "" {
		webapi.Config.Set(webAddr, "web", "port")
	}

	if peerAddr != "" {
		tracker.Config.Set(peerAddr, "cluster", "port")
	}

	if etcdAddr != "" {
		tracker.Config.Set(etcdAddr, "etcd", "port")
	}

	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	log.SetPrefix("[heaverd-ng] ")

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go webapi.Start(wg, time.Now().UnixNano())
	go tracker.Start(wg)
	wg.Wait()
}
