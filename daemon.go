package main

import (
	"encoding/json"
	"fmt"
	"heaverd-ng/libscore"
	"log"
	"net"
	"os/exec"
	"time"
)

type Host struct {
	info     libscore.Info
	lastSeen int64
	stale    bool
}

var (
	cluster   = make(map[string]*Host)
	logPrefix = "[log] [daemon.go]"
)

const (
	port = ":1444"
)

func runClusterDaemon() {
	var (
		logPostfix = "[runClusterDaemon]"
	)

	log.Println(logPrefix, logPostfix, "start listening at", port)

	host, err := initHost()
	if err != nil {
		log.Fatal(logPrefix, logPostfix, "[error]", err)
	}

	go selfRefreshing(host)

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(logPrefix, logPostfix, "[error]", err)
	}

	hostsChan := make(chan libscore.Info)
	go clusterListening(listener, hostsChan)
	go clusterRefreshing(hostsChan)

	for {
		notifyCluster(host)
	}
}

func initHost() (*libscore.Info, error) {
	host := libscore.Info{}
	err := host.Refresh()
	if err != nil {
		return nil, err
	}
	return &host, nil
}

func clusterRefreshing(hostsChan chan libscore.Info) {
	var (
		logPostfix = "[clusterRefreshing]"
	)

	for {
		select {
		case host := <-hostsChan:
			if _, ok := cluster[host.Hostname]; !ok {
				log.Println(logPrefix, logPostfix, "new host registered:", host.Hostname)
				cluster[host.Hostname] = &Host{}
			}
			cluster[host.Hostname].info = host
			cluster[host.Hostname].lastSeen = time.Now().Unix()
			cluster[host.Hostname].stale = false
		default:
			for name, host := range cluster {
				if host.stale == true {
					log.Println(logPrefix, logPostfix, "host is droped:", name)
					delete(cluster, name)
					continue
				}
				if time.Now().Unix()-host.lastSeen > 5 {
					log.Println(logPrefix, logPostfix, "host is stale:", name)
					cluster[name].stale = true
				}
			}
		}
		time.Sleep(time.Second)
	}
}

func clusterListening(conn net.Listener, hostsChan chan libscore.Info) {
	var (
		logPostfix = "[clusterListening]"
	)

	for {
		socket, err := conn.Accept()
		if err != nil {
			log.Println(logPrefix, logPostfix, "[error]", err)
			continue
		}
		decoder := json.NewDecoder(socket)
		host := libscore.Info{}
		err = decoder.Decode(&host)
		if err != nil {
			log.Println(logPrefix, logPostfix, "[error]", err)
			continue
		}

		hostsChan <- host
		socket.Close()
	}
}

func notifyCluster(host *libscore.Info) {
	var (
		logPostfix = "[notifyCluster]"
	)

	json, err := json.Marshal(host)
	if err != nil {
		log.Println(logPrefix, logPostfix, "[error]", err)
	}
	cmd := exec.Command("./cluster-query", "notify", fmt.Sprintf("%s", json))
	err = cmd.Run()
	if err != nil {
		log.Println(logPrefix, logPostfix, "[error]", err)
	}
}

func selfRefreshing(host *libscore.Info) {
	var (
		logPostfix = "[selfRefreshing]"
	)

	for {
		err := host.Refresh()
		if err != nil {
			log.Println(logPrefix, logPostfix, "[error]", err)
		}
		time.Sleep(time.Second)
	}
}

func GetCluster() map[string]libscore.Info {
	result := make(map[string]libscore.Info)
	for name, host := range cluster {
		result[name] = host.info
	}
	return result
}
