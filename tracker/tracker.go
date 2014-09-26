package tracker

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"heaverd-ng/heaver"
	"heaverd-ng/libscore"
	"heaverd-ng/libstats/lxc"
	"net"
	"os"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/op/go-logging"
	"github.com/zazab/zhash"
)

var (
	etcdc                 = &etcd.Client{}
	Hostinfo              = &libscore.Hostinfo{}
	intentContainerStatus = "pending"
	log                   = &logging.Logger{}
)

var config = zhash.NewHash()

func Run(configPath string, logger *logging.Logger) {
	log = logger
	err := readConfig(configPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	port, _ := config.GetString("cluster", "port")
	Hostinfo.Pools, _ = config.GetStringSlice("cluster", "pools")

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Info("started at port: %s", port)
	go messageListening(listener)

	etcdc = getEtcdClient()
	_, err = etcdc.CreateDir("hosts/", 0)
	_, err = etcdc.CreateDir("containers/", 0)
	etcdPort, _ := config.GetString("etcd", "port")
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

func CreateIntent(params map[string]string) error {
	intent, _ := json.Marshal(lxc.Container{
		Name:   params["containerName"],
		Host:   params["targetHost"],
		Status: intentContainerStatus,
		Image:  params["image"],
	})

	_, err := etcdc.Create("containers/"+params["containerName"], string(intent), 5)
	if err != nil {
		return err
	}

	log.Info("Intent: host", params["targetHost"])
	log.Info("Intent: container", params["containerName"])
	if params["poolName"] != "" {
		log.Info("Intent: pool", params["poolName"])
	}
	log.Info("Intent: image", params["image"])

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
				log.Error("[error]", err)
			}
			switch message.Type {
			case "container-create":
				var containerName string

				err := json.Unmarshal(message.Body, &containerName)
				if err != nil {
					log.Error("[error]", err)
					fmt.Fprintf(messageSocket, "Error: "+err.Error())
					return
				}

				result, err := createContainer(containerName)
				if err != nil {
					log.Error("[error]", err)
					fmt.Fprintf(messageSocket, "Error: "+err.Error())
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
					log.Error("[error]", err)
				}
				err = heaver.Control(Control.ContainerName, Control.Action)
				if err != nil {
					fmt.Fprintf(messageSocket, "Error: "+err.Error())
				} else {
					fmt.Fprintf(messageSocket, "ok")
				}
			default:
				log.Notice("unknown message")
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

	log.Notice("creating container", name, "on host", Hostinfo.Hostname)

	_, err = etcdc.Delete("containers/"+name, false)
	if err != nil {
		return "", err
	}

	newContainer, err := heaver.Create(container.Name, container.Image)
	if err != nil {
		return "", err
	}

	newContainer.Host = Hostinfo.Hostname

	err = hostinfoUpdate()
	if err != nil {
		log.Error(err.Error())
	}

	result, _ := json.Marshal(newContainer)
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
	port, _ := config.GetString("etcd", "port")
	return etcd.NewClient([]string{"http://localhost:" + port})
}

func readConfig(path string) error {
	f, err := os.Open(path)
	if err == nil {
		config.ReadHash(bufio.NewReader(f))
		return nil
	} else {
		return err
	}
}
