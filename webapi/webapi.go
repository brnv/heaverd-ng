package webapi

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"text/template"

	"heaverd-ng/libscore"
	"heaverd-ng/tracker"

	"github.com/brnv/web"
	"github.com/op/go-logging"
	"github.com/zazab/zhash"
)

type (
	Context   struct{}
	ApiAnswer struct {
		Status string `json:"status"`
		Msg    string `json:"msg"`
		Error  string `json:"error"`
	}
)

var (
	rootRouter = web.New(Context{})
	apiRouter  = rootRouter.Subrouter(Context{}, "/v2")
	config     = zhash.NewHash()
	log        = &logging.Logger{}
)

const etcdErrCodeKeyExists = "105"

func Run(configPath string, seed int64, logger *logging.Logger) {
	log = logger
	err := readConfig(configPath)
	if err != nil {
		log.Fatal(err.Error())
	}
	rand.Seed(seed)

	rootRouter.
		Middleware(web.StaticMiddleware("www")).
		Get("/score", handleScore, "Графики пулов хостов")

	apiRouter.
		Get("/", handleHelp, "Справка по API").
		Get("/h/:hid/stats", handleStats, "Статистика хоста :hid").
		Get("/h/", handleHostList, "Список всех хостов").
		Head("/c/:cid", handleHostByContainer, "Найти хост по контейнеру").
		Post("/c/:cid", handleContainerCreate, "Создать контейнер :cid (балансировка)").
		Post("/p/:poolid/:cid", handleContainerCreate, "Создать контейнер в пуле :poolid (балансировка)").
		Delete("/h/:hid/:cid", handleContainerDestroy, "Удалить контейнер").
		Delete("/c/:cid", handleContainerDestroy, "Удалить контейнер").
		Post("/h/:hid/:cid/start", handleContainerStart, "Стартануть контейнер").
		Post("/c/:cid/start", handleContainerStart, "Стартануть контейнер").
		Post("/h/:hid/:cid/stop", handleContainerStop, "Стоп контейнера").
		Post("/c/:cid/stop", handleContainerStop, "Стоп контейнера").
		Post("/h/:hid", handleHostOperation, "Операции с хостом").
		Get("/h/:hid", handleHostContainersList, "Список контейнеров на хосте").
		Get("/h/:hid/ping", handleHostPing, "Пинг сервера").
		Get("/c/", handleClusterContainersList, "Все контейнеры").
		Post("/h/:hid/:cid", handleHostContainerCreate, "Создать контейнер на хосте").
		Put("/h/:hid/:cid", handleHostContainerUpdate, "Обновить контейнер").
		Get("/h/:hid/:cid", handleContainerInfo, "Информация о контейнере").
		Get("/h/:hid/:cid/ping", handleContainerPing, "Пинг контейнера")

	port, _ := config.GetString("web", "port")
	log.Info("started at port: %s", port)
	log.Fatal(http.ListenAndServe(":"+port, rootRouter))
}

func handleScore(w web.ResponseWriter, r *web.Request) {
	r.Header.Set("Content-Type", "text/html")
	templates, _ := config.GetString("templates", "dir")
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
	w.Header().Set("Content-Type", "application/json")
	cluster, _ := json.Marshal(tracker.Cluster())
	fmt.Fprint(w, string(cluster))
}

func handleContainerCreate(w web.ResponseWriter, r *web.Request) {
	params := struct {
		Image         []string `json:"image"`
		Key           string   `json:"key"`
		ContainerName string
		PoolName      string
		TargetHost    string
	}{}

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		log.Error(err.Error())
	}

	r.ParseForm()
	containerName := r.PathParams["cid"]
	poolName := r.PathParams["poolid"]

	targetHost, err := getPreferedHost(containerName, poolName)
	if err != nil {
		log.Error(err.Error())
		apiAnswer(w, "error", "", fmt.Sprintf("%v", err), http.StatusNotFound)
		return
	}

	params.ContainerName = containerName
	params.PoolName = poolName
	params.TargetHost = targetHost

	err = tracker.CreateIntent(params.PoolName, params.ContainerName,
		params.TargetHost, params.Image, params.Key)
	if err != nil {
		if strings.Contains(err.Error(), etcdErrCodeKeyExists) {
			apiAnswer(w, "error", "", "Not unique name", http.StatusConflict)
		}
		return
	}

	createMessage := makeMessage("container-create", params.ContainerName)

	peerAddr, _ := config.GetString("cluster", "port")
	hostAnswer, _ := sendTcpMessage(targetHost+":"+peerAddr, createMessage)

	if strings.Contains(hostAnswer, "Error") {
		apiAnswer(w, "error", "", hostAnswer, http.StatusNotFound)
		return
	}

	apiAnswer(w, "ok", hostAnswer, "", http.StatusCreated)
}

func containerControl(action string, w web.ResponseWriter, r *web.Request) {
	var hostname string
	containerName := r.PathParams["cid"]

	if _, ok := r.PathParams["hid"]; ok {
		hostname = r.PathParams["hid"]
	} else {
		hostname = getHostnameByContainer(containerName)
		if hostname == "" {
			apiAnswer(w, "error", "", "unknown container", http.StatusNotFound)
			return
		}
	}

	if !checkHostname(hostname) {
		apiAnswer(w, "error", "", "unknown host", http.StatusNotFound)
		return
	}

	if !checkContainer(hostname, containerName) {
		apiAnswer(w, "error", "", "unknown container", http.StatusNotFound)
		return
	}

	controlMessage := makeMessage("container-control", struct {
		ContainerName string
		Action        string
	}{
		containerName,
		action,
	})

	peerAddr, _ := config.GetString("cluster", "port")
	answer, _ := sendTcpMessage(hostname+":"+peerAddr, controlMessage)
	switch answer {
	case "ok":
		apiAnswer(w, "ok", "", "", http.StatusNoContent)
	default:
		apiAnswer(w, "error", "", answer, http.StatusConflict)
	}
}

func apiAnswer(w web.ResponseWriter, status string, msg string, err string, code int) {
	w.Header().Set("Content-Type", "application/json")
	answer, _ := json.Marshal(ApiAnswer{
		Status: status,
		Msg:    msg,
		Error:  err,
	})
	w.WriteHeader(code)
	fmt.Fprint(w, string(answer))
}

func getPreferedHost(containerName string, poolName string) (string, error) {
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

func readConfig(path string) error {
	f, err := os.Open(path)
	if err == nil {
		config.ReadHash(bufio.NewReader(f))
		return nil
	} else {
		return err
	}
}

func handleHostContainerCreate(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Создать контейнер на этом хосте", 501)
}
func handleHostContainerUpdate(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Обновить настройки контейнера", 501)
}
func handleContainerStart(w web.ResponseWriter, r *web.Request) {
	containerControl("start", w, r)
}
func handleContainerStop(w web.ResponseWriter, r *web.Request) {
	containerControl("stop", w, r)
}
func handleContainerDestroy(w web.ResponseWriter, r *web.Request) {
	containerControl("destroy", w, r)
}
func handleContainerInfo(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Получить инфороцию о контейнере", 501)
}
func handleContainerPing(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Пингануть сервер", 501)
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
