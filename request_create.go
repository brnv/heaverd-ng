package main

import (
	"encoding/json"
	"strings"

	"github.com/brnv/heaverd-ng/libscore"
)

type ContainerCreateRequest struct {
	BaseRequest
	PoolName string
	Image    []string
	SshKey   string
	Token    int64
	Ip       string
}

const etcdErrCodeKeyExists = "105"

func (request ContainerCreateRequest) Execute() Response {
	request.Action = "create"
	request.RequestHost = request.GetMostSuitableHost()

	errorResponse := request.Validate()
	if errorResponse != nil {
		return errorResponse
	}

	err := StoreRequestAsIntent(request)
	if err != nil {
		if strings.Contains(err.Error(), etcdErrCodeKeyExists) {
			return NotUniqueNameResponse{
				BaseResponse: BaseResponse{
					ResponseHost: request.RequestHost,
				},
			}
		}

		return ContainerCreateErrorResponse{
			BaseResponse: BaseResponse{
				ResponseHost: request.RequestHost,
				Error:        err.Error(),
			},
		}
	}

	payload, _ := json.Marshal(request)
	response, err := request.Send(request.RequestHost, payload)

	if err != nil {
		return ContainerCreateErrorResponse{
			BaseResponse: BaseResponse{
				ResponseHost: request.RequestHost,
				Error:        err.Error(),
			},
		}
	}

	if response.(ClusterResponse).Error != "" {
		return ContainerCreateErrorResponse{
			BaseResponse: BaseResponse{
				ResponseHost: request.RequestHost,
				Error:        response.(ClusterResponse).Error,
			},
		}
	}

	return ContainerCreateResponse{
		BaseResponse: response.(ClusterResponse).BaseResponse,
		Container:    response.(ClusterResponse).Container,
	}
}

func (request ContainerCreateRequest) Validate() Response {
	if request.RequestHost == "" {
		return NoSuitableHostFoundErrorResponse{
			BaseResponse: BaseResponse{
				ResponseHost: Hostinfo.Hostname,
			},
		}
	}

	cluster := Cluster()
	for _, host := range cluster {
		if request.Token > host.LastUpdateTimestamp {
			return StaleDataResponse{
				BaseResponse: BaseResponse{
					ResponseHost: Hostinfo.Hostname,
				},
			}
		}
	}

	if len(request.Image) == 0 {
		response := ImageMustBeGivenErrorResponse{}
		response.ResponseHost = Hostinfo.Hostname
		return response
	}

	return nil
}

func (request ContainerCreateRequest) GetMostSuitableHost() string {
	segments := libscore.Segments(Cluster(request.PoolName))
	suitableHost, _ := libscore.ChooseHost(request.ContainerName, segments)

	return suitableHost
}
