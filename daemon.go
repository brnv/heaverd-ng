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

var (
	cluster   map[string]libscore.Host
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

	cluster = make(map[string]libscore.Host)
	cluster = cluster

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(logPrefix, logPostfix, "[error]", err)
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
		host := libscore.Host{}
		err = decoder.Decode(&host)
		if err != nil {
			log.Println(logPrefix, logPostfix, "[error]", err)
			continue
		}
		_, ok := cluster[host.Hostname]
		if ok != true {
			log.Println(logPrefix, logPostfix, "new host registered:", host.Hostname)
		}
		cluster[host.Hostname] = host
		socket.Close()
	}
}

func notifyCluster(host *libscore.Host) {
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

func selfRefreshing(host *libscore.Host) {
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

func GetCluster() map[string]libscore.Host {
	return cluster
}
