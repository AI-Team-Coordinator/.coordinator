package infra

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// ErrorResponse matches Core: {"error": "E500: message"}
type ErrorResponse struct {
	Error     string `json:"error"`
	Details   string `json:"details,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
}

func AddErrorCode(code, message string, httpStatus ...int) (string, int) {
	status := 0
	if len(httpStatus) > 0 {
		status = httpStatus[0]
	}
	if status == 0 {
		status = GetHTTPStatus(code)
	}
	return fmt.Sprintf("%s: %s", code, message), status
}

func GetHTTPStatus(code string) int {
	if len(code) < 4 {
		return http.StatusInternalServerError
	}
	rangeCode := code[:4]
	switch {
	case strings.HasPrefix(rangeCode, "E00"):
		return http.StatusUnauthorized
	case strings.HasPrefix(rangeCode, "E10"):
		return http.StatusBadRequest
	case strings.HasPrefix(rangeCode, "E20"):
		return http.StatusNotFound
	case strings.HasPrefix(rangeCode, "E30"):
		return http.StatusServiceUnavailable
	case code == "E409":
		return http.StatusConflict
	case strings.HasPrefix(rangeCode, "E40"):
		return http.StatusBadRequest
	case strings.HasPrefix(rangeCode, "E50"):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func ReturnJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("[coordinator] write json: %v", err)
	}
}

func ReturnError(w http.ResponseWriter, code, message string, httpStatus ...int) {
	errorMsg, status := AddErrorCode(code, message, httpStatus...)
	if status >= 500 {
		log.Printf("[coordinator] %s", errorMsg)
	}
	ReturnJSON(w, status, ErrorResponse{Error: errorMsg})
}
