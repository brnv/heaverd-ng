package main

import (
	"encoding/json"
	"heaverd-ng/libscore"
	"log"
	"net"
	"os/exec"
	"time"
)

const (
	port            = ":1444"
	logDaemonPrefix = "[log] [heaverd-cluster]"
)

var cluster = make(map[string]libscore.Host)
var host = libscore.Host{}

func runClusterDaemon() {
	log.Println(logDaemonPrefix, "start listening at", port)

	notifyReadyChan := make(chan bool)
	go selfRefresh(notifyReadyChan)

	go waitForNewHosts()

	for {
		select {
		case <-notifyReadyChan:
			notifyCluster()
		}
	}
}

func waitForNewHosts() {
	connection, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(logDaemonPrefix, "[error]", err)
	}

	for {
		socket, err := connection.Accept()
		if err != nil {
			log.Fatal(logDaemonPrefix, "[error]", err)
		}
		decoder := json.NewDecoder(socket)
		host := libscore.Host{}
		err = decoder.Decode(&host)
		if err != nil {
			log.Println(logDaemonPrefix, "[error]", err)
		}
		cluster[host.Hostname] = host
		socket.Close()
	}
}

func notifyCluster() {
	json, err := json.Marshal(host)
	if err != nil {
		log.Println(logDaemonPrefix, "[error]", err)
	}
	cmd := exec.Command("./cluster-query.sh", "notify", string(json))
	err = cmd.Run()
	if err != nil {
		log.Fatal(logDaemonPrefix, "[error] [cluster-query]", err)
	}
}

func selfRefresh(notifyReadyChan chan bool) {
	for {
		err := host.Refresh()
		if err != nil {
			log.Fatal(logDaemonPrefix, "[error]", err)
		}
		notifyReadyChan <- true
		time.Sleep(5 * time.Second)
	}
}

func DaemonCluster() map[string]libscore.Host {
	return cluster
}
