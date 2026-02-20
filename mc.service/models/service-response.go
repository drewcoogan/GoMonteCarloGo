package models

type serviceResponse[T any] struct {
	Data  *T     `json:"data"`
	Error string `json:"error"`
}

func GetServiceResponseOk[T any](data *T) serviceResponse[T] {
	return serviceResponse[T]{
		Data:  data,
		Error: "",
	}
}

func GetServiceResponseError[T any](errorMessage string) serviceResponse[T] {
	return serviceResponse[T]{
		Data:  nil,
		Error: errorMessage,
	}
}
