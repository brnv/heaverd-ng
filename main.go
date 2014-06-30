package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
)

type BoxesType struct {
	Active bool
	Ips    map[string][]string
}

type Statistics struct {
	Alive     bool
	Boxes     map[string]BoxesType
	Fs        int
	Ips_free  int
	La        map[string]float32
	Last_seen int64
	Now       int64
	Oom       []string
	Ram       map[string]int
	Score     float32
	Stale     bool
}

var hosts = make(map[string]*Statistics)

func statsJson(w http.ResponseWriter, r *http.Request) {
	json, _ := json.Marshal(hosts)
	fmt.Fprintf(w, "%s", json)
}

func hostManagement(w http.ResponseWriter, r *http.Request) {
	var restQuery = regexp.MustCompile(`^/h/([A-Za-z0-9_-]*)[/]?([A-Za-z0-9_-]*)$`).
		FindStringSubmatch(r.URL.Path)
	host := restQuery[1]
	command := restQuery[2]

	if host == "" {
		http.Error(w, "300 Multiple Choices", 300)
		json, _ := json.Marshal(hosts)
		fmt.Fprintf(w, "%s", json)
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
		http.Error(w, "501 Not Implemented", 501)
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
