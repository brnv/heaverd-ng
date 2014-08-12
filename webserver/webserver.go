package webserver

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
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
	cluster, _ := json.Marshal(tracker.Cluster())
	fmt.Fprint(w, string(cluster))
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

	intentMessage, err := json.Marshal(struct {
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

	for _, host := range tracker.Cluster() {
		nodeConnection, err := net.Dial("tcp",
			fmt.Sprintf("%s%s", host.Hostname, ":1444"))
		defer nodeConnection.Close()
		if err != nil {
			log.Println("[error]", err)
		}
		fmt.Fprint(nodeConnection, fmt.Sprintf("%s", intentMessage))

		nodeAnswer, err := bufio.NewReader(nodeConnection).ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Println("[error]", err)
			}
		}
		if string(nodeAnswer) == "fail" {
			http.Error(w, "Not unique container name", 409)
			return
		}
	}

	targetHost, err := getPreferedHost(containerName)
	if err != nil {
		log.Fatal("[error]", err)
	}

	createMessage, err := json.Marshal(struct {
		Type string
		Body interface{}
	}{
		"container-create",
		tracker.Intent{Id: intentId, ContainerName: containerName},
	})

	hostConnection, err := net.Dial("tcp",
		fmt.Sprintf("%s%s", targetHost, ":1444"))
	if err != nil {
		log.Fatal("[error]", err)
	}
	fmt.Fprint(hostConnection, fmt.Sprintf("%s", createMessage))

	hostAnswer, err := bufio.NewReader(hostConnection).ReadString('\n')
	if err != nil {
		if err != io.EOF {
			log.Println("[error]", err)
		}
	}
	http.Error(w, string(hostAnswer), 201)
}

func getPreferedHost(containerName string) (string, error) {
	segments := libscore.Segments(tracker.Cluster())
	host, err := libscore.ChooseHost(containerName, segments)
	if err != nil {
		return "", err
	}
	return host, nil
}
