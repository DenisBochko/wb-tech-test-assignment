package handler

const (
	statusSuccess = "success"
)

type responseWithMessage struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
