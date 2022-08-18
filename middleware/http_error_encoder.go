package middleware

import (
	"context"
	"encoding/json"
	"net/http"
)

type ErrorWrapper struct {
	Error string `json:"error"`
}

func ErrorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	w.WriteHeader(err2code(err))
	json.NewEncoder(w).Encode(ErrorWrapper{Error: err.Error()})
}

func err2code(err error) int {
	return http.StatusInternalServerError
}
