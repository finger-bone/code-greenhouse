package router

import "time"

type Wrapper struct {
	Timestamp    string      `json:"timestamp"`
	IsError      bool        `json:"isError"`
	ErrorMessage string      `json:"errorMessage"`
	Data         interface{} `json:"data"`
}

func BuildError(errorMessage string) *Wrapper {
	return &Wrapper{
		Timestamp:    time.Now().Format(time.RFC3339),
		IsError:      true,
		ErrorMessage: errorMessage,
		Data:         nil,
	}
}

func BuildResponse(data interface{}) *Wrapper {
	return &Wrapper{
		Timestamp:    time.Now().Format(time.RFC3339),
		IsError:      false,
		ErrorMessage: "",
		Data:         data,
	}
}
