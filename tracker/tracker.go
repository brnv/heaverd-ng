package tracker

import (
	"encoding/json"
	"fmt"
	"heaverd-ng/heaver"
	"heaverd-ng/libscore"
	"log"
	"net"
	"os/exec"
	"time"
)

type (
	Intent struct {
		Id            string
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
	cluster       = make(map[string]*Host)
	intents       = make(map[string]Intent)
	hostsChan     = make(chan libscore.Hostinfo)
	intentsChan   = make(chan Intent)
	localhostInfo = &libscore.Hostinfo{}
)

func Start(port string) {
	go clusterListening(port)
	go clusterUpdating()

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
		notifyCluster()
	}
}

func Cluster() map[string]libscore.Hostinfo {
	result := make(map[string]libscore.Hostinfo)
	for name, host := range cluster {
		result[name] = host.info
	}
	return result
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
			log.Println("new intent", intent.Id, intent.ContainerName)
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
	message := struct {
		Type string
		Body json.RawMessage
	}{}

	messageListener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("[error]", err)
	}

	log.Println("started at port :", port)

	for {
		messageSocket, err := messageListener.Accept()
		if err != nil {
			log.Println("[error]", err)
			continue
		}

		go func() {
			defer messageSocket.Close()
			err = json.NewDecoder(messageSocket).Decode(&message)
			if err != nil {
				log.Println("[error]", err)
			}
			switch message.Type {
			case "hostinfo-update":
				host := libscore.Hostinfo{}
				err := json.Unmarshal(message.Body, &host)
				if err != nil {
					log.Println("[error]", err)
				}
				hostsChan <- host
			case "container-create-intent":
				intent := Intent{}
				err := json.Unmarshal(message.Body, &intent)
				if err != nil {
					log.Println("[error]", err)
				}
				if !existContainer(intent.ContainerName) &&
					!existIntent(intent.ContainerName) {
					intentsChan <- intent
					fmt.Fprintf(messageSocket, fmt.Sprintf("ok"))
				} else {
					// TODO причина отказа
					fmt.Fprintf(messageSocket, fmt.Sprintf("not_unique_name"))
				}
			case "container-create":
				intent := Intent{}
				err := json.Unmarshal(message.Body, &intent)
				if err != nil {
					log.Println("[error]", err)
				}
				if i, ok := intents[intent.Id]; ok {
					log.Println("approved intent", intents[intent.Id])
					log.Println("creating container", intents[intent.Id].ContainerName)
					container := heaver.Create(i.ContainerName)
					container.Host = localhostInfo.Hostname
					result, _ := json.Marshal(container)
					fmt.Fprintf(messageSocket, string(result))
				}
			case "container-control":
				var Control struct {
					ContainerName string
					Action        string
				}
				err := json.Unmarshal(message.Body, &Control)
				if err != nil {
					log.Println("[error]", err)
				}
				switch heaver.Control(Control.ContainerName, Control.Action) {
				case true:
					fmt.Fprintf(messageSocket, "true")
				case false:
					fmt.Fprintf(messageSocket, "false")
				}
			default:
				log.Println("unknown message")
			}
		}()
	}
}

func existIntent(intentContainerName string) bool {
	for _, intent := range intents {
		if intent.ContainerName == intentContainerName {
			return true
		}
	}
	return false
}

func existContainer(containerName string) bool {
	for _, host := range cluster {
		for _, container := range host.info.Containers {
			if container.Name == containerName {
				return true
			}
		}
	}
	return false
}

func notifyCluster() {
	message, err := json.Marshal(struct {
		Type string
		Body interface{}
	}{
		"hostinfo-update",
		localhostInfo,
	})
	if err != nil {
		log.Println("[error]", err)
	}
	cmd := exec.Command("heaverd-tracker-query", "-action=notify", "-message="+string(message))
	err = cmd.Run()
	if err != nil {
		log.Fatal("[error]", err)
	}
}
