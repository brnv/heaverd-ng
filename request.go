package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"net"
)

type Request interface {
	Execute()
}

type (
	BaseRequest struct {
		Host          string
		ContainerName string
	}
)

func (request BaseRequest) GetHostnameByContainer() (string, error) {
	for hostname, host := range Cluster() {
		if _, ok := host.Containers[request.ContainerName]; ok {
			return hostname, nil
		}
	}

	return "", errors.New("")
}

func (request BaseRequest) IsHostExists() bool {
	if _, ok := Cluster()[request.Host]; !ok {
		return false
	}

	return true
}

func (request BaseRequest) IsContainerExists() bool {
	if _, ok := Cluster()[request.Host].Containers[request.ContainerName]; !ok {
		return false
	}

	return true
}

func (request BaseRequest) MakeMessage(tag string, body interface{}) []byte {
	message, _ := json.Marshal(struct {
		Tag  string
		Body interface{}
	}{
		tag,
		body,
	})

	return message
}

func (request BaseRequest) SendMessage(message []byte) ([]byte, error) {
	connection, err := net.Dial("tcp", request.Host+":"+clusterPort)
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
