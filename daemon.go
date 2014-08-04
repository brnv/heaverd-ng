package main

import (
	"encoding/json"
	"heaverd-ng/libscore"
	"log"
	"net"
	"os/exec"
	"time"
)

var cluster *map[string]libscore.Host

const (
	port            = ":1444"
	logDaemonPrefix = "[log] [heaverd] [cluster] [daemon]"
)

func runClusterDaemon() {
	log.Println(logDaemonPrefix, "start listening at", port)

	host, err := initHost()
	if err != nil {
		log.Fatal(logDaemonPrefix, "[error]", err)
	}

	go selfRefreshing(host)

	cluster := make(map[string]libscore.Host)

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(logDaemonPrefix, "[error]", err)
	}

	go clusterListening(listener)

	for {
		notifyCluster(host)
	}
}

func initHost() (*libscore.Host, error) {
	host := libscore.Host{}
	err := host.Refresh()
	if err != nil {
		return nil, err
	}
	return &host, nil
}

func clusterListening(conn net.Listener) {
	for {
		socket, err := conn.Accept()
		if err != nil {
			log.Fatal(logDaemonPrefix, "[error]", err)
		}
		decoder := json.NewDecoder(socket)
		host := libscore.Host{}
		err = decoder.Decode(&host)
		cl := *cluster
		_, ok := cl[host.Hostname]
		if ok != true {
			log.Println(logDaemonPrefix, "new host registered:", host.Hostname)
		}
		cl[host.Hostname] = host
		socket.Close()
	}
}

func notifyCluster(host *libscore.Host) {
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

func selfRefreshing(host *libscore.Host) {
	for {
		err := host.Refresh()
		if err != nil {
			log.Fatal(logDaemonPrefix, "[error]", err)
		}
		time.Sleep(time.Second)
	}
}

func GetCluster() map[string]libscore.Host {
	return *cluster
}
