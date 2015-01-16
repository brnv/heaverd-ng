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

	ContainerControlResponse struct {
		BaseResponse
		Token int64 `json:"token"`
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
		Token int64 `json:"token"`
	}

	ContainerStopResponse struct {
		BaseResponse
		Token int64 `json:"token"`
	}

	ContainerDestroyResponse struct {
		BaseResponse
		Token int64 `json:"token"`
	}
)

func ResponseSend(w web.ResponseWriter, response Response) {
	answer, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(answer)
}

func (response ContainerControlResponse) Write(w web.ResponseWriter) {
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

func (response ContainerPushResponse) Write(w web.ResponseWriter) {
	response.Status = "ok"
	w.WriteHeader(http.StatusOK)
	ResponseSend(w, response)
}
