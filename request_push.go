package main

import "encoding/json"

type ContainerPushRequest struct {
	BaseRequest
	Image string
}

func (request ContainerPushRequest) Execute() Response {
	request.Action = "push"
	errorResponse := request.GetErrorResponse()
	if errorResponse != nil {
		return errorResponse
	}

	request.RequestHost = request.GetTargetHostname()

	payload, _ := json.Marshal(request)

	raw, err := request.SendMessage(payload)
	if err != nil {
		return ContainerPushErrorResponse{
			BaseResponse: BaseResponse{
				ResponseHost: request.RequestHost,
				Error:        err.Error(),
			},
		}
	}

	var response ContainerPushResponse

	err = json.Unmarshal(raw, &response)

	if response.Error != "" {
		return ContainerPushErrorResponse{
			BaseResponse: response.BaseResponse,
		}
	}

	return response
}
