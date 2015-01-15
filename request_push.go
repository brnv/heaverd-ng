package main

type ContainerPushRequest struct {
	BaseRequest
	Image string `json:"image"`
}

func (request ContainerPushRequest) Execute() Response {
	request.RequestHost = request.GetTargetHostname()

	response := ContainerPushResponse{
		Image: request.Image,
	}
	response.ResponseHost = request.RequestHost

	return response
}
