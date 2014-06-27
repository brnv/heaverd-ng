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
	Last_seen string
	//Now string //date
	//	Oom // [ ]
	Ram   map[string]int
	Score float32
	Stale bool
}

func main() {
	listen := flag.String("listen", ":8081", "")

	flag.Parse()

	http.HandleFunc("/stats/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "501 Not Implemented", 501)

		hosts := make(map[string]*Statistics)
		hosts["lxbox"] = &Statistics{}

		hosts["lxbox"].Alive = true

		//Boxes := make(map[string]BoxesType)
		json, _ := json.Marshal(hosts)
		fmt.Fprintf(w, "%s", json)
	})

	http.HandleFunc("/c/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "501 Not Implemented", 501)

		var container = regexp.MustCompile(`^/c/([A-Za-z0-9_-]*)$`).
			FindStringSubmatch(r.URL.Path)[1]

		fmt.Fprintf(w, "Container: %s", container)
	})

	log.Println("[server] starting listener at", *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
