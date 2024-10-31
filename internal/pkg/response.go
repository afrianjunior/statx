package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ErrorResponse struct {
	IsSuccess bool   `json:"is_success"`
	Stack     string `json:"stack"`
	Message   string `json:"message"`
}

type BaseResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func jsonResponse(w http.ResponseWriter, d any, c int) {
	dj, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(c)

	_, _ = fmt.Fprintf(w, "%s", dj)
}

// jsonResponseUsingBase write json body response using struct BaseResponse
func jsonResponseUsingBase(w http.ResponseWriter, msg string, payload any, err error, c int) {
	resp := BaseResponse{
		Message: msg,
		Success: err == nil,
		Data:    payload,
	}

	dj, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(c)

	_, _ = fmt.Fprintf(w, "%s", dj)
}
