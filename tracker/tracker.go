package tracker

import (
	"encoding/json"
	"fmt"
	"heaverd-ng/libscore"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"
)

type MessageHeader struct {
	MessageType string
}

type HostinfoMessage struct {
	MessageHeader
	libscore.Info
}

type IntentMessage struct {
	MessageHeader
	Intent
}

type ContainerCreateMessage struct {
	MessageHeader
	Intent
}

type Host struct {
	info     libscore.Info
	lastSeen int64
	stale    bool
}

type Intent struct {
	Id            int
	ContainerName string
	Creationtime  int64
}

var (
	cluster = make(map[string]*Host)
	intents = make(map[int]Intent)
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
	intentsChan := make(chan Intent)
	go clusterListening(listener, hostsChan, intentsChan)
	go clusterRefreshing(hostsChan, intentsChan)

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

func clusterRefreshing(hostsChan chan libscore.Info, intentsChan chan Intent) {
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
		case intent := <-intentsChan:
			intents[intent.Id] = intent
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

func clusterListening(
	conn net.Listener,
	hostsChan chan libscore.Info,
	intentsChan chan Intent) {
	for {
		socket, err := conn.Accept()
		defer socket.Close()

		if err != nil {
			log.Println("[error]", err)
			continue
		}

		message := make([]byte, 1024)
		socket.Read(message)

		decoder := json.NewDecoder(strings.NewReader(string(message)))
		header := MessageHeader{}
		decoder.Decode(&header)

		decoder = json.NewDecoder(strings.NewReader(string(message)))
		switch header.MessageType {
		case "host":
			host := libscore.Info{}
			err = decoder.Decode(&host)
			if err != nil {
				log.Println("[error]", err)
				continue
			}
			hostsChan <- host
		case "intent":
			intent := Intent{}
			err = decoder.Decode(&intent)
			if err != nil {
				log.Println("[error]", err)
				continue
			}
			intentsChan <- intent
		case "container-create":
			intent := Intent{}
			err = decoder.Decode(&intent)
			if err != nil {
				log.Println("[error]", err)
				continue
			}
			if _, ok := intents[intent.Id]; ok {
				log.Println("Ready to create")
			}
		}
	}
}

func notifyCluster(host *libscore.Info) {
	message := HostinfoMessage{
		MessageHeader{MessageType: "host"},
		*host,
	}
	json, err := json.Marshal(message)
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

func GetCluster() map[string]libscore.Info {
	result := make(map[string]libscore.Info)
	for name, host := range cluster {
		result[name] = host.info
	}
	return result
}
