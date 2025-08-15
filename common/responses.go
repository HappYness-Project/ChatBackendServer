package common

import (
	"encoding/json"
	"net/http"
)

type ProblemDetails struct {
	Type      string `json:"type,omitempty"`
	Title     string `json:"title"`
	ErrorCode string `json:"error_code"`
	Detail    string `json:"detail,omitempty"`
	Instance  string `json:"instance,omitempty"`
}
type JSONResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func WriteJsonWithEncode(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
func ErrorResponse(w http.ResponseWriter, status int, problem ProblemDetails) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(problem)
}
