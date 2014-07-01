package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

type Statistics struct {
	Alive bool `json:"alive"`
	Boxes map[string]struct {
		Active bool                `json:"active"`
		Ips    map[string][]string `json:"ips"`
	} `json:"boxes"`
	Fs        int                `json:"fs"`
	Ips_free  int                `json:"ips_free"`
	La        map[string]float32 `json:"la"`
	Last_seen int64              `json:"last_seen"`
	Now       float64            `json:"now"`
	Oom       []string           `json:"oom"`
	Ram       map[string]int     `json:"ram"`
	Score     float32            `json:"score"`
	Stale     bool               `json:"stale"`
}

var hosts = make(map[string]*Statistics)

func statsJson(w http.ResponseWriter, r *http.Request) {
	stats, _ := json.Marshal(hosts)
	fmt.Fprintf(w, "%s", stats)
}

func hostManagement(w http.ResponseWriter, r *http.Request) {
	var restQuery = regexp.MustCompile(`^/h/([A-Za-z0-9_-]*)[/]?([A-Za-z0-9_-]*)$`).
		FindStringSubmatch(r.URL.Path)
	hostname := restQuery[1]
	command := restQuery[2]

	if hostname == "" {
		http.Error(w, "300 Multiple Choices", 300)
		stats, _ := json.Marshal(hosts)
		fmt.Fprintf(w, "%s", stats)
		return
	}

	if command == "ping" && r.Method == "GET" {
		http.Error(w, "501 Not Implemented", 501)
		return
	} else if command != "" {
		http.Error(w, "404 Not Found", 404)
		return
	}

	if r.Method == "GET" {
		http.Error(w, "501 Not Implemented", 501)
		return
	}

	if r.Method == "PUT" {
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

	if r.Method == "POST" {
		http.Error(w, "501 Not Implemented", 501)
		return
	}
}

func containerManagement(w http.ResponseWriter, r *http.Request) {
	var container = regexp.MustCompile(`^/c/([A-Za-z0-9_-]*)$`).
		FindStringSubmatch(r.URL.Path)[1]

	fmt.Fprintf(w, "Container: %s", container)
}

func apiHelp(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Справка по API в запрошенном формате")
}

func main() {
	var listen = flag.String("listen", ":8081", "")
	flag.Parse()

	http.HandleFunc("/", apiHelp)
	http.HandleFunc("/stats/", statsJson)
	http.HandleFunc("/h/", hostManagement)
	http.HandleFunc("/c/", containerManagement)

	log.Println("[server] starting listener at", *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
