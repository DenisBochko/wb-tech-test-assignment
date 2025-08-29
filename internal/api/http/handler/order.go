package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"wb-tech-test-assignment/internal/apperrors"

	"wb-tech-test-assignment/internal/model"

	"github.com/go-chi/chi/v5"
)

type OrderService interface {
	GetOrder(ctx context.Context, orderUID string) (model.Order, error)
}

func GetOrder(ctx context.Context, svc OrderService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		orderUID := chi.URLParam(r, "orderUID")

		order, err := svc.GetOrder(ctx, orderUID)
		if err != nil {
			if errors.Is(err, apperrors.ErrOrderNotFound) {
				w.WriteHeader(http.StatusNotFound)

				errResp := responseWithMessage{
					Status:  statusError,
					Message: apperrors.ErrOrderNotFound.Error(),
				}

				if err := json.NewEncoder(w).Encode(errResp); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}

				return
			}

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

		resp := responseWithData{
			Status: statusSuccess,
			Data:   order,
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
