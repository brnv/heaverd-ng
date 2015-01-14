package main

import "encoding/json"

type (
	ContainerControlRequest struct {
		BaseRequest
	}

	ContainerStartRequest struct {
		ContainerControlRequest
	}

	ContainerStopRequest struct {
		ContainerControlRequest
	}
)

func (request ContainerStartRequest) Execute() Response {
	return request.GetResponse("start")
}

func (request ContainerStopRequest) Execute() Response {
	return request.GetResponse("stop")
}

func (
	request ContainerControlRequest,
) SendControlMessage(action string) ([]byte, error) {
	return request.SendMessage(
		request.MakeMessage("container-control", struct {
			ContainerName string
			Action        string
		}{
			request.ContainerName,
			action,
		}),
	)
}

func (request ContainerControlRequest) GetErrorResponse() Response {
	if request.Host == "" {
		var err error
		request.Host, err = request.GetHostnameByContainer()
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

func (request ContainerControlRequest) GetResponse(action string) Response {
	errorResponse := request.GetErrorResponse()
	if errorResponse != nil {
		return errorResponse
	}

	raw, err := request.SendControlMessage(action)

	if err != nil {
		response := ContainerControlErrorResponse{
			ErrorResponse: ErrorResponse{
				Error: err.Error(),
			},
		}
		return response
	}

	answer := struct {
		From       string
		Msg        string
		Error      string
		LastUpdate int64
		Code       int
	}{}

	err = json.Unmarshal(raw, &answer)
	if err != nil {
		response := ServerErrorResponse{
			ErrorResponse: ErrorResponse{
				Error: err.Error(),
			},
		}
		return response
	}

	switch answer.Code {
	case 409:
		return HeaverErrorResponse{
			ErrorResponse: ErrorResponse{
				Error: answer.Error,
			},
		}
	}

	return ContainerControlResponse{
		BaseResponse: BaseResponse{
			ResponseHost: answer.From,
		},
		Token: answer.LastUpdate,
	}
}
