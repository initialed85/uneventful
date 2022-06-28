package domains

import (
	"encoding/json"
	"fmt"
	"github.com/initialed85/uneventful/pkg/lifecycles"
	"github.com/initialed85/uneventful/pkg/models"
	"github.com/initialed85/uneventful/pkg/workers/http_worker"
	"github.com/segmentio/ksuid"
	"io/ioutil"
	"net/http"
)

type Server interface {
	lifecycles.Worker
}

type ServerImplementation struct {
	lifecycles.Worker
	domainName string
	reader     models.Reader
	caller     models.Caller
	httpServer *http_worker.Worker
}

func NewServer(name string, domainName string, reader models.Reader, caller models.Caller) *ServerImplementation {
	s := ServerImplementation{domainName: domainName, reader: reader, caller: caller}

	s.httpServer = http_worker.New(name, defaultHTTPServerPort, map[string]http.HandlerFunc{fmt.Sprintf("/%v/", domainName): s.handle})

	s.Worker = lifecycles.NewLazyWorker(name, s.setup, s.teardown)

	return &s
}

func (s *ServerImplementation) setup() (err error) {
	return lifecycles.Setup(s.reader, s.caller, s.httpServer)
}

func (s *ServerImplementation) teardown() (err error) {
	return lifecycles.Teardown(s.httpServer, s.caller, s.reader)
}

func (s *ServerImplementation) handle(responseWriter http.ResponseWriter, request *http.Request) {
	if !(request.Method == http.MethodGet || request.Method == http.MethodPost) {
		if handledErrorResponse(fmt.Errorf("method must be %v or %v", http.MethodGet, http.MethodPost), nil, responseWriter, request, 400, s) {
			return
		}
	}

	pathParts := http_worker.GetURLPathParts(request.URL)

	if len(pathParts) != 3 {
		if handledErrorResponse(fmt.Errorf("path must be '%v/[entity ksuid]/[endpoint]'", s.domainName), nil, responseWriter, request, 400, s) {
			return
		}
	}

	entityID, err := ksuid.Parse(pathParts[1])
	if handledErrorResponse(err, fmt.Errorf("entity ksuid %#+v could not be parsed: %v", pathParts[1], err), responseWriter, request, 400, s) {
		return
	}

	endpoint := pathParts[2]

	var handler models.Handler

	if request.Method == http.MethodGet {
		handler, err = s.reader.GetHandler(endpoint)
	}

	if request.Method == http.MethodPost {
		handler, err = s.caller.GetHandler(endpoint)
	}

	if err != nil {
		if handledErrorResponse(err, fmt.Errorf("no handler for endpoint=%#+v", endpoint), responseWriter, request, 400, s) {
			return
		}
	}

	var data []byte
	var requestBody interface{}
	var responseBody interface{}

	if request.Method == http.MethodPost {
		data, err = ioutil.ReadAll(request.Body)
		if handledErrorResponse(err, fmt.Errorf("failed to read data from request body: %v", err), responseWriter, request, 400, s) {
			return
		}

		defer func() {
			_ = request.Body.Close()
		}()

		err = json.Unmarshal(data, &requestBody)
		if handledErrorResponse(err, fmt.Errorf("failed to parse JSON from request body: %v", err), responseWriter, request, 400, s) {
			return
		}
	}

	responseBody, err = handler(entityID, requestBody)

	if handledErrorResponse(err, fmt.Errorf("failed to handle endpoint=%#+v: %v", endpoint, err), responseWriter, request, 400, s) {
		return
	}

	if request.Method == http.MethodGet {
		_ = http_worker.HandleResponse(responseWriter, request, 200, responseBody)
	}

	if request.Method == http.MethodPost {
		_ = http_worker.HandleResponse(responseWriter, request, 200, http_worker.GetSuccessResponse(fmt.Sprintf("handled endpoint=%#+v", endpoint)))
	}
}
