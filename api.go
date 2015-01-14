package main

import (
	"encoding/json"
	"html/template"

	"github.com/brnv/web"

	"io"
	"math/rand"
	"net/http"
)

type (
	Context struct{}
)

const (
	etcdErrCodeKeyExists = "105"
	apiPrefix            = "/v2"
)

var (
	rootRouter  = web.New(Context{})
	apiRouter   = rootRouter.Subrouter(Context{}, apiPrefix)
	webPort     string
	staticDir   string
	clusterPort string
)

type ApiAnswer struct {
	Status     string      `json:"status"`
	From       interface{} `json:"from"`
	Msg        interface{} `json:"msg"`
	Error      interface{} `json:"error"`
	LastUpdate interface{} `json:"lastupdate"`
}

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
		Get("/c/:cid/stats", handleContainerStats, "Container :cid statistics").
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
	j, _ := json.Marshal(apiRouter)
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func handleContainerStats(w web.ResponseWriter, r *web.Request) {
	containerName := r.PathParams["cid"]
	stats := []byte{}

	for _, host := range Cluster() {
		for _, container := range host.Containers {
			if container.Name == containerName {
				stats, _ = json.Marshal(container)
				break
			}
		}

		if len(stats) > 0 {
			break
		}
	}

	if len(stats) == 0 {
		w.WriteHeader(http.StatusNotFound)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(stats)
}

func handleStats(w web.ResponseWriter, r *web.Request) {
	hostname := r.PathParams["hid"]
	stats := []byte{}

	if hostname == "" {
		stats, _ = json.Marshal(Hostinfo)
	} else {
		stats, _ = json.Marshal(Cluster()[hostname])
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(stats)
}

func handleHostsList(w web.ResponseWriter, r *web.Request) {
	cluster := Cluster()

	if len(cluster) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
	} else if _, ok := cluster[Hostinfo.Hostname]; !ok {
		w.WriteHeader(http.StatusInternalServerError)
	}

	hostsList, _ := json.Marshal(cluster)

	w.Header().Set("Content-Type", "application/json")
	w.Write(hostsList)
}

func handleContainerCreate(w web.ResponseWriter, r *web.Request) {
	request := ContainerCreateRequest{}
	request.ContainerName = r.PathParams["cid"]
	request.PoolName = r.PathParams["poolid"]

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		if err != io.EOF {
			log.Error(err.Error())
		}
	}

	response := request.Execute()

	response.Send(w)
}

func handleContainerStart(w web.ResponseWriter, r *web.Request) {
	request := ContainerStartRequest{}
	request.ContainerName = r.PathParams["cid"]
	request.Host = r.PathParams["hid"]

	response := request.Execute()

	response.Send(w)
}

func handleContainerStop(w web.ResponseWriter, r *web.Request) {
	request := ContainerStopRequest{}
	request.ContainerName = r.PathParams["cid"]
	request.Host = r.PathParams["hid"]

	response := request.Execute()

	response.Send(w)
}

func handleContainerDestroy(w web.ResponseWriter, r *web.Request) {
	request := ContainerDestroyRequest{}
	request.ContainerName = r.PathParams["cid"]
	request.Host = r.PathParams["hid"]

	response := request.Execute()

	response.Send(w)
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
