package http_worker

import (
	"encoding/json"
	"log"
)

type Response struct {
	Success bool   `json:"success"`
	Detail  string `json:"detail,omitempty"`
}

type ErrorResponse struct {
	Response
	Method string `json:"method"`
	URL    string `json:"url"`
	Error  string `json:"error"`
}

var (
	unknownError     = ErrorResponse{Method: "__unknown__", URL: "__unknown__", Error: "Internal error"}
	unknownErrorJSON []byte
)

func init() {
	var err error

	unknownErrorJSON, err = json.MarshalIndent(unknownError, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
}
