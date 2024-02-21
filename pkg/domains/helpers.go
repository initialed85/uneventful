package domains

import (
	"log"
	"net/http"

	"github.com/initialed85/uneventful/pkg/workers/http_worker"
)

func handledErrorResponse(innerErr error, outerErr error, responseWriter http.ResponseWriter, request *http.Request, statusCode int, server Server) bool {
	if outerErr == nil {
		outerErr = innerErr
	}

	if innerErr == nil {
		return false
	}

	err := http_worker.HandleErrorResponse(responseWriter, request, statusCode, outerErr)
	if err != nil {
		log.Printf("%v - warning: %v", server.GetName(), err)
	}

	return true
}
