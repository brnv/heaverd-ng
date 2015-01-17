package main

import (
	"bufio"
	"encoding/json"
	"io"
	"net"
)

type BaseRequest struct {
	Action        string
	RequestHost   string
	ContainerName string
}

func (request BaseRequest) Validate() Response {
	if request.RequestHost == "" {
		response := CantFindHostnameByContainerResponse{}
		response.ResponseHost = Hostinfo.Hostname
		return response
	}

	if !request.IsHostExists() {
		response := HostNotFoundResponse{}
		response.ResponseHost = Hostinfo.Hostname
		return response
	}

	if !request.IsContainerExists() {
		response := ContainerNotFoundResponse{}
		response.ResponseHost = Hostinfo.Hostname
		return response
	}

	return nil
}

func (request BaseRequest) Send(host string, payload []byte) (Response, error) {
	connection, err := net.Dial("tcp", host+":"+clusterPort)
	if err != nil {
		return nil, err
	}

	defer connection.Close()

	connection.Write(payload)

	answer, err := bufio.NewReader(connection).ReadBytes('\n')
	if err != nil {
		if err != io.EOF {
			return nil, err
		}
	}

	var response ClusterResponse
	_ = json.Unmarshal(answer, &response)

	return response, nil
}

func (request BaseRequest) FindHostname() string {
	for hostname, host := range Cluster() {
		if _, ok := host.Containers[request.ContainerName]; ok {
			return hostname
		}
	}

	return ""
}

func (request BaseRequest) IsHostExists() bool {
	if _, ok := Cluster()[request.RequestHost]; !ok {
		return false
	}

	return true
}

func (request BaseRequest) IsContainerExists() bool {
	if _, ok := Cluster()[request.RequestHost].
		Containers[request.ContainerName]; !ok {
		return false
	}

	return true
}
