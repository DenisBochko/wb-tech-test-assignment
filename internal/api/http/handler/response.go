package handler

const (
	statusSuccess = "success"
	statusError   = "error"
)

type responseWithMessage struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type responseWithData struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}
