package main

import (
	"encoding/json"
	"strings"

	"github.com/brnv/heaverd-ng/liblxc"
	"github.com/brnv/heaverd-ng/libscore"
)

type Request interface {
	Execute()
}

type CreateRequest struct {
	Host          string
	ContainerName string
	PoolName      string
	Image         []string `json:"image"`
	SshKey        string   `json:"key"`
	Token         int64
}

func (request CreateRequest) GetMostSuitableHost() (string, error) {
	segments := libscore.Segments(Cluster(request.PoolName))

	suitableHost, err := libscore.ChooseHost(request.ContainerName, segments)
	if err != nil {
		return "", err
	}

	return suitableHost, nil
}

func (request CreateRequest) Execute() Response {
	targetHost, err := request.GetMostSuitableHost()

	if err != nil {
		response := CantAssignAnyHostResponse{}
		response.Error = err.Error()
		response.ResponseHost = Hostinfo.Hostname
		return response
	}

	if request.Token > Cluster()[targetHost].LastUpdateTimestamp {
		response := StaleDataResponse{}
		response.ResponseHost = Hostinfo.Hostname
		return response
	}

	request.Host = targetHost
	err = StoreRequestAsIntent(request)
	if err != nil {
		if strings.Contains(err.Error(), etcdErrCodeKeyExists) {
			response := NotUniqueNameResponse{}
			response.ResponseHost = targetHost
			return response
		}
	}

	creationMessage := makeMessage("container-create", request.ContainerName)
	hostAnswer, _ := sendMessageToHost(targetHost, clusterPort, creationMessage)

	if strings.Contains(string(hostAnswer), "Error") {
		response := ErrorResponse{}
		response.Error = string(hostAnswer)
		return response
	}

	createdContainer := lxc.Container{}
	json.Unmarshal(hostAnswer, &createdContainer)

	response := ContainerCreatedResponse{
		Container: createdContainer,
	}
	response.ResponseHost = targetHost

	return response
}
