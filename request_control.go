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

	ContainerDestroyRequest struct {
		ContainerControlRequest
	}
)

func (request ContainerStartRequest) Execute() Response {
	request.Action = "start"
	return request.Send()
}

func (request ContainerStopRequest) Execute() Response {
	request.Action = "stop"
	return request.Send()
}

func (request ContainerDestroyRequest) Execute() Response {
	request.Action = "destroy"
	return request.Send()
}

func (request ContainerControlRequest) Send() Response {
	request.RequestHost = request.GetTargetHostname()

	errorResponse := request.GetErrorResponse()
	if errorResponse != nil {
		return errorResponse
	}

	payload, _ := json.Marshal(request)

	raw, err := request.SendMessage(payload)

	if err != nil {
		response := ContainerControlErrorResponse{
			ErrorResponse: ErrorResponse{
				Error: err.Error(),
			},
		}
		return response
	}

	answer := struct {
		From  string
		Msg   string
		Error string
		Token int64
		Code  int
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
		response := HeaverErrorResponse{}
		response.Error = answer.Error
		response.ResponseHost = answer.From
		return response
	}

	return ContainerControlResponse{
		BaseResponse: BaseResponse{
			ResponseHost: answer.From,
		},
		Token: answer.Token,
	}

}

func (request ContainerControlRequest) GetErrorResponse() Response {
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
