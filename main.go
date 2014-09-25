package main

import (
	"heaverd-ng/tracker"
	"heaverd-ng/webapi"
	"log"
	"sync"
	"time"
)

var (
	configPath = "config/config.toml"
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	log.SetPrefix("[heaverd-ng] ")

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() { webapi.Run(configPath, time.Now().UnixNano()); wg.Done() }()
	go func() { tracker.Run(configPath); wg.Done() }()
	wg.Wait()
}
