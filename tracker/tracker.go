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

type (
	MessageHeader struct {
		MessageType string
	}
	HostinfoMessage struct {
		MessageHeader
		libscore.Hostinfo
	}
	IntentMessage struct {
		MessageHeader
		Intent
	}
	ContainerCreateMessage struct {
		MessageHeader
		Intent
	}
	Intent struct {
		Id            int
		ContainerName string
		CreatedAt     int64
	}
	Host struct {
		info     libscore.Hostinfo
		lastSeen int64
		stale    bool
	}
)

var (
	cluster     = make(map[string]*Host)
	intents     = make(map[int]Intent)
	hostsChan   = make(chan libscore.Hostinfo)
	intentsChan = make(chan Intent)
)

func Start(port string) {
	log.Println("started at port", port)
	go clusterListening(port)
	go clusterUpdating()

	localhostInfo := &libscore.Hostinfo{}
	err := localhostInfo.Refresh()
	if err != nil {
		log.Fatal("[error]", err)
	}
	go func() {
		for {
			err := localhostInfo.Refresh()
			if err != nil {
				log.Println("[error]", err)
			}
			time.Sleep(time.Second)
		}
	}()

	for {
		notifyCluster(localhostInfo)
	}
}

func clusterUpdating() {
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
			log.Println("new intent", intent.Id)
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

func clusterListening(port string) {
	messageListener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("[error]", err)
	}
	for {
		messageSocket, err := messageListener.Accept()
		defer messageSocket.Close()

		if err != nil {
			log.Println("[error]", err)
			continue
		}

		message := make([]byte, 1024)
		messageSocket.Read(message)

		decoder := json.NewDecoder(strings.NewReader(string(message)))
		header := MessageHeader{}
		decoder.Decode(&header)

		decoder = json.NewDecoder(strings.NewReader(string(message)))
		switch header.MessageType {
		case "host":
			host := libscore.Hostinfo{}
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
			if isContainerNameUnique(intent.ContainerName) == true &&
				hasSameIntent(intent.ContainerName) == false {
				intentsChan <- intent
				fmt.Fprintf(messageSocket, fmt.Sprintf("OK"))
			} else {
				// TODO причина отказа
				fmt.Fprintf(messageSocket, fmt.Sprintf("fail"))
			}
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

func hasSameIntent(intentContainerName string) bool {
	for _, intent := range intents {
		if intent.ContainerName == intentContainerName {
			return true
		}
	}
	return false
}

func isContainerNameUnique(containerName string) bool {
	for _, host := range cluster {
		for _, container := range host.info.Containers {
			if container.Name == containerName {
				return false
			}
		}
	}
	return true
}

func GetCluster() map[string]libscore.Hostinfo {
	result := make(map[string]libscore.Hostinfo)
	for name, host := range cluster {
		result[name] = host.info
	}
	return result
}

func notifyCluster(host *libscore.Hostinfo) {
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
