package api

import (
	"encoding/json"
	"net/http"
)

// write data to http response
func writeResponse(w http.ResponseWriter, code int, data interface{}) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	rsp, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Json marshal error", http.StatusInternalServerError)
		return
	}
	w.Write(rsp)
}

// Http response with data
func httpResponse(w http.ResponseWriter, data interface{}) {
	writeResponse(w, http.StatusOK, data)
}

// Http response with error
type ErrorResponse struct {
	Code    int           `json:"code,omitempty"`
	Message string        `json:"message,omitempty"`
	Details []interface{} `json:"details,omitempty"`
}

func httpError(w http.ResponseWriter, code int, err string) {
	errRsp := &ErrorResponse{
		Code:    code,
		Message: err,
	}
	writeResponse(w, code, errRsp)
}
