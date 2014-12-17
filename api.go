package main

import (
	"encoding/json"
	"html/template"

	"github.com/brnv/go-lxc"
	"github.com/brnv/heaverd-ng/libscore"
	"github.com/brnv/web"

	"bufio"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strings"
)

type (
	Context   struct{}
	ApiAnswer struct {
		Status     string      `json:"status"`
		From       interface{} `json:"from"`
		Msg        interface{} `json:"msg"`
		Error      interface{} `json:"error"`
		LastUpdate interface{} `json:"lastupdate"`
	}
)

const (
	etcdErrCodeKeyExists = "105"
	apiVersion           = "v2"
)

var (
	rootRouter  = web.New(Context{})
	apiRouter   = rootRouter.Subrouter(Context{}, "/"+apiVersion)
	webPort     string
	staticDir   string
	clusterPort string
)

func WebapiRun(params map[string]interface{}) {
	webPort = params["webPort"].(string)
	staticDir = params["staticDir"].(string)
	clusterPort = params["clusterPort"].(string)
	rand.Seed(params["seed"].(int64))

	rootRouter.
		Middleware(web.StaticMiddleware(staticDir)).
		Get("/score", handleScore, "Cluster graphs")

	apiRouter.
		Get("/", handleHelp, "Show this help").
		Get("/h/:hid/stats", handleStats, "Host :his resources statistics").
		Get("/h/", handleHostsList, "Hosts list").
		Post("/c/:cid", handleContainerCreate, "Create container :cid using balancer").
		Post("/p/:poolid/:cid", handleContainerCreate,
		"Create container :cid inside :poolid pool using balancer").
		Post("/h/:hid/:cid/start", handleContainerStart, "Start container").
		Post("/c/:cid/start", handleContainerStart, "Start container").
		Post("/h/:hid/:cid/stop", handleContainerStop, "Terminate container").
		Post("/c/:cid/stop", handleContainerStop, "Terminate container").
		Delete("/h/:hid/:cid", handleContainerDestroy, "Destroy container").
		Delete("/c/:cid", handleContainerDestroy, "Destroy container").
		Post("/h/:hid/:cid", handleHostContainerCreate, "Create container :cid on host :hid (tbd)").
		Put("/h/:hid/:cid", handleHostContainerUpdate, "Update container :cid settings (tbd)").
		Get("/h/:hid/:cid", handleContainerInfo, "Container :cid infromation (tbd)").
		Get("/h/:hid/:cid/ping", handleContainerPing, "Ping container (tbd)").
		Post("/h/:hid", handleHostOperation, "Host operation (tdb)").
		Get("/h/:hid", handleHostContainersList, "Containers list on host :hid (tbd)").
		Head("/c/:cid", handleHostByContainer, "Get host by conatiner name (tbd)")

	log.Info("web api is on :%s", webPort)
	log.Fatal(http.ListenAndServe(":"+webPort, rootRouter))
}

func handleScore(w web.ResponseWriter, r *web.Request) {
	w.Header().Set("Content-Type", "text/html")
	template.Must(template.
		ParseFiles(staticDir+"templates/index.tpl")).Execute(w, nil)
}

func handleHelp(w web.ResponseWriter, r *web.Request) {
	w.Header().Set("Content-Type", "application/json")
	j, _ := json.Marshal(apiRouter)
	w.Write(j)
}

func handleStats(w web.ResponseWriter, r *web.Request) {
	hostname := r.PathParams["hid"]
	stats := []byte{}
	if hostname == "" {
		stats, _ = json.Marshal(Hostinfo)
	} else {
		stats, _ = json.Marshal(Cluster()[hostname])
	}
	w.Write(stats)
}

func handleHostsList(w web.ResponseWriter, r *web.Request) {
	w.Header().Set("Content-Type", "application/json")
	cluster, _ := json.Marshal(Cluster())
	w.Write(cluster)
}

func handleContainerCreate(w web.ResponseWriter, r *web.Request) {
	intent := Intent{}
	err := json.NewDecoder(r.Body).Decode(&intent)
	if err != nil {
		if err != io.EOF {
			log.Error(err.Error())
		}
	}

	r.ParseForm()
	containerName := r.PathParams["cid"]
	poolName := r.PathParams["poolid"]

	targetHost, err := getPreferedHost(containerName, poolName)
	if err != nil {
		log.Error(err.Error())
		apiAnswer(w, "error", nil, http.StatusNotFound, nil, err.Error(), nil)
		return
	}

	intent.ContainerName = containerName
	intent.PoolName = poolName
	intent.TargetHost = targetHost

	err = CreateIntent(intent)
	if err != nil {
		if strings.Contains(err.Error(), etcdErrCodeKeyExists) {
			apiAnswer(w, "error", targetHost, http.StatusConflict, nil, "Not unique name", nil)
		}
		return
	}

	createMessage := makeMessage("container-create", intent.ContainerName)

	hostAnswer, _ := sendTcpMessage(targetHost, clusterPort, createMessage)

	if strings.Contains(string(hostAnswer), "Error") {
		apiAnswer(w, "error", targetHost, http.StatusNotFound, nil, string(hostAnswer), nil)
		return
	}

	newContainer := lxc.Container{}
	json.Unmarshal(hostAnswer, &newContainer)

	apiAnswer(w, "ok", targetHost, http.StatusCreated, newContainer, "", nil)
}

func handleContainerStart(w web.ResponseWriter, r *web.Request) {
	controlContainer("start", w, r)
}

func handleContainerStop(w web.ResponseWriter, r *web.Request) {
	controlContainer("stop", w, r)
}

func handleContainerDestroy(w web.ResponseWriter, r *web.Request) {
	controlContainer("destroy", w, r)
}

func controlContainer(action string, w web.ResponseWriter, r *web.Request) {
	var hostname interface{}
	containerName := r.PathParams["cid"]

	if _, ok := r.PathParams["hid"]; ok {
		hostname = r.PathParams["hid"]
	} else {
		hostname = getHostnameByContainer(containerName)
		if hostname == nil {
			apiAnswer(w, "error", hostname, http.StatusNotFound, nil, "Unknown container", nil)
			return
		}
	}

	if !isHostExists(hostname) {
		apiAnswer(w, "error", nil, http.StatusNotFound, nil, "Unknown host", nil)
		return
	}

	if !isContainerExists(hostname, containerName) {
		apiAnswer(w, "error", hostname, http.StatusNotFound, nil, "Unknown container", nil)
		return
	}

	controlMessage := makeMessage("container-control", struct {
		ContainerName string
		Action        string
	}{
		containerName,
		action,
	})

	rawAnswer, _ := sendTcpMessage(hostname, clusterPort, controlMessage)
	answer := struct {
		From       string
		Code       int
		Text       string
		Error      string
		LastUpdate int64
	}{}

	err := json.Unmarshal(rawAnswer, &answer)
	if err != nil {
		apiAnswer(w, "error", nil, http.StatusInternalServerError, nil, err.Error(), nil)
		return
	}

	status := "ok"
	if answer.Code != 200 {
		status = "error"
	}

	apiAnswer(w, status, answer.From, answer.Code, answer.Text, answer.Error, answer.LastUpdate)
}

func apiAnswer(w web.ResponseWriter, status string, from interface{}, code int,
	msg interface{}, err interface{}, lastUpdate interface{}) {

	w.Header().Set("Content-Type", "application/json")
	answer, _ := json.Marshal(ApiAnswer{
		Status:     status,
		From:       from,
		Msg:        msg,
		Error:      err,
		LastUpdate: lastUpdate,
	})
	w.WriteHeader(code)
	w.Write(answer)
}

func getPreferedHost(containerName string, poolName string) (string, error) {
	segments := libscore.Segments(Cluster(poolName))
	host, err := libscore.ChooseHost(containerName, segments)
	if err != nil {
		return "", err
	}
	return host, nil
}

func sendTcpMessage(host interface{}, port string, message []byte) ([]byte, error) {
	connection, err := net.Dial("tcp", host.(string)+":"+port)
	if err != nil {
		return []byte{}, err
	}
	defer connection.Close()

	connection.Write(message)

	answer, err := bufio.NewReader(connection).ReadString('\n')
	if err != nil {
		if err != io.EOF {
			return []byte{}, err
		}
	}

	return []byte(answer), nil
}

func isHostExists(name interface{}) bool {
	if _, ok := Cluster()[name.(string)]; !ok {
		return false
	}
	return true
}

func isContainerExists(hostname interface{}, name string) bool {
	if _, ok := Cluster()[hostname.(string)].Containers[name]; !ok {
		return false
	}
	return true
}

func getHostnameByContainer(containerName string) interface{} {
	for hostname, host := range Cluster() {
		if _, ok := host.Containers[containerName]; ok {
			return hostname
		}
	}
	return nil
}

func makeMessage(header string, body interface{}) []byte {
	message, _ := json.Marshal(struct {
		Type string
		Body interface{}
	}{
		header,
		body,
	})
	return message
}

func handleHostContainerCreate(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "handleHostContainerCreate", 501)
}
func handleHostContainerUpdate(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "handleHostContainerUpdate", 501)
}
func handleContainerInfo(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "handleContainerInfo", 501)
}
func handleContainerPing(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "handleContainerPing", 501)
}
func handleHostOperation(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "handleHostOperation", 501)
}
func handleHostContainersList(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "handleHostContainersList", 501)
}
func handleHostByContainer(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "handleHostByContainer", 501)
}
