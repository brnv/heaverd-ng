package webserver

import (
	"bufio"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"heaverd-ng/libscore"
	"heaverd-ng/tracker"

	"github.com/gocraft/web"
)

type Context struct {
	clusterListen string
}

func Start(port string, clusterListen string, seed int64) {
	rand.Seed(seed)

	context := &Context{
		clusterListen: clusterListen,
	}

	router := web.New(Context{}).
		Middleware(web.LoggerMiddleware).
		Middleware(web.ShowErrorsMiddleware).

		//http://confluence.rn/display/ENV/RESTful+API
		//Common
		Get("/", handleHelp).
		Get("/stats/", handleStats).
		//Hosts
		Get("/h/", handleHostList).
		Post("/h/:hid", handleHostOperation).
		Get("/h/:hid", handleHostContainersList).
		Get("/h/:hid/ping", handleHostPing).
		//Containers
		Get("/c/", handleClusterContainersList).
		Get("/c/:cid", handleHostByContainer).
		Post("/c/:cid", context.handleContainerCreate).
		Post("/h/:hid/:cid", handleHostContainerCreate).
		Put("/h/:hid/:cid", handleHostContainerUpdate).
		Delete("/h/:hid/:cid", handleContainerDelete).
		Get("/h/:hid/:cid", handleContainerInfo).
		Post("/h/:hid/:cid/start", handleContainerStart).
		Post("/h/:hid/:cid/stop", handleContainerStop).
		//Post("/h/:hid/:cid/freeze", )
		//Post("/h/:hid/:cid/unfreeze", )
		//Get("/h/:hid/:cid/tarball", )
		Get("/h/:hid/:cid/ping", handleContainerPing)
	//Get("/h/:hid/:cid/attach", )

	log.Println("started at port :", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func handleHelp(w web.ResponseWriter, r *web.Request) {
	fmt.Fprintf(w, "Справка по API в запрошенном формате")
}

func handleStats(w web.ResponseWriter, r *web.Request) {
	cluster, _ := json.Marshal(tracker.Cluster())
	fmt.Fprint(w, string(cluster))
}

func handleHostList(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 501)
}

func handleHostOperation(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 501)
}

func handleHostContainersList(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "", 501)
}

func handleHostPing(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "pong", 204)
}

func handleClusterContainersList(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Список всех контейнеров на всех хостах", 501)
}

func handleHostByContainer(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Имя хоста, на котором расположен указанный контейнер", 501)
}

func handleHostInformationRequest(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Полная информация о хосте", 501)
}

func (c *Context) handleContainerCreate(w web.ResponseWriter, r *web.Request) {
	containerName := r.PathParams["cid"]
	h := md5.New()
	fmt.Fprint(h, containerName+strconv.FormatInt(time.Now().Unix(), 10))
	intentId := fmt.Sprintf("%x", h.Sum(nil))

	intentMessage, _ := json.Marshal(struct {
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

	for _, host := range tracker.Cluster() {
		log.Println(host.Hostname + ":" + c.clusterListen)
		nodeConnection, err := net.Dial("tcp", host.Hostname+":"+c.clusterListen)
		defer nodeConnection.Close()
		if err != nil {
			log.Println("[error]", err)
		}
		fmt.Fprint(nodeConnection, string(intentMessage))

		nodeAnswer, err := bufio.NewReader(nodeConnection).ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Println("[error]", err)
			}
		}
		if string(nodeAnswer) == "not_unique_name" {
			http.Error(w, "Not unique container name", 409)
			return
		}
	}

	targetHost, err := getPreferedHost(containerName)
	if err != nil {
		log.Println("[error]", err)
		http.Error(w, fmt.Sprintf("%v", err), 502)
		return
	}

	createMessage, err := json.Marshal(struct {
		Type string
		Body interface{}
	}{
		"container-create",
		tracker.Intent{Id: intentId, ContainerName: containerName},
	})

	hostConnection, err := net.Dial("tcp", targetHost+":"+c.clusterListen)
	if err != nil {
		log.Fatal("[error]", err)
	}
	fmt.Fprint(hostConnection, string(createMessage))

	hostAnswer, err := bufio.NewReader(hostConnection).ReadString('\n')
	if err != nil {
		if err != io.EOF {
			log.Println("[error]", err)
		}
	}
	http.Error(w, string(hostAnswer), 201)
}

func handleHostContainerCreate(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Создать контейнер на этом хосте", 501)
}

func handleHostContainerUpdate(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Обновить настройки контейнера", 501)
}

func handleContainerDelete(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Удалить контейнер", 501)
}

func handleContainerInfo(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Получить инфороцию о контейнере", 501)
}

func handleContainerStart(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Стартануть контейнер", 501)
}

func handleContainerStop(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Стопнуть контейнер", 501)
}

func handleContainerPing(w web.ResponseWriter, r *web.Request) {
	http.Error(w, "Пингануть сервер", 501)
}

func getPreferedHost(containerName string) (string, error) {
	segments := libscore.Segments(tracker.Cluster())
	host, err := libscore.ChooseHost(containerName, segments)
	if err != nil {
		return "", err
	}
	return host, nil
}
