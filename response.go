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

	ErrorResponse struct {
		BaseResponse
		Error string `json:"error"`
	}

	CantAssignAnyHostResponse struct {
		ErrorResponse
	}

	StaleDataResponse struct {
		ErrorResponse
	}

	NotUniqueNameResponse struct {
		ErrorResponse
	}

	ContainerCreatedResponse struct {
		BaseResponse
		Msg       string        `json:"msg"`
		Container lxc.Container `json:"container"`
	}
)

func (response CantAssignAnyHostResponse) Send(w web.ResponseWriter) {
	response.Status = "error"
	response.Error = "No suitable host found"
	w.WriteHeader(http.StatusNotFound)
	Send(w, response)
}

func (response StaleDataResponse) Send(w web.ResponseWriter) {
	response.Status = "error"
	response.Error = "Stale data"
	w.WriteHeader(http.StatusTeapot)
	Send(w, response)
}

func (response NotUniqueNameResponse) Send(w web.ResponseWriter) {
	response.Status = "error"
	response.Error = "Not unique name"
	w.WriteHeader(http.StatusConflict)
	Send(w, response)
}

func (response ContainerCreatedResponse) Send(w web.ResponseWriter) {
	response.Status = "ok"
	response.Msg = "Container created"
	w.WriteHeader(http.StatusCreated)
	Send(w, response)
}

func (response ErrorResponse) Send(w web.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
}

func (response BaseResponse) Send(w web.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
}

func Send(w web.ResponseWriter, response Response) {
	answer, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(answer)
}
