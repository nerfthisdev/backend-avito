package httpapi

import (
	"encoding/json"
	"net/http"
)

type ErrorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	var body ErrorBody
	body.Error.Code = code
	body.Error.Message = message

	_ = json.NewEncoder(w).Encode(body)
}
