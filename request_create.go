package main

import (
	"encoding/json"
	"strings"

	"github.com/brnv/heaverd-ng/liblxc"
	"github.com/brnv/heaverd-ng/libscore"
)

type ContainerCreateRequest struct {
	BaseRequest
	PoolName string
	Image    []string `json:"image"`
	SshKey   string   `json:"key"`
	Token    int64
}

func (request ContainerCreateRequest) GetMostSuitableHost() (string, error) {
	segments := libscore.Segments(Cluster(request.PoolName))

	suitableHost, err := libscore.ChooseHost(request.ContainerName, segments)
	if err != nil {
		return "", err
	}

	return suitableHost, nil
}

func (request ContainerCreateRequest) Execute() Response {
	var err error
	request.Host, err = request.GetMostSuitableHost()

	if err != nil {
		response := CantAssignAnyHostResponse{}
		response.Error = err.Error()
		response.ResponseHost = Hostinfo.Hostname
		return response
	}

	if request.Token > Cluster()[request.Host].LastUpdateTimestamp {
		response := StaleDataResponse{}
		response.ResponseHost = Hostinfo.Hostname
		return response
	}

	err = StoreRequestAsIntent(request)
	if err != nil {
		if strings.Contains(err.Error(), etcdErrCodeKeyExists) {
			response := NotUniqueNameResponse{}
			response.ResponseHost = request.Host
			return response
		}
	}

	hostAnswer, err := request.SendMessage(
		request.MakeMessage("container-create", request.ContainerName))

	if err != nil {
		response := ContainerCreationErrorResponse{
			ErrorResponse: ErrorResponse{
				Error: err.Error(),
			},
		}
		return response
	}

	if strings.Contains(string(hostAnswer), "Error") {
		response := ContainerCreationErrorResponse{
			ErrorResponse: ErrorResponse{
				Error: string(hostAnswer),
			},
		}
		return response
	}

	createdContainer := lxc.Container{}
	json.Unmarshal(hostAnswer, &createdContainer)

	response := ContainerCreatedResponse{
		Container: createdContainer,
	}
	response.ResponseHost = request.Host

	return response
}
