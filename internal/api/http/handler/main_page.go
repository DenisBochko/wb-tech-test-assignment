package handler

import (
	"encoding/json"
	"html/template"
	"net/http"
)

const (
	PathToHTMLTemplate = "templates/index.html"
)

func MainPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	t, err := template.ParseFiles(PathToHTMLTemplate)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusInternalServerError)

		errResp := responseWithMessage{
			Status:  statusError,
			Message: err.Error(),
		}

		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	if err := t.Execute(w, nil); err != nil {
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusInternalServerError)

		errResp := responseWithMessage{
			Status:  statusError,
			Message: err.Error(),
		}

		if err := json.NewEncoder(w).Encode(errResp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}
}
