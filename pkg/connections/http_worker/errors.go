package http_worker

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Method string `json:"method"`
	URL    string `json:"url"`
	Error  string `json:"error"`
}

var (
	unknownError = ErrorResponse{
		Method: "__unknown__",
		URL:    "__unknown__",
		Error:  "Internal error",
	}
	unknownErrorJSON []byte
)

func init() {
	var err error

	unknownErrorJSON, err = json.MarshalIndent(unknownError, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
}

func HandleErrorResponse(
	responseWriter http.ResponseWriter,
	request *http.Request,
	statusCode int,
	errorToSend error,
) error {
	response := ErrorResponse{
		Method: request.Method,
		URL:    request.URL.String(),
		Error:  errorToSend.Error(),
	}

	return HandleResponse(
		responseWriter,
		request,
		statusCode,
		response,
	)
}
