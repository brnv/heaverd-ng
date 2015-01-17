package main

import "encoding/json"

type ContainerPushRequest struct {
	BaseRequest
	Image string
}

func (request ContainerPushRequest) Execute() Response {
	request.Action = "push"

	if request.RequestHost == "" {
		request.RequestHost = request.FindHostname()
	}

	errorResponse := request.Validate()
	if errorResponse != nil {
		return errorResponse
	}

	payload, _ := json.Marshal(request)
	response, err := request.Send(request.RequestHost, payload)

	if err != nil {
		if response == nil {
			return ContainerPushErrorResponse{
				BaseResponse: BaseResponse{
					ResponseHost: request.RequestHost,
					Error:        err.Error(),
				},
			}
		}
		return response
	}

	if response.(ClusterResponse).Error != "" {
		return ContainerPushErrorResponse{
			BaseResponse: BaseResponse{
				ResponseHost: request.RequestHost,
				Error:        response.(ClusterResponse).Error,
			},
		}
	}

	return ContainerPushResponse{
		BaseResponse: response.(ClusterResponse).BaseResponse,
	}
}

func (request ContainerPushRequest) Validate() Response {
	if request.Image == "" {
		response := ImageMustBeGivenErrorResponse{}
		response.ResponseHost = Hostinfo.Hostname
		return response
	}

	return request.BaseRequest.Validate()
}
