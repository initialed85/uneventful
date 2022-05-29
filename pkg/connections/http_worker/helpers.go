package http_worker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func GetURLPathParts(url *url.URL) []string {
	return strings.Split(
		strings.Trim(
			url.Path,
			"/",
		),
		"/",
	)
}

func HandleResponse(
	responseWriter http.ResponseWriter,
	request *http.Request,
	statusCode int,
	response interface{},
) error {
	responseJSON, jsonErr := json.MarshalIndent(response, "", "    ")
	if jsonErr != nil {
		statusCode = 500
		responseJSON = unknownErrorJSON
	}

	responseWriter.WriteHeader(statusCode)
	_, writeErr := responseWriter.Write(responseJSON)
	if writeErr != nil {
		if jsonErr != nil {
			return fmt.Errorf(
				fmt.Sprintf(
					"experienced writeErr=%v while handling jsonErr=%v",
					writeErr,
					jsonErr,
				),
			)
		}
		return writeErr
	}

	return nil
}
