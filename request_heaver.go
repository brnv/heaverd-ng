package main

import "encoding/json"

type (
	ContainerStartRequest struct {
		BaseRequest
	}

	ContainerStopRequest struct {
		BaseRequest
	}

	ContainerDestroyRequest struct {
		BaseRequest
	}
)

func (request ContainerStartRequest) Execute() Response {
	request.Action = "start"

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
			return ContainerStartErrorResponse{
				BaseResponse: BaseResponse{
					ResponseHost: request.RequestHost,
					Error:        err.Error(),
				},
			}
		}
		return response
	}

	if response.(ClusterResponse).Error != "" {
		return ContainerStartErrorResponse{
			BaseResponse: BaseResponse{
				ResponseHost: request.RequestHost,
				Error:        response.(ClusterResponse).Error,
			},
		}
	}

	return ContainerStartResponse{
		BaseResponse: response.(ClusterResponse).BaseResponse,
	}
}

func (request ContainerStopRequest) Execute() Response {
	request.Action = "stop"

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
			return ContainerStopErrorResponse{
				BaseResponse: BaseResponse{
					ResponseHost: request.RequestHost,
					Error:        err.Error(),
				},
			}
		}
		return response
	}

	if response.(ClusterResponse).Error != "" {
		return ContainerStopErrorResponse{
			BaseResponse: BaseResponse{
				ResponseHost: request.RequestHost,
				Error:        response.(ClusterResponse).Error,
			},
		}
	}

	return ContainerStopResponse{
		BaseResponse: response.(ClusterResponse).BaseResponse,
	}
}

func (request ContainerDestroyRequest) Execute() Response {
	request.Action = "destroy"

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
		return ContainerDestroyErrorResponse{
			BaseResponse: BaseResponse{
				ResponseHost: request.RequestHost,
				Error:        err.Error(),
			},
		}
	}

	if response.(ClusterResponse).Error != "" {
		return ContainerDestroyErrorResponse{
			BaseResponse: BaseResponse{
				ResponseHost: request.RequestHost,
				Error:        response.(ClusterResponse).Error,
			},
		}
	}

	return ContainerDestroyErrorResponse{
		BaseResponse: response.(ClusterResponse).BaseResponse,
	}
}
