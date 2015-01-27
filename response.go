package main

import (
	"encoding/json"
	"net/http"

	"github.com/brnv/heaverd-ng/liblxc"
	"github.com/brnv/web"
)

type Response interface {
	Write(w web.ResponseWriter)
}

type (
	BaseResponse struct {
		ResponseHost string `json:"from"`
		Status       string `json:"status"`
		Error        string `json:"error"`
	}

	ClusterResponse struct {
		BaseResponse
		Token     int64 `json:"token"`
		Container lxc.Container
	}

	ContainerCreateResponse struct {
		BaseResponse
		Container lxc.Container `json:"container"`
	}

	ContainerPushResponse struct {
		BaseResponse
	}

	ContainerStartResponse struct {
		BaseResponse
	}

	ContainerStopResponse struct {
		BaseResponse
	}

	ContainerDestroyResponse struct {
		BaseResponse
		Token int64 `json:"token"`
	}

	ContainerRecreateResponse struct {
		BaseResponse
	}
)

func ResponseSend(w web.ResponseWriter, response Response) {
	answer, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(answer)
}

func (response ClusterResponse) Write(w web.ResponseWriter) {
	response.Status = "ok"
	w.WriteHeader(http.StatusOK)
	ResponseSend(w, response)
}

func (response ContainerStartResponse) Write(w web.ResponseWriter) {
	response.Status = "ok"
	w.WriteHeader(http.StatusOK)
	ResponseSend(w, response)
}

func (response ContainerStopResponse) Write(w web.ResponseWriter) {
	response.Status = "ok"
	w.WriteHeader(http.StatusOK)
	ResponseSend(w, response)
}

func (response ContainerDestroyResponse) Write(w web.ResponseWriter) {
	response.Status = "ok"
	w.WriteHeader(http.StatusOK)
	ResponseSend(w, response)
}

func (response ContainerCreateResponse) Write(w web.ResponseWriter) {
	response.Status = "ok"
	w.WriteHeader(http.StatusCreated)
	ResponseSend(w, response)
}

func (response ContainerRecreateResponse) Write(w web.ResponseWriter) {
	response.Status = "ok"
	w.WriteHeader(http.StatusOK)
	ResponseSend(w, response)
}

func (response ContainerPushResponse) Write(w web.ResponseWriter) {
	response.Status = "ok"
	w.WriteHeader(http.StatusOK)
	ResponseSend(w, response)
}
