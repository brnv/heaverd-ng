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
	Image    []string
	SshKey   string
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
	request.RequestHost, err = request.GetMostSuitableHost()
	if err != nil {
		response := CantAssignAnyHostResponse{}
		response.Error = err.Error()
		response.ResponseHost = Hostinfo.Hostname
		return response
	}

	if request.Token > Cluster()[request.RequestHost].LastUpdateTimestamp {
		response := StaleDataResponse{}
		response.ResponseHost = Hostinfo.Hostname
		return response
	}

	request.Action = "create"

	err = StoreRequestAsIntent(request)
	if err != nil {
		if strings.Contains(err.Error(), etcdErrCodeKeyExists) {
			response := NotUniqueNameResponse{}
			response.ResponseHost = request.RequestHost
			return response
		}
	}

	payload, _ := json.Marshal(request)

	raw, err := request.SendMessage(payload)

	if err != nil {
		response := ContainerCreationErrorResponse{
			BaseResponse: BaseResponse{
				Error: err.Error(),
			},
		}
		return response
	}

	if strings.Contains(string(raw), "Error") {
		response := ContainerCreationErrorResponse{
			BaseResponse: BaseResponse{
				Error: string(raw),
			},
		}
		return response
	}

	createdContainer := lxc.Container{}
	json.Unmarshal(raw, &createdContainer)

	response := ContainerCreateResponse{
		Container: createdContainer,
	}
	response.ResponseHost = request.RequestHost

	return response
}
