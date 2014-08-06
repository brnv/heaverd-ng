package tracker

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
	cluster = make(map[string]*Host)
)

func Start(port string) {
	log.Println("started at port", port)

	host, err := initHost()
	if err != nil {
		log.Fatal("[error]", err)
	}

	go selfRefreshing(host)

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("[error]", err)
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
	for {
		select {
		case host := <-hostsChan:
			if _, ok := cluster[host.Hostname]; !ok {
				log.Println("new host:", host.Hostname)
				cluster[host.Hostname] = &Host{}
			}
			cluster[host.Hostname].info = host
			cluster[host.Hostname].lastSeen = time.Now().Unix()
			cluster[host.Hostname].stale = false
		default:
			for name, host := range cluster {
				if host.stale == true {
					log.Println("host is droped:", name)
					delete(cluster, name)
					continue
				}
				if time.Now().Unix()-host.lastSeen > 5 {
					log.Println("host is stale:", name)
					cluster[name].stale = true
				}
			}
		}
		time.Sleep(time.Second)
	}
}

func clusterListening(conn net.Listener, hostsChan chan libscore.Info) {
	for {
		socket, err := conn.Accept()
		if err != nil {
			log.Println("[error]", err)
			continue
		}
		decoder := json.NewDecoder(socket)
		host := libscore.Info{}
		err = decoder.Decode(&host)
		if err != nil {
			log.Println("[error]", err)
			continue
		}

		hostsChan <- host
		socket.Close()
	}
}

func notifyCluster(host *libscore.Info) {
	json, err := json.Marshal(host)
	if err != nil {
		log.Println("[error]", err)
	}
	cmd := exec.Command("heaverd-tracker-query", "notify", fmt.Sprintf("%s", json))
	err = cmd.Run()
	if err != nil {
		log.Fatal("[error]", err)
	}
}

func selfRefreshing(host *libscore.Info) {
	for {
		err := host.Refresh()
		if err != nil {
			log.Println("[error]", err)
		}
		time.Sleep(time.Second)
	}
}

func getCluster() map[string]libscore.Info {
	result := make(map[string]libscore.Info)
	for name, host := range cluster {
		result[name] = host.info
	}
	return result
}

func GetPreferedHost(containerName string) (string, error) {
	segments := libscore.Segments(getCluster())
	host, err := libscore.ChooseHost(containerName, segments)
	if err != nil {
		return "", err
	}
	return host, nil
}
