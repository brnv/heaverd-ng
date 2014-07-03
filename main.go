package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocraft/web"
	"io/ioutil"
	"net/http"
	"regexp"
)

const (
	HOST_RE      = `^/h/([A-Za-z0-9_-]*)[/]?([A-Za-z0-9_-]*)$`
	CONTAINER_RE = `^/c/([A-Za-z0-9_-]*)$`
)

var (
	hosts = make(map[string]*Statistics)
)

type Statistics struct {
	Alive bool `json:"alive"`
	Boxes map[string]struct {
		Active bool                `json:"active"`
		Ips    map[string][]string `json:"ips"`
	} `json:"boxes"`
	Fs       int                `json:"fs"`
	IpsFree  int                `json:"ips_free"`
	La       map[string]float32 `json:"la"`
	LastSeen int64              `json:"last_seen"`
	Now      float64            `json:"now"`
	Oom      []string           `json:"oom"`
	Ram      map[string]int     `json:"ram"`
	Score    float32            `json:"score"`
	Stale    bool               `json:"stale"`
}

type Context struct{}

func handleHelpRequest(w web.ResponseWriter, r *web.Request) {
	fmt.Fprintf(w, "Справка по API в запрошенном формате")
}

func handleStatisticsRequest(w web.ResponseWriter, r *web.Request) {
	stats, _ := json.Marshal(hosts)
	fmt.Fprintf(w, "%s", stats)
}

func handleHostListRequest(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 300)
	stats, _ := json.Marshal(hosts)
	fmt.Fprintf(w, "%s", stats)
}

func handleHostPing(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 501)
}

func handleHostCreateRequest(w web.ResponseWriter, r *web.Request) {
	var restQuery = regexp.MustCompile(HOST_RE).FindStringSubmatch(r.URL.Path)
	hostname := restQuery[1]

	rawdata, err := ioutil.ReadAll(r.Body)

	//TODO if rawdata is empty, send 400 code

	if err != nil {
		http.Error(w, "Error while reading request body", 500)
		fmt.Fprintf(w, "%s", err)
		return
	}

	var hostdata = Statistics{}
	json.Unmarshal(rawdata, &hostdata)
	hosts[hostname] = &hostdata

	http.Error(w, "201 Created", 201)
	return
}

func handleHostInformationRequest(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 501)
}

func handleHostOperationRequest(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 501)
}

func handleFindContainerHostRequest(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 501)
}

func handleContainerCreateRequest(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 501)
}

func main() {
	router := web.New(Context{}).
		Middleware(web.LoggerMiddleware).
		Middleware(web.ShowErrorsMiddleware).
		Get("/", handleHelpRequest).
		Get("/stats/", handleStatisticsRequest).
		Get("/h/", handleHostListRequest).
		Get("/h/:hid/ping", handleHostPing).
		Put("/h/:hid", handleHostCreateRequest).
		Get("/h/:hid", handleHostInformationRequest).
		Post("/h/:hid", handleHostOperationRequest).
		Get("/c/:cid", handleFindContainerHostRequest).
		Post("/c/:cid", handleContainerCreateRequest)

	http.ListenAndServe("192.168.4.84:8081", router)
}
