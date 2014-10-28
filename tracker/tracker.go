package tracker

import (
	"encoding/json"

	"github.com/brnv/go-heaver"
	"github.com/brnv/go-lxc"
	"github.com/brnv/heaverd-ng/libscore"
	"github.com/coreos/go-etcd/etcd"
	"github.com/op/go-logging"

	"errors"
	"net"
	"time"
)

var (
	etcdc                 = &etcd.Client{}
	Hostinfo              = &libscore.Hostinfo{}
	intentContainerStatus = "pending"
	log                   = logging.MustGetLogger("heaverd-ng")
	containerLogin        = "root"
	containerPassword     = "123123"
	clusterPort           string
	clusterPools          []string
	etcdPort              string
)

type Intent struct {
	Image         []string `json:"image"`
	Key           string   `json:"key"`
	ContainerName string
	PoolName      string
	TargetHost    string
}

func Run(params map[string]interface{}) {
	clusterPort = params["clusterPort"].(string)
	Hostinfo.Pools = params["clusterPools"].([]string)
	etcdPort = params["etcdPort"].(string)

	listener, err := net.Listen("tcp", ":"+clusterPort)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Info("started at port: %s", clusterPort)
	go messageListening(listener)

	etcdc = getEtcdClient()
	_, err = etcdc.CreateDir("hosts/", 0)
	_, err = etcdc.CreateDir("containers/", 0)
	log.Info("etcd port: %s", etcdPort)

	for {
		err = hostinfoUpdate()
		if err != nil {
			log.Notice(err.Error())
			etcdc = getEtcdClient()
		}
		time.Sleep(time.Second)
	}
}

func CreateIntent(intent Intent) error {
	intentContainer, _ := json.Marshal(lxc.Container{
		Name:   intent.ContainerName,
		Host:   intent.TargetHost,
		Status: intentContainerStatus,
		Image:  intent.Image,
		Key:    intent.Key,
	})
	_, err := etcdc.Create("containers/"+intent.ContainerName, string(intentContainer), 5)
	if err != nil {
		return err
	}
	return nil
}

func Cluster(poolName ...string) map[string]libscore.Hostinfo {
	result := make(map[string]libscore.Hostinfo)

	hosts, err := etcdc.Get("hosts/", true, true)
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

func messageListening(listener net.Listener) {
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
						_, _ = etcdc.Delete(
							"containers/"+Control.ContainerName, false)
					}
					messageSocket.Write([]byte("ok"))
				}
			default:
				log.Notice("unknown message %s", message)
			}
		}()
	}
}

func createContainer(name string) (newContainer lxc.Container, err error) {
	rawContainer, _ := etcdc.Get("containers/"+name, false, false)

	container := lxc.Container{}
	err = json.Unmarshal([]byte(rawContainer.Node.Value), &container)
	if err != nil {
		return newContainer, err
	}

	if container.Status != intentContainerStatus {
		return newContainer, errors.New("Container is " +
			container.Status + ", not " + intentContainerStatus)
	}

	log.Notice("creating container %s on host %s ...", name, Hostinfo.Hostname)

	_, err = etcdc.Delete("containers/"+name, false)
	if err != nil {
		return newContainer, err
	}

	newContainer, err = heaver.Create(container.Name, container.Image, container.Key)
	if err != nil {
		return newContainer, err
	}

	newContainer.Host = Hostinfo.Hostname

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
	return etcd.NewClient([]string{"http://localhost:" + etcdPort})
}
