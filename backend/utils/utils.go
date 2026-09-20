package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func ParseJSON(r *http.Request, payload any) error {
	log.Printf("Parsing JSON request body %v", r.Body)
	if r.Body == nil {
		return fmt.Errorf("missing request body")
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(payload)
}

func WriteSuccess(w http.ResponseWriter, status int, message string, data any) error {
	resp := SuccessResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(resp)
}

func WriteError(w http.ResponseWriter, status int, message string, err any) error {
	resp := ErrorResponse{
		Status:  "error",
		Message: message,
		Error:   err,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(resp)
}
