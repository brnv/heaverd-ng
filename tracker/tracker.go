package tracker

import (
	"encoding/json"
	"errors"
	"fmt"
	"heaverd-ng/heaver"
	"heaverd-ng/libscore"
	"heaverd-ng/libstats/lxc"
	"log"
	"net"
	"sync"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/zazab/zhash"
)

var (
	etcdc                 = &etcd.Client{}
	Hostinfo              = &libscore.Hostinfo{}
	intentContainerStatus = "pending"
	defaultPools          = []string{"test"}
)

var Config = zhash.NewHash()

func Start(wg *sync.WaitGroup) {
	port, _ := Config.GetString("cluster", "port")
	pools, err := Config.GetStringSlice("cluster", "pools")
	if err != nil {
		pools = defaultPools
	}
	Hostinfo.Pools = pools

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Println("[error]", err)
		wg.Done()
	}

	log.Println("started at port:", port)
	go messageListening(listener)

	etcdc = getEtcdClient()
	_, err = etcdc.CreateDir("hosts/", 0)
	_, err = etcdc.CreateDir("containers/", 0)
	etcdPort, _ := Config.GetString("etcd", "port")
	log.Println("etcd port:", etcdPort)

	for {
		err = hostinfoUpdate()
		if err != nil {
			log.Println("[error]", err)
			etcdc = getEtcdClient()
		}
		time.Sleep(time.Second)
	}
}

func CreateIntent(targetHost string, containerName string, poolName string) bool {
	intent, _ := json.Marshal(lxc.Container{
		Name:   containerName,
		Host:   targetHost,
		Status: intentContainerStatus,
	})

	_, err := etcdc.Create("containers/"+containerName, string(intent), 5)
	if err != nil {
		log.Println("[error]", err)
		return false
	}

	log.Println("Intent: host", targetHost+", container", containerName,
		"pool", poolName)
	return true
}

func Cluster(poolName ...string) map[string]libscore.Hostinfo {
	result := make(map[string]libscore.Hostinfo)

	hosts, err := etcdc.Get("hosts/", true, true)
	if err != nil {
		log.Println("[error]", err)
		return result
	}

	for _, node := range hosts.Node.Nodes {
		host := libscore.Hostinfo{}
		err := json.Unmarshal([]byte(node.Value), &host)
		if err != nil {
			log.Println("[error]", err)
		}
		if poolName != nil && poolName[0] != "" {
			for _, h := range host.Pools {
				if h == poolName[0] {
					result[host.Hostname] = host
				}
			}
		} else {
			result[host.Hostname] = host
		}
	}

	return result
}

func messageListening(listener net.Listener) {
	message := struct {
		Type string
		Body json.RawMessage
	}{}

	for {
		messageSocket, err := listener.Accept()
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
			case "container-create":
				var containerName string

				err := json.Unmarshal(message.Body, &containerName)
				if err != nil {
					log.Println("[error]", err)
					fmt.Fprintf(messageSocket, fmt.Sprintf("%v", err))
					return
				}

				result, err := createContainer(containerName)
				if err != nil {
					log.Println("[error]", err)
					fmt.Fprintf(messageSocket, "Heaver error, check logs on "+
						Hostinfo.Hostname)
					return
				}
				fmt.Fprintf(messageSocket, result)
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
					fmt.Fprintf(messageSocket, "done")
				case false:
					fmt.Fprintf(messageSocket, "not_done")
				}
			default:
				log.Println("unknown message")
			}
		}()
	}
}

func createContainer(name string) (string, error) {
	rawContainer, _ := etcdc.Get("containers/"+name, false, false)

	container := lxc.Container{}
	err := json.Unmarshal([]byte(rawContainer.Node.Value), &container)

	if err != nil {
		return "", err
	}

	if container.Status != intentContainerStatus {
		return "", errors.New("Container is " + container.Status + ", not " +
			intentContainerStatus)
	}

	log.Println("creating container", name, "on host", Hostinfo.Hostname)

	_, err = etcdc.Delete("containers/"+name, false)
	if err != nil {
		return "", err
	}

	created, err := heaver.Create(name)
	if err != nil {
		return "", err
	}

	created.Host = Hostinfo.Hostname

	err = hostinfoUpdate()
	if err != nil {
		log.Println("[error]", err)
	}

	result, _ := json.Marshal(created)
	return string(result), nil
}

func hostinfoUpdate() error {
	err := Hostinfo.Refresh()
	if err != nil {
		return err
	}

	host, _ := json.Marshal(Hostinfo)
	_, err = etcdc.Set("hosts/"+Hostinfo.Hostname, string(host), 5)
	if err != nil {
		return err
	}

	for _, c := range Hostinfo.Containers {
		container, _ := json.Marshal(c)
		_, err = etcdc.Set("containers/"+c.Name, string(container), 5)
		if err != nil {
			return err
		}
	}

	return nil
}

func getEtcdClient() *etcd.Client {
	port, _ := Config.GetString("etcd", "port")
	return etcd.NewClient([]string{"http://localhost:" + port})
}
