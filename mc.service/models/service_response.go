package models

type ServiceResponse[T any] struct {
	Data  *T     `json:"data"`
	Error string `json:"error"`
}

func GetServiceResponseOk[T any](data *T) ServiceResponse[T] {
	return ServiceResponse[T]{
		Data:  data,
		Error: "",
	}
}

func GetServiceResponseError(errorMessage string) ServiceResponse[any] {
	return ServiceResponse[any]{
		Data:  nil,
		Error: errorMessage,
	}
}
