package main

import (
	"bufio"
	"errors"
	"io"
	"net"
)

type BaseRequest struct {
	Action        string
	RequestHost   string
	ContainerName string
}

func (request BaseRequest) GetTargetHostname() string {
	if request.RequestHost == "" {
		host, _ := request.GetHostnameByContainer()
		return host
	}
	return request.RequestHost
}

func (request BaseRequest) GetHostnameByContainer() (string, error) {
	for hostname, host := range Cluster() {
		if _, ok := host.Containers[request.ContainerName]; ok {
			return hostname, nil
		}
	}

	return "", errors.New("")
}

func (request BaseRequest) IsHostExists() bool {
	if _, ok := Cluster()[request.RequestHost]; !ok {
		return false
	}

	return true
}

func (request BaseRequest) IsContainerExists() bool {
	if _, ok := Cluster()[request.RequestHost].Containers[request.ContainerName]; !ok {
		return false
	}

	return true
}

func (request BaseRequest) SendMessage(message []byte) ([]byte, error) {
	connection, err := net.Dial("tcp", request.RequestHost+":"+clusterPort)
	if err != nil {
		return []byte{}, err
	}
	defer connection.Close()

	connection.Write(message)

	answer, err := bufio.NewReader(connection).ReadString('\n')
	if err != nil {
		if err != io.EOF {
			return []byte{}, err
		}
	}

	return []byte(answer), nil
}

func (request BaseRequest) GetErrorResponse() Response {
	if request.RequestHost == "" {
		var err error
		request.RequestHost, err = request.GetHostnameByContainer()
		if err != nil {
			response := CantFindContainerHostnameResponse{}
			response.ResponseHost = Hostinfo.Hostname
			return response
		}
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
