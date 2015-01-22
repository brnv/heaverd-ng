package main

import (
	"encoding/json"

	"github.com/brnv/heaverd-ng/libheaver"
	"github.com/brnv/heaverd-ng/liblxc"
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
	storedKeyTtl          uint64
)

func ClusterRun(params map[string]interface{}) {
	Hostinfo.Pools = params["clusterPools"].([]string)
	storage = getEtcdClient(params["etcdPort"].(string))
	storedKeyTtl = uint64(params["etcdKeyTtl"].(int64))

	go listenForMessages(params["clusterPort"].(string))

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
			log.Fatal(err)
		}
	}
}

func updateContainers(containers map[string]lxc.Container) error {
	for _, c := range containers {
		container, _ := json.Marshal(c)
		_, err := storage.Set("containers/"+c.Name, string(container), storedKeyTtl)
		if err != nil {
			return err
		}
	}
	return nil
}

func StoreRequestAsIntent(request ContainerCreateRequest) error {
	intentContainer, _ := json.Marshal(lxc.Container{
		Name:   request.ContainerName,
		Host:   request.RequestHost,
		Status: intentContainerStatus,
		Image:  request.Image,
		Key:    request.SshKey,
		Ip:     request.Ip,
	})
	_, err := storage.Create("containers/"+request.ContainerName, string(intentContainer), storedKeyTtl)
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

	for {
		request := make(map[string]interface{})

		socket, err := listener.Accept()
		if err != nil {
			log.Error(err.Error())
			continue
		}
		defer socket.Close()

		err = json.NewDecoder(socket).Decode(&request)
		if err != nil {
			log.Error(err.Error())
		}

		timestamp := time.Now().UnixNano()

		log.Info("RECEIVER: %d %v", timestamp, request)

		var errorMessage string

		switch request["Action"] {
		case "create":
			container, err := createContainer(request["ContainerName"].(string))
			if err != nil {
				errorMessage = err.Error()
			}

			response, _ := json.Marshal(ClusterResponse{
				BaseResponse: BaseResponse{
					ResponseHost: Hostinfo.Hostname,
					Error:        errorMessage,
				},
				Container: container,
			})

			socket.Write(response)

		case "start":
			err := heaver.Start(
				request["ContainerName"].(string),
			)

			if err != nil {
				errorMessage = err.Error()
			}

			response, _ := json.Marshal(ClusterResponse{
				BaseResponse: BaseResponse{
					ResponseHost: Hostinfo.Hostname,
					Error:        errorMessage,
				},
			})

			socket.Write(response)

		case "stop":
			err := heaver.Stop(
				request["ContainerName"].(string),
			)

			if err != nil {
				errorMessage = err.Error()
			}

			response, _ := json.Marshal(ClusterResponse{
				BaseResponse: BaseResponse{
					ResponseHost: Hostinfo.Hostname,
					Error:        errorMessage,
				},
			})

			socket.Write(response)

		case "destroy":
			err := heaver.Destroy(
				request["ContainerName"].(string),
			)

			if err != nil {
				errorMessage = err.Error()
			}

			response, _ := json.Marshal(ClusterResponse{
				BaseResponse: BaseResponse{
					ResponseHost: Hostinfo.Hostname,
					Error:        errorMessage,
				},
				Token: timestamp,
			})

			storage.Delete(
				"containers/"+request["ContainerName"].(string), false)

			socket.Write(response)

		case "push":
			err := heaver.Push(
				request["ContainerName"].(string),
				request["Image"].(string),
			)

			if err != nil {
				errorMessage = err.Error()
			}

			response, _ := json.Marshal(ContainerPushResponse{
				BaseResponse: BaseResponse{
					ResponseHost: Hostinfo.Hostname,
					Error:        errorMessage,
				},
			})

			socket.Write(response)

		default:
			socket.Write(answer(400, "error", "", timestamp))
			log.Info("RECEIVER: %d UNDEFINED", timestamp)
		}

		socket.Close()

		log.Info("RECEIVER: %d DONE", timestamp)
	}
}

func answer(code int, text string, err string, time int64) []byte {
	answer, _ := json.Marshal(struct {
		From  string
		Msg   string
		Error string
		Token int64
		Code  int
	}{
		From:  Hostinfo.Hostname,
		Msg:   text,
		Error: err,
		Token: time,
		Code:  code,
	})
	return answer
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

	_, err = storage.Delete("containers/"+name, false)
	if err != nil {
		return lxc.Container{}, err
	}

	newContainer, err := heaver.Create(
		container.Name, container.Image, container.Key, container.Ip,
	)
	newContainer.Host = Hostinfo.Hostname
	if err != nil {
		return newContainer, err
	}

	err = hostinfoUpdate()
	if err != nil {
		log.Error(err.Error())
	}

	return newContainer, nil
}

func hostinfoUpdate() error {
	containers, err := heaver.ListContainers(Hostinfo.Hostname)
	if err != nil {
		return err
	}

	err = updateContainers(containers)
	if err != nil {
		return err
	}

	err = Hostinfo.Refresh()
	if err != nil {
		return err
	}

	host, _ := json.Marshal(Hostinfo)
	_, err = storage.Set("hosts/"+Hostinfo.Hostname, string(host), storedKeyTtl)
	if err != nil {
		return err
	}
	return nil
}

func getEtcdClient(port string) *etcd.Client {
	log.Info("obtaining etcd client on :%s", port)
	return etcd.NewClient([]string{"http://localhost:" + port})
}
