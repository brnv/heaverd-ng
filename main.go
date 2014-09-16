package main

import (
	"flag"
	"heaverd-ng/tracker"
	"heaverd-ng/webapi"
	"log"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

var (
	webApiConfig  = "config/webapi.toml"
	trackerConfig = "config/tracker.toml"
)

func main() {
	if _, err := toml.DecodeFile(webApiConfig, &webapi.Conf); err != nil {
		log.Fatal("[error]", err)
	}
	if _, err := toml.DecodeFile(trackerConfig, &tracker.Conf); err != nil {
		log.Fatal("[error]", err)
	}

	log.Println("[error]", tracker.Conf)

	flag.StringVar(&webapi.Conf.Web.Port, "web-addr",
		webapi.Conf.Web.Port, "")
	flag.StringVar(&tracker.Conf.Cluster.Port, "peer-addr",
		tracker.Conf.Cluster.Port, "")
	flag.StringVar(&tracker.Conf.Storage.Port, "etcd-addr",
		tracker.Conf.Storage.Port, "")
	flag.Parse()

	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	log.SetPrefix("[heaverd-ng] ")

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go webapi.Start(wg, time.Now().UnixNano())
	go tracker.Start(wg)
	wg.Wait()
}
