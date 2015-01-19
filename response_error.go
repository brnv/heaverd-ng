package main

import (
	"net/http"

	"github.com/brnv/web"
)

type (
	NoSuitableHostFoundErrorResponse struct {
		BaseResponse
	}

	StaleDataResponse struct {
		BaseResponse
	}

	NotUniqueNameResponse struct {
		BaseResponse
	}

	ContainerCreateErrorResponse struct {
		BaseResponse
	}

	CantFindHostnameByContainerResponse struct {
		BaseResponse
	}

	ContainerNotFoundResponse struct {
		BaseResponse
	}

	HostNotFoundResponse struct {
		BaseResponse
	}

	ContainerPushErrorResponse struct {
		BaseResponse
	}

	ContainerStartErrorResponse struct {
		BaseResponse
	}

	ContainerStopErrorResponse struct {
		BaseResponse
	}

	ContainerDestroyErrorResponse struct {
		BaseResponse
	}

	ImageMustBeGivenErrorResponse struct {
		BaseResponse
	}
)

func (response ImageMustBeGivenErrorResponse) Write(w web.ResponseWriter) {
	response.Status = "error"
	response.Error = "Image must be given"
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	ResponseSend(w, response)
}

func (response ContainerStopErrorResponse) Write(w web.ResponseWriter) {
	response.Status = "error"
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	ResponseSend(w, response)
}

func (response ContainerDestroyErrorResponse) Write(w web.ResponseWriter) {
	response.Status = "error"
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	ResponseSend(w, response)
}

func (response ContainerStartErrorResponse) Write(w web.ResponseWriter) {
	response.Status = "error"
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	ResponseSend(w, response)
}

func (response ContainerPushErrorResponse) Write(w web.ResponseWriter) {
	response.Status = "error"
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	ResponseSend(w, response)
}

func (response CantFindHostnameByContainerResponse) Write(w web.ResponseWriter) {
	response.Status = "error"
	response.Error = "Can't find host by given container"
	w.WriteHeader(http.StatusNotFound)
	ResponseSend(w, response)
}

func (response HostNotFoundResponse) Write(w web.ResponseWriter) {
	response.Status = "error"
	response.Error = "Host not found"
	w.WriteHeader(http.StatusNotFound)
	ResponseSend(w, response)
}

func (response ContainerNotFoundResponse) Write(w web.ResponseWriter) {
	response.Status = "error"
	response.Error = "Container not found"
	w.WriteHeader(http.StatusNotFound)
	ResponseSend(w, response)
}

func (response NoSuitableHostFoundErrorResponse) Write(w web.ResponseWriter) {
	response.Status = "error"
	response.Error = "No suitable host found"
	w.WriteHeader(http.StatusNotFound)
	ResponseSend(w, response)
}

func (response StaleDataResponse) Write(w web.ResponseWriter) {
	response.Status = "error"
	response.Error = "Stale data"
	w.WriteHeader(http.StatusTeapot)
	ResponseSend(w, response)
}

func (response NotUniqueNameResponse) Write(w web.ResponseWriter) {
	response.Status = "error"
	response.Error = "Not unique name"
	w.WriteHeader(http.StatusConflict)
	ResponseSend(w, response)
}

func (response ContainerCreateErrorResponse) Write(w web.ResponseWriter) {
	response.Status = "error"
	w.WriteHeader(http.StatusInternalServerError)
	ResponseSend(w, response)
}
