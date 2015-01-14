package main

import (
	"encoding/json"
	"net/http"

	"github.com/brnv/heaverd-ng/liblxc"
	"github.com/brnv/web"
)

type Response interface {
	Send(w web.ResponseWriter)
}

type (
	BaseResponse struct {
		ResponseHost string `json:"from"`
		Status       string `json:"status"`
	}

	ContainerControlResponse struct {
		BaseResponse
		Token int64 `json:"lastupdate"`
	}

	ContainerCreatedResponse struct {
		BaseResponse
		Msg       string        `json:"msg"`
		Container lxc.Container `json:"container"`
	}
)

func ResponseSend(w web.ResponseWriter, response Response) {
	answer, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(answer)
}

func (response ContainerControlResponse) Send(w web.ResponseWriter) {
	response.Status = "ok"
	w.WriteHeader(http.StatusOK)
	ResponseSend(w, response)
}

func (response ContainerCreatedResponse) Send(w web.ResponseWriter) {
	response.Status = "ok"
	response.Msg = "Container created"
	w.WriteHeader(http.StatusCreated)
	ResponseSend(w, response)
}
