package main

import (
	"encoding/json"

	"github.com/brnv/go-heaver"
	"github.com/brnv/go-lxc"
	"github.com/brnv/heaverd-ng/libscore"
	"github.com/coreos/go-etcd/etcd"

	"errors"
	"net"
	"time"
)

var (
	storage               = &etcd.Client{}
	Hostinfo              = &libscore.Hostinfo{}
	intentContainerStatus = "pending"
	containerLogin        = "root"
	containerPassword     = "123123"
	clusterPools          []string
	storedKeyTtl          = uint64(10)
)

type Intent struct {
	Image         []string `json:"image"`
	Key           string   `json:"key"`
	ContainerName string
	PoolName      string
	TargetHost    string
}

func ClusterRun(params map[string]interface{}) {
	go listenForMessages(params["clusterPort"].(string))

	Hostinfo.Pools = params["clusterPools"].([]string)

	storage = getEtcdClient(params["etcdPort"].(string))

	err := Hostinfo.Refresh()
	if err != nil {
		log.Fatal(err.Error())
	}

	storeContainers(Hostinfo.Containers)

	for {
		err = hostinfoUpdate()
		if err != nil {
			log.Error(err.Error())
			storage = getEtcdClient(params["etcdPort"].(string))
		}
		time.Sleep(time.Second)
	}
}

func storeContainers(containers map[string]lxc.Container) {
	for _, c := range containers {
		container, _ := json.Marshal(c)
		_, err := storage.Create("containers/"+c.Name, string(container), storedKeyTtl)
		if err != nil {
			log.Warning("%v", err)
		}
	}
}

func updateContainers(containers map[string]lxc.Container) error {
	for _, c := range containers {
		c.KeyModifiedTimestamp = time.Now().UnixNano()
		container, _ := json.Marshal(c)
		_, err := storage.Update("containers/"+c.Name, string(container), storedKeyTtl)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateIntent(intent Intent) error {
	intentContainer, _ := json.Marshal(lxc.Container{
		Name:   intent.ContainerName,
		Host:   intent.TargetHost,
		Status: intentContainerStatus,
		Image:  intent.Image,
		Key:    intent.Key,
	})
	_, err := storage.Create("containers/"+intent.ContainerName, string(intentContainer), 5)
	if err != nil {
		return err
	}
	return nil
}

func Cluster(poolName ...string) map[string]libscore.Hostinfo {
	result := make(map[string]libscore.Hostinfo)

	hosts, err := storage.Get("hosts/", true, true)
	if err != nil {
		log.Error(err.Error())
		return result
	}

	for _, node := range hosts.Node.Nodes {
		host := libscore.Hostinfo{}
		err := json.Unmarshal([]byte(node.Value), &host)
		if err != nil {
			log.Error(err.Error())
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

func listenForMessages(port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Info("listening for messages on :%s", port)

	message := struct {
		Type string
		Body json.RawMessage
	}{}

	for {
		messageSocket, err := listener.Accept()
		if err != nil {
			log.Error(err.Error())
			continue
		}

		go func() {
			defer messageSocket.Close()
			err = json.NewDecoder(messageSocket).Decode(&message)
			if err != nil {
				log.Error(err.Error())
			}
			switch message.Type {
			case "container-create":
				var containerName string

				err := json.Unmarshal(message.Body, &containerName)
				if err != nil {
					log.Error(err.Error())
					messageSocket.Write([]byte("Error:" + err.Error()))
					return
				}

				newContainer, err := createContainer(containerName)
				if err != nil {
					log.Error(err.Error())
					messageSocket.Write([]byte("Error:" + err.Error()))
					return
				}

				result, _ := json.Marshal(newContainer)
				messageSocket.Write(result)
			case "container-control":
				var Control struct {
					ContainerName string
					Action        string
				}
				err := json.Unmarshal(message.Body, &Control)
				if err != nil {
					log.Error(err.Error())
				}
				err = heaver.Control(Control.ContainerName, Control.Action)
				if err != nil {
					messageSocket.Write([]byte("Error:" + err.Error()))
				} else {
					if Control.Action == "destroy" {
						_, _ = storage.Delete(
							"containers/"+Control.ContainerName, false)
					}

					err = hostinfoUpdate()
					if err != nil {
						messageSocket.Write([]byte("Error:" + err.Error()))
						return
					}

					messageSocket.Write([]byte("ok"))
				}
			default:
				log.Notice("unknown message %s", message)
			}
		}()
	}
}

func createContainer(name string) (lxc.Container, error) {
	rawContainer, err := storage.Get("containers/"+name, false, false)
	if err != nil {
		return lxc.Container{}, err
	}

	container := lxc.Container{}
	err = json.Unmarshal([]byte(rawContainer.Node.Value), &container)
	if err != nil {
		return lxc.Container{}, err
	}

	if container.Status != intentContainerStatus {
		return lxc.Container{}, errors.New("Container is " +
			container.Status + ", not " + intentContainerStatus)
	}

	log.Notice("creating container %s on host %s...", name, Hostinfo.Hostname)

	_, err = storage.Delete("containers/"+name, false)
	if err != nil {
		return lxc.Container{}, err
	}

	newContainer, err := heaver.Create(container.Name, container.Image, container.Key)
	newContainer.Host = Hostinfo.Hostname
	if err != nil {
		return newContainer, err
	}

	err = hostinfoUpdate()
	if err != nil {
		log.Error(err.Error())
	}

	log.Notice("... done")

	return newContainer, nil
}

func hostinfoUpdate() error {
	err := Hostinfo.Refresh()
	if err != nil {
		return err
	}

	host, _ := json.Marshal(Hostinfo)
	_, err = storage.Set("hosts/"+Hostinfo.Hostname, string(host), storedKeyTtl)
	if err != nil {
		return err
	}

	err = updateContainers(Hostinfo.Containers)
	if err != nil {
		return err
	}

	return nil
}

func getEtcdClient(port string) *etcd.Client {
	log.Info("obtaining etcd client on :%s", port)
	return etcd.NewClient([]string{"http://localhost:" + port})
}
