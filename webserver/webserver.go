package webserver

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"heaverd-ng/libscore"
	"heaverd-ng/tracker"

	"github.com/gocraft/web"
)

type Context struct{}

func handleHelpRequest(w web.ResponseWriter, r *web.Request) {
	fmt.Fprintf(w, "Справка по API в запрошенном формате")
}

func handleStatisticsRequest(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 501)
}

func handleHostListRequest(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 501)
}

func handleHostPing(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 501)
}

func handleHostCreateRequest(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 501)
}

func handleHostInformationRequest(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 501)
}

func handleHostOperationRequest(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 501)
}

func handleFindHostByContainerRequest(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 501)
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

	log.Println("started at port", port)
	log.Fatal(http.ListenAndServe(port, router))
}

func handleContainerCreateRequest(w web.ResponseWriter, r *web.Request) {
	containerName := r.PathParams["cid"]

	intentId := rand.Intn(5000)

	message, err := json.Marshal(struct {
		Type string
		Body interface{}
	}{
		"container-create-intent",
		tracker.Intent{
			Id:            intentId,
			ContainerName: containerName,
			CreatedAt:     time.Now().Unix(),
		},
	})

	if err != nil {
		log.Fatal("[error]", err)
	}

	cluster := tracker.Cluster()
	for _, host := range cluster {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s%s", host.Hostname, ":1444"))
		if err != nil {
			log.Fatal("[error]", err)
		}

		fmt.Fprintf(conn, fmt.Sprintf("%s", message))

		// FIXME find a better way to do that
		answer := make([]byte, 10)
		conn.Read(answer)
		if strings.Contains(string(answer), "fail") {
			http.Error(w, "Not unique container name", 400)
			return
		}
	}

	targetHost, err := getPreferedHost(containerName)
	if err != nil {
		log.Fatal("[error]", err)
	}

	containerCreateMessage, err := json.Marshal(struct {
		Type string
		Body interface{}
	}{
		"container-create",
		tracker.Intent{Id: intentId, ContainerName: containerName},
	})

	conn, err := net.Dial("tcp", fmt.Sprintf("%s%s", targetHost, ":1444"))
	if err != nil {
		log.Fatal("[error]", err)
	}
	fmt.Fprintf(conn, fmt.Sprintf("%s", containerCreateMessage))
}

func getPreferedHost(containerName string) (string, error) {
	segments := libscore.Segments(tracker.Cluster())
	host, err := libscore.ChooseHost(containerName, segments)
	if err != nil {
		return "", err
	}
	return host, nil
}
