package webserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os/exec"
	"regexp"
	"time"

	"heaverd-ng/libscore"
	"heaverd-ng/tracker"

	"github.com/gocraft/web"
)

var (
	hosts       = make(map[string]Statistics)
	hostRe      = regexp.MustCompile(`^/h/([A-Za-z0-9_-]*)[/]?([A-Za-z0-9_-]*)$`)
	containerRe = regexp.MustCompile(`^/c/([A-Za-z0-9_-]*)$`)
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
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprint(w, string(stats))
}

func handleHostListRequest(w web.ResponseWriter, r *web.Request) {
	_, err := json.Marshal(hosts)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	http.Error(w, "", 300)
}

func handleHostPing(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 501)
}

func handleHostCreateRequest(w web.ResponseWriter, r *web.Request) {
	hostname := hostRe.FindStringSubmatch(r.URL.Path)[1]

	rawdata, err := ioutil.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if string(rawdata) == "" {
		http.Error(w, "", 400)
		return
	}

	hostdata := Statistics{}
	json.Unmarshal(rawdata, &hostdata)
	hosts[hostname] = hostdata

	http.Error(w, "201 Created", 201)
	return
}

func handleHostInformationRequest(w web.ResponseWriter, r *web.Request) {
	hostname := hostRe.FindStringSubmatch(r.URL.Path)[1]

	hostinfo, err := json.Marshal(hosts[hostname])

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if string(hostinfo) == "null" {
		http.Error(w, "", 404)
		return
	}

	fmt.Fprint(w, hostinfo)
}

func handleHostOperationRequest(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 501)
}

func handleFindHostByContainerRequest(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 501)
}

func handleContainerCreateRequest(w web.ResponseWriter, r *web.Request) {
	containerName := containerRe.FindStringSubmatch(r.URL.Path)[1]
	intentId := rand.Intn(5000)
	intent := tracker.Intent{
		Id:            intentId,
		ContainerName: containerName,
		Creationtime:  time.Now().Unix(),
	}
	intentMessage := tracker.IntentMessage{
		tracker.MessageHeader{MessageType: "intent"},
		intent,
	}
	jsoned, err := json.Marshal(intentMessage)
	if err != nil {
		log.Println("[error]", err)
	}
	cmd := exec.Command("heaverd-tracker-query", "intent", fmt.Sprintf("%s", jsoned))
	cmd.Run()
	if err != nil {
		log.Fatal("[error]", err)
	}

	host, err := getPreferedHost(containerName)
	if err != nil {
		log.Fatal("[error]", err)
	}
	containerCreateMessage := tracker.ContainerCreateMessage{
		tracker.MessageHeader{MessageType: "container-create"},
		tracker.Intent{Id: intentId, ContainerName: containerName},
	}
	jsoned, err = json.Marshal(containerCreateMessage)
	if err != nil {
		log.Println("[error]", err)
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s%s", host, ":1444"))
	if err != nil {
		log.Fatal("[error]", err)
	}
	fmt.Fprintf(conn, fmt.Sprintf("%s", jsoned))
}

func getPreferedHost(containerName string) (string, error) {
	segments := libscore.Segments(tracker.GetCluster())
	host, err := libscore.ChooseHost(containerName, segments)
	if err != nil {
		return "", err
	}
	return host, nil
}

func Start(port string) {
	rand.Seed(time.Now().UnixNano())

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
		//	Get("/c/:cid", handleFindHostByContainerRequest).
		Get("/c/:cid", handleContainerCreateRequest)

	go http.ListenAndServe(port, router)
	log.Println("started at port", port)
	for {
		time.Sleep(time.Second)
	}
}
