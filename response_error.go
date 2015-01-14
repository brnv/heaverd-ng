package main

import (
	"net/http"

	"github.com/brnv/web"
)

type (
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

func (response CantFindContainerHostnameResponse) Send(w web.ResponseWriter) {
	response.Status = "error"
	response.Error = "Can't find host by given container"
	w.WriteHeader(http.StatusNotFound)
	ResponseSend(w, response)
}

func (response HeaverErrorResponse) Send(w web.ResponseWriter) {
	response.Status = "error"
	w.WriteHeader(http.StatusConflict)
	ResponseSend(w, response)
}

func (response HostNotFoundResponse) Send(w web.ResponseWriter) {
	response.Status = "error"
	response.Error = "Host not found"
	w.WriteHeader(http.StatusNotFound)
	ResponseSend(w, response)
}

func (response ContainerNotFoundResponse) Send(w web.ResponseWriter) {
	response.Status = "error"
	response.Error = "Container not found"
	w.WriteHeader(http.StatusNotFound)
	ResponseSend(w, response)
}

func (response CantAssignAnyHostResponse) Send(w web.ResponseWriter) {
	response.Status = "error"
	response.Error = "No suitable host found"
	w.WriteHeader(http.StatusNotFound)
	ResponseSend(w, response)
}

func (response StaleDataResponse) Send(w web.ResponseWriter) {
	response.Status = "error"
	response.Error = "Stale data"
	w.WriteHeader(http.StatusTeapot)
	ResponseSend(w, response)
}

func (response NotUniqueNameResponse) Send(w web.ResponseWriter) {
	response.Status = "error"
	response.Error = "Not unique name"
	w.WriteHeader(http.StatusConflict)
	ResponseSend(w, response)
}

func (response ContainerCreationErrorResponse) Send(w web.ResponseWriter) {
	response.Status = "error"
	w.WriteHeader(http.StatusInternalServerError)
	ResponseSend(w, response)
}

func (response ContainerControlErrorResponse) Send(w web.ResponseWriter) {
	response.Status = "error"
	w.WriteHeader(http.StatusInternalServerError)
	ResponseSend(w, response)
}

func (response ServerErrorResponse) Send(w web.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
}
