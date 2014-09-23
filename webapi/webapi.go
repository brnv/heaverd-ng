package webapi

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
	"text/template"

	"heaverd-ng/libscore"
	"heaverd-ng/tracker"

	"github.com/brnv/web"
	"github.com/zazab/zhash"
)

type (
	Context struct{}
)

var (
	rootRouter = web.New(Context{})
	apiRouter  = rootRouter.Subrouter(Context{}, "/v2")
	Config     = zhash.NewHash()
)

func Start(wg *sync.WaitGroup, seed int64) {
	rand.Seed(seed)

	rootRouter.Middleware(web.LoggerMiddleware).
		Middleware(web.ShowErrorsMiddleware).
		Middleware(web.StaticMiddleware("www")).
		Get("/score", handleScore, "Графики пулов хостов")

	apiRouter.Get("/", handleHelp, "Справка по API").
		Get("/h/:hid/stats", handleStats, "Статистика хоста :hid").
		Get("/h/", handleHostList, "Список всех хостов").
		Post("/c/:cid", handleContainerCreate, "Создать контейнер :cid (балансировка)").
		Post("/p/:poolid/:cid", handleContainerCreate, "Создать контейнер в пуле :poolid").
		Delete("/h/:hid/:cid", handleContainerDestroy, "Удалить контейнер").
		Post("/h/:hid/:cid/start", handleContainerStart, "Стартануть контейнер").
		Post("/c/:cid/start", handleContainerStart, "Стартануть контейнер").
		Post("/h/:hid/:cid/stop", handleContainerStop, "Стоп контейнера").
		Post("/h/:hid", handleHostOperation, "Операции с хостом").
		Get("/h/:hid", handleHostContainersList, "Список контейнеров на хосте").
		Get("/h/:hid/ping", handleHostPing, "Пинг сервера").
		Get("/c/", handleClusterContainersList, "Все контейнеры").
		Get("/c/:cid", handleHostByContainer, "Найти хост по контейнеру").
		Post("/h/:hid/:cid", handleHostContainerCreate, "Создать контейнер на хосте").
		Put("/h/:hid/:cid", handleHostContainerUpdate, "Обновить контейнер").
		Get("/h/:hid/:cid", handleContainerInfo, "Информация о контейнере").
		Get("/h/:hid/:cid/ping", handleContainerPing, "Пинг контейнера")
	//Post("/h/:hid/:cid/freeze", ).
	//Post("/h/:hid/:cid/unfreeze", ).
	//Get("/h/:hid/:cid/tarball", ).
	//Get("/h/:hid/:cid/attach", )

	port, _ := Config.GetString("web", "port")
	log.Println("started at port:", port)
	log.Fatal(http.ListenAndServe(":"+port, rootRouter))
}

func handleScore(w web.ResponseWriter, r *web.Request) {
	r.Header.Set("Content-Type", "text/html")
	templates, _ := Config.GetString("templates", "dir")
	template.Must(template.
		ParseFiles(templates+"/index.tpl")).Execute(w, nil)
}

func handleHelp(w web.ResponseWriter, r *web.Request) {
	r.Header.Set("Content-Type", "text/html")
	j, _ := json.Marshal(apiRouter)
	fmt.Fprintf(w, string(j))
}

func handleStats(w web.ResponseWriter, r *web.Request) {
	hostname := r.PathParams["hid"]
	stats := []byte{}
	if hostname == "" {
		stats, _ = json.Marshal(tracker.Hostinfo)
	} else {
		stats, _ = json.Marshal(tracker.Cluster()[hostname])
	}
	fmt.Fprint(w, string(stats))
}

func handleHostList(w web.ResponseWriter, r *web.Request) {
	r.Header.Set("Content-Type", "application/json")
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

func handleContainerCreate(w web.ResponseWriter, r *web.Request) {
	poolName := r.PathParams["poolid"]
	containerName := r.PathParams["cid"]

	targetHost, err := getPreferedHost(poolName, containerName)
	if err != nil {
		log.Println("[error]", err)
		http.Error(w, fmt.Sprintf("%v", err), 502)
		return
	}

	ready := tracker.CreateIntent(targetHost, containerName, poolName)

	if ready != true {
		logMessage := "Not unique container name"
		log.Println(logMessage)
		http.Error(w, logMessage, 502)
		return
	}

	createMessage := makeMessage("container-create", containerName)

	peerAddr, _ := tracker.Config.GetString("cluster", "port")
	hostAnswer, _ := sendTcpMessage(targetHost+":"+peerAddr, createMessage)
	http.Error(w, hostAnswer, 201)
}

func handleHostContainerCreate(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Создать контейнер на этом хосте", 501)
}

func handleHostContainerUpdate(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Обновить настройки контейнера", 501)
}

func handleContainerDestroy(w web.ResponseWriter, r *web.Request) {
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

	peerAddr, _ := tracker.Config.GetString("cluster", "port")
	answer, _ := sendTcpMessage(hostname+":"+peerAddr, controlMessage)
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

func handleContainerStart(w web.ResponseWriter, r *web.Request) {
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

	peerAddr, _ := tracker.Config.GetString("cluster", "port")
	answer, _ := sendTcpMessage(hostname+":"+peerAddr, controlMessage)
	switch answer {
	case "done":
		http.Error(w, "", 204)
	case "not done":
		http.Error(w, "Not started", 502)
	}
}

func handleContainerStop(w web.ResponseWriter, r *web.Request) {
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

	peerAddr, _ := tracker.Config.GetString("cluster", "port")
	answer, _ := sendTcpMessage(hostname+":"+peerAddr, controlMessage)
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

func getPreferedHost(poolName string, containerName string) (string, error) {
	segments := libscore.Segments(tracker.Cluster(poolName))
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
