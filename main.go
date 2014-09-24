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
	webApiConfigPath  = "/etc/heaverd-ng/webapi.toml"
	webPort           = "8081"
	webTemplates      = "/mnt/a.baranov/www/templates/"
	trackerConfigPath = "/etc/heaverd-ng/tracker.toml"
	clusterPort       = "1444"
	etcdPort          = "4001"
)

func main() {
	flag.StringVar(&webApiConfigPath, "webapi-config", webApiConfigPath, "")
	flag.StringVar(&trackerConfigPath, "tracker-config", trackerConfigPath, "")
	flag.StringVar(&webPort, "web-port", webPort, "")
	flag.StringVar(&clusterPort, "cluster-port", clusterPort, "")
	flag.StringVar(&etcdPort, "etcd-port", etcdPort, "")
	flag.Parse()

	webapi.Config.Set(webPort, "web", "port")
	webapi.Config.Set(webTemplates, "templates", "dir")
	tracker.Config.Set(clusterPort, "cluster", "port")
	tracker.Config.Set(etcdPort, "etcd", "port")

	f, err := os.Open(webApiConfigPath)
	if err == nil {
		webapi.Config.ReadHash(bufio.NewReader(f))
	} else {
		log.Println("[error]", "can't read configuration file ", webApiConfigPath)
	}

	f, err = os.Open(trackerConfigPath)
	if err == nil {
		tracker.Config.ReadHash(bufio.NewReader(f))
	} else {
		log.Println("[error]", "can't read configuration file ", trackerConfigPath)
	}

	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	log.SetPrefix("[heaverd-ng] ")

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go webapi.Start(wg, time.Now().UnixNano())
	go tracker.Start(wg)
	wg.Wait()
}
