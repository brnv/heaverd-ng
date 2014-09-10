package webserver

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"sync"

	"heaverd-ng/libscore"
	"heaverd-ng/tracker"

	"github.com/gocraft/web"
)

type Context struct {
	peerAddr string
}

func Start(wg *sync.WaitGroup, webAddr string, peerAddr string, seed int64) {
	rand.Seed(seed)

	context := &Context{
		peerAddr: peerAddr,
	}

	router := web.New(Context{}).
		Middleware(web.LoggerMiddleware).
		Middleware(web.ShowErrorsMiddleware).
		Get("/", handleHelp).
		Get("/stats/", handleStats).
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
		Post("/c/:cid/start", context.handleContainerStart).
		Post("/h/:hid/:cid/stop", context.handleContainerStop).
		//Post("/h/:hid/:cid/freeze", )
		//Post("/h/:hid/:cid/unfreeze", )
		//Get("/h/:hid/:cid/tarball", )
		Get("/h/:hid/:cid/ping", handleContainerPing)
	//Get("/h/:hid/:cid/attach", )

	log.Println("started at port:", webAddr)
	log.Fatal(http.ListenAndServe(":"+webAddr, router))
	wg.Done()
}

func handleHelp(w web.ResponseWriter, r *web.Request) {
	fmt.Fprintf(w, "Справка по API в запрошенном формате")
}

func handleStats(w web.ResponseWriter, r *web.Request) {
	stats, _ := json.Marshal(tracker.Hostinfo)
	fmt.Fprint(w, string(stats))
}

func handleHostList(w web.ResponseWriter, r *web.Request) {
	cluster, _ := json.Marshal(tracker.Cluster())
	fmt.Fprint(w, string(cluster))
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

	targetHost, err := getPreferedHost(containerName)
	if err != nil {
		log.Println("[error]", err)
		http.Error(w, fmt.Sprintf("%v", err), 502)
		return
	}

	ready := tracker.CreateIntent(targetHost, containerName)

	if ready != true {
		logMessage := "Not unique container name"
		log.Println(logMessage)
		http.Error(w, logMessage, 502)
		return
	}

	createMessage := makeMessage("container-create", containerName)

	hostAnswer, _ := sendTcpMessage(targetHost+":"+c.peerAddr, createMessage)
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

	controlMessage := makeMessage("container-control", struct {
		ContainerName string
		Action        string
	}{
		containerName,
		"destroy",
	})

	answer, _ := sendTcpMessage(hostname+":"+c.peerAddr, controlMessage)
	switch answer {
	case "done":
		http.Error(w, "", 204)
	case "not_done":
		http.Error(w, "Not destroyed", 504)
	}
}

func handleContainerInfo(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Получить инфороцию о контейнере", 501)
}

func (c *Context) handleContainerStart(w web.ResponseWriter, r *web.Request) {
	var hostname string
	containerName := r.PathParams["cid"]

	if _, ok := r.PathParams["hid"]; ok {
		hostname = r.PathParams["hid"]
	} else {
		hostname = getHostnameByContainer(containerName)
		if hostname == "" {
			http.Error(w, "Unknown container", 404)
			return
		}
	}

	if !checkHostname(hostname) {
		http.Error(w, "Unknown host", 404)
		return
	}

	if !checkContainer(hostname, containerName) {
		http.Error(w, "Unknown container", 404)
		return
	}

	controlMessage := makeMessage("container-control", struct {
		ContainerName string
		Action        string
	}{
		containerName,
		"start",
	})

	answer, _ := sendTcpMessage(hostname+":"+c.peerAddr, controlMessage)
	switch answer {
	case "done":
		http.Error(w, "", 204)
	case "not done":
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

	controlMessage := makeMessage("container-control", struct {
		ContainerName string
		Action        string
	}{
		containerName,
		"stop",
	})

	answer, _ := sendTcpMessage(hostname+":"+c.peerAddr, controlMessage)
	switch answer {
	case "done":
		http.Error(w, "", 204)
	case "not_done":
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

func sendTcpMessage(host string, message string) (string, error) {
	connection, err := net.Dial("tcp", host)
	defer connection.Close()
	if err != nil {
		return "", err
	}
	fmt.Fprint(connection, message)
	answer, err := bufio.NewReader(connection).ReadString('\n')
	if err != nil {
		if err != io.EOF {
			return "", err
		}
	}
	return string(answer), nil
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

func getHostnameByContainer(containerName string) string {
	for hostname, host := range tracker.Cluster() {
		if _, ok := host.Containers[containerName]; ok {
			return hostname
		}
	}
	return ""
}

func makeMessage(header string, body interface{}) string {
	message, _ := json.Marshal(struct {
		Type string
		Body interface{}
	}{
		header,
		body,
	})
	return string(message)
}
