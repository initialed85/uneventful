package http_worker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func GetURLPathParts(url *url.URL) []string {
	return strings.Split(strings.Trim(url.Path, "/"), "/")
}

func GetSuccessResponse(detail string) Response {
	return Response{Success: true, Detail: detail}
}

func GetErrorResponse(detail string, method string, url string, err error) ErrorResponse {
	return ErrorResponse{Response: Response{Success: false, Detail: detail}, Method: method, URL: url, Error: err.Error()}
}

func HandleResponse(responseWriter http.ResponseWriter, request *http.Request, statusCode int, response interface{}) error {
	_ = request

	responseJSON, jsonErr := json.MarshalIndent(response, "", "    ")
	if jsonErr != nil {
		statusCode = 500
		responseJSON = unknownErrorJSON
	}

	responseWriter.WriteHeader(statusCode)
	_, writeErr := responseWriter.Write(responseJSON)
	if writeErr != nil {
		if jsonErr != nil {
			return fmt.Errorf(fmt.Sprintf("experienced writeErr=%v while handling jsonErr=%v", writeErr, jsonErr))
		}
		return writeErr
	}

	return nil
}

func HandleErrorResponse(responseWriter http.ResponseWriter, request *http.Request, statusCode int, errorToSend error) error {
	return HandleResponse(responseWriter, request, statusCode, GetErrorResponse("An error occurred", request.Method, request.URL.String(), errorToSend))
}
