package webserver

import (
	"bufio"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"heaverd-ng/libscore"
	"heaverd-ng/tracker"

	"github.com/gocraft/web"
)

type Context struct {
	clusterListen string
}

func Start(port string, clusterListen string, seed int64) {
	rand.Seed(seed)

	context := &Context{
		clusterListen: clusterListen,
	}

	router := web.New(Context{}).
		Middleware(web.LoggerMiddleware).
		Middleware(web.ShowErrorsMiddleware).

		//http://confluence.rn/display/ENV/RESTful+API
		//Common
		Get("/", handleHelp).
		Get("/stats/", handleStats).
		//Hosts
		Get("/h/", handleHostList).
		Post("/h/:hid", handleHostOperation).
		Get("/h/:hid", handleHostContainersList).
		Get("/h/:hid/ping", handleHostPing).
		//Containers
		Get("/c/", handleClusterContainersList).
		Get("/c/:cid", handleHostByContainer).
		Post("/c/:cid", context.handleContainerCreate).
		Post("/h/:hid/:cid", handleHostContainerCreate).
		Put("/h/:hid/:cid", handleHostContainerUpdate).
		Delete("/h/:hid/:cid", context.handleContainerDestroy).
		Get("/h/:hid/:cid", handleContainerInfo).
		Post("/h/:hid/:cid/start", context.handleContainerStart).
		Post("/h/:hid/:cid/stop", context.handleContainerStop).
		//Post("/h/:hid/:cid/freeze", )
		//Post("/h/:hid/:cid/unfreeze", )
		//Get("/h/:hid/:cid/tarball", )
		Get("/h/:hid/:cid/ping", handleContainerPing)
	//Get("/h/:hid/:cid/attach", )

	log.Println("started at port :", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func handleHelp(w web.ResponseWriter, r *web.Request) {
	fmt.Fprintf(w, "Справка по API в запрошенном формате")
}

func handleStats(w web.ResponseWriter, r *web.Request) {
	cluster, _ := json.Marshal(tracker.Cluster())
	fmt.Fprint(w, string(cluster))
}

func handleHostList(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 501)
}

func handleHostOperation(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 501)
}

func handleHostContainersList(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 501)
}

func handleHostPing(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "pong", 204)
}

func handleClusterContainersList(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Список всех контейнеров на всех хостах", 501)
}

func handleHostByContainer(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Имя хоста, на котором расположен указанный контейнер", 501)
}

func handleHostInformationRequest(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Полная информация о хосте", 501)
}

func (c *Context) handleContainerCreate(w web.ResponseWriter, r *web.Request) {
	containerName := r.PathParams["cid"]

	h := md5.New()
	fmt.Fprint(h, containerName+strconv.FormatInt(time.Now().Unix(), 10))
	intentId := fmt.Sprintf("%x", h.Sum(nil))

	intentMessage, _ := json.Marshal(struct {
		Type string
		Body interface{}
	}{
		"container-create-intent",
		tracker.Intent{
			Id:            intentId,
			ContainerName: containerName,
		},
	})

	for _, host := range tracker.Cluster() {
		nodeAnswer := sendTcpMessage(host.Hostname+":"+c.clusterListen,
			string(intentMessage))
		switch nodeAnswer {
		case "not_unique_name":
			http.Error(w, "Not unique container name", 409)
			return
		case "intent_still_exists":
			http.Error(w, "Intent still exists", 409)
			return
		}
	}

	targetHost, err := getPreferedHost(containerName)
	if err != nil {
		log.Println("[error]", err)
		http.Error(w, fmt.Sprintf("%v", err), 502)
		return
	}

	createMessage, _ := json.Marshal(struct {
		Type string
		Body interface{}
	}{
		"container-create",
		tracker.Intent{Id: intentId},
	})

	hostAnswer := sendTcpMessage(targetHost+":"+c.clusterListen, string(createMessage))

	http.Error(w, hostAnswer, 201)
}

func handleHostContainerCreate(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Создать контейнер на этом хосте", 501)
}

func handleHostContainerUpdate(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Обновить настройки контейнера", 501)
}

func (c *Context) handleContainerDestroy(w web.ResponseWriter, r *web.Request) {
	hostname := r.PathParams["hid"]
	containerName := r.PathParams["cid"]

	if !checkHostname(hostname) {
		http.Error(w, "Unknown host", 404)
		return
	}

	if !checkContainer(hostname, containerName) {
		http.Error(w, "Unknown container", 404)
		return
	}

	controlMessage, _ := json.Marshal(struct {
		Type string
		Body interface{}
	}{
		"container-control",
		struct {
			ContainerName string
			Action        string
		}{
			containerName,
			"destroy",
		},
	})

	switch sendTcpMessage(hostname+":"+c.clusterListen, string(controlMessage)) {
	case "true":
		http.Error(w, "", 204)
	case "false":
		http.Error(w, "Not destroyed", 504)
	}
}

func handleContainerInfo(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Получить инфороцию о контейнере", 501)
}

func (c *Context) handleContainerStart(w web.ResponseWriter, r *web.Request) {
	hostname := r.PathParams["hid"]
	containerName := r.PathParams["cid"]

	if !checkHostname(hostname) {
		http.Error(w, "Unknown host", 404)
		return
	}

	if !checkContainer(hostname, containerName) {
		http.Error(w, "Unknown container", 404)
		return
	}

	controlMessage, _ := json.Marshal(struct {
		Type string
		Body interface{}
	}{
		"container-control",
		struct {
			ContainerName string
			Action        string
		}{
			containerName,
			"start",
		},
	})

	switch sendTcpMessage(hostname+":"+c.clusterListen, string(controlMessage)) {
	case "true":
		http.Error(w, "", 204)
	case "false":
		http.Error(w, "Not started", 502)
	}
}

func (c *Context) handleContainerStop(w web.ResponseWriter, r *web.Request) {
	hostname := r.PathParams["hid"]
	containerName := r.PathParams["cid"]

	if !checkHostname(hostname) {
		http.Error(w, "Unknown host", 404)
		return
	}

	if !checkContainer(hostname, containerName) {
		http.Error(w, "Unknown container", 404)
		return
	}

	controlMessage, _ := json.Marshal(struct {
		Type string
		Body interface{}
	}{
		"container-control",
		struct {
			ContainerName string
			Action        string
		}{
			containerName,
			"stop",
		},
	})

	switch sendTcpMessage(hostname+":"+c.clusterListen, string(controlMessage)) {
	case "true":
		http.Error(w, "", 204)
	case "false":
		http.Error(w, "Not stopped", 502)
	}
}

func handleContainerPing(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Пингануть сервер", 501)
}

func getPreferedHost(containerName string) (string, error) {
	segments := libscore.Segments(tracker.Cluster())
	host, err := libscore.ChooseHost(containerName, segments)
	if err != nil {
		return "", err
	}
	return host, nil
}

func sendTcpMessage(target string, message string) string {
	connection, err := net.Dial("tcp", target)
	defer connection.Close()
	if err != nil {
		log.Println("[error]", err)
	}

	fmt.Fprint(connection, message)

	answer, err := bufio.NewReader(connection).ReadString('\n')
	if err != nil {
		if err != io.EOF {
			log.Println("[error]", err)
		}
	}

	return string(answer)
}

func checkHostname(name string) bool {
	if _, ok := tracker.Cluster()[name]; !ok {
		return false
	}
	return true
}

func checkContainer(hostname string, name string) bool {
	if _, ok := tracker.Cluster()[hostname].Containers[name]; !ok {
		return false
	}
	return true
}
