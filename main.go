package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
)

func main() {
	listen := flag.String("listen", ":8081", "")

	flag.Parse()

	http.HandleFunc("/stats/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "501 Not Implemented", 501)
	})

	http.HandleFunc("/c/", func(w http.ResponseWriter, r *http.Request) {
		var container = regexp.MustCompile(`^/c/([A-Za-z0-9_-]*)$`).
			FindStringSubmatch(r.URL.Path)[1]

		fmt.Fprintf(w, "Container: %s", container)
	})

	log.Println("[server] starting listener at", *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}
