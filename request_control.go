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
	errorResponse := request.GetErrorResponse()
	if errorResponse != nil {
		return errorResponse
	}

	request.RequestHost = request.GetTargetHostname()

	payload, _ := json.Marshal(request)

	raw, err := request.SendMessage(payload)
	if err != nil {
		return ContainerStartErrorResponse{
			BaseResponse: BaseResponse{
				ResponseHost: request.RequestHost,
				Error:        err.Error(),
			},
		}
	}

	var response ContainerStartResponse

	err = json.Unmarshal(raw, &response)

	if response.Error != "" {
		return ContainerStartErrorResponse{
			BaseResponse: response.BaseResponse,
		}
	}

	return response
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
			BaseResponse: BaseResponse{
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
			BaseResponse: BaseResponse{
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
