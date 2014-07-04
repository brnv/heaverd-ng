package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocraft/web"
	"io/ioutil"
	"net/http"
	"regexp"
)

var (
	hosts        = make(map[string]*Statistics)
	HOST_RE      = regexp.MustCompile(`^/h/([A-Za-z0-9_-]*)[/]?([A-Za-z0-9_-]*)$`)
	CONTAINER_RE = regexp.MustCompile(`^/c/([A-Za-z0-9_-]*)$`)
)

type Statistics struct {
	Alive    bool               `json:"alive"`
	Fs       int                `json:"fs"`
	IpsFree  int                `json:"ips_free"`
	La       map[string]float32 `json:"la"`
	LastSeen int64              `json:"last_seen"`
	Now      float64            `json:"now"`
	Oom      []string           `json:"oom"`
	Ram      map[string]int     `json:"ram"`
	Score    float32            `json:"score"`
	Stale    bool               `json:"stale"`
	Boxes    map[string]struct {
		Active bool                `json:"active"`
		Ips    map[string][]string `json:"ips"`
	} `json:"boxes"`
}

type Context struct{}

func handleHelpRequest(w web.ResponseWriter, r *web.Request) {
	fmt.Fprintf(w, "Справка по API в запрошенном формате")
}

func handleStatisticsRequest(w web.ResponseWriter, r *web.Request) {
	stats, err := json.Marshal(hosts)

	if err != nil {
		http.Error(w, "", 500)
		fmt.Fprintf(w, "%s", err)
		return
	}

	fmt.Fprintf(w, "%s", stats)
}

func handleHostListRequest(w web.ResponseWriter, r *web.Request) {
	stats, err := json.Marshal(hosts)

	if err != nil {
		http.Error(w, "", 500)
		fmt.Fprintf(w, "%s", err)
		return
	}

	fmt.Fprintf(w, "%s", stats)
	http.Error(w, "", 300)
}

func handleHostPing(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 501)
}

func handleHostCreateRequest(w web.ResponseWriter, r *web.Request) {
	hostname := HOST_RE.FindStringSubmatch(r.URL.Path)[1]

	rawdata, err := ioutil.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Error while reading request body", 500)
		fmt.Fprintf(w, "%s", err)
		return
	}

	if string(rawdata) == "" {
		http.Error(w, "", 400)
		return
	}

	var hostdata = Statistics{}
	json.Unmarshal(rawdata, &hostdata)
	hosts[hostname] = &hostdata

	http.Error(w, "201 Created", 201)
	return
}

func handleHostInformationRequest(w web.ResponseWriter, r *web.Request) {
	hostname := HOST_RE.FindStringSubmatch(r.URL.Path)[1]

	hostinfo, err := json.Marshal(hosts[hostname])

	if err != nil {
		http.Error(w, "", 500)
		fmt.Fprintf(w, "%s", err)
		return
	}

	if string(hostinfo) == "null" {
		http.Error(w, "", 404)
		return
	}

	fmt.Fprintf(w, "%s", hostinfo)
}

func handleHostOperationRequest(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 501)
}

func handleFindHostByContainerRequest(w web.ResponseWriter, r *web.Request) {
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
		Get("/c/:cid", handleFindHostByContainerRequest).
		Post("/c/:cid", handleContainerCreateRequest)

	http.ListenAndServe("192.168.4.84:8081", router)
}
