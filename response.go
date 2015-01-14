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

	ErrorResponse struct {
		BaseResponse
		Error string `json:"error"`
	}

	ServerErrorResponse struct {
		ErrorResponse
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

	ContainerCreationErrorResponse struct {
		ErrorResponse
	}

	ContainerControlErrorResponse struct {
		ErrorResponse
	}

	ContainerCreatedResponse struct {
		BaseResponse
		Msg       string        `json:"msg"`
		Container lxc.Container `json:"container"`
	}

	CantFindContainerHostnameResponse struct {
		ErrorResponse
	}

	ContainerNotFoundResponse struct {
		ErrorResponse
	}

	HostNotFoundResponse struct {
		ErrorResponse
	}

	HeaverErrorResponse struct {
		ErrorResponse
	}
)

func (response ContainerControlResponse) Send(w web.ResponseWriter) {
	response.Status = "ok"
	w.WriteHeader(http.StatusOK)
	Send(w, response)
}

func (response CantFindContainerHostnameResponse) Send(w web.ResponseWriter) {
	response.Status = "error"
	response.Error = "Can't find host by given container"
	w.WriteHeader(http.StatusNotFound)
	Send(w, response)
}

func (response HeaverErrorResponse) Send(w web.ResponseWriter) {
	response.Status = "error"
	w.WriteHeader(http.StatusConflict)
	Send(w, response)
}

func (response HostNotFoundResponse) Send(w web.ResponseWriter) {
	response.Status = "error"
	response.Error = "Host not found"
	w.WriteHeader(http.StatusNotFound)
	Send(w, response)
}

func (response ContainerNotFoundResponse) Send(w web.ResponseWriter) {
	response.Status = "error"
	response.Error = "Container not found"
	w.WriteHeader(http.StatusNotFound)
	Send(w, response)
}

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

func (response ContainerCreationErrorResponse) Send(w web.ResponseWriter) {
	response.Status = "error"
	w.WriteHeader(http.StatusInternalServerError)
	Send(w, response)
}

func (response ContainerControlErrorResponse) Send(w web.ResponseWriter) {
	response.Status = "error"
	w.WriteHeader(http.StatusInternalServerError)
	Send(w, response)
}

func (response ServerErrorResponse) Send(w web.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
}

func Send(w web.ResponseWriter, response Response) {
	answer, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.Write(answer)
}
