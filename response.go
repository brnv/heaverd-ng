package main

import (
	"encoding/json"
	"fmt"
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
	}

	ContainerControlResponse struct {
		BaseResponse
		Token int64 `json:"token"`
	}

	ContainerCreatedResponse struct {
		BaseResponse
		Msg       string        `json:"msg"`
		Container lxc.Container `json:"container"`
	}

	ContainerPushResponse struct {
		BaseResponse
		Image string
		Msg   string `json:"msg"`
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

func (response ContainerCreatedResponse) Write(w web.ResponseWriter) {
	response.Status = "ok"
	response.Msg = "Container created"
	w.WriteHeader(http.StatusCreated)
	ResponseSend(w, response)
}

func (response ContainerPushResponse) Write(w web.ResponseWriter) {
	response.Status = "ok"
	response.Msg = fmt.Sprintf(
		"Container's rootfs pushed into %s image", response.Image,
	)
	w.WriteHeader(http.StatusOK)
	ResponseSend(w, response)
}
