package wallet

import (
	"encoding/json"
	"fmt"
	"github.com/initialed85/uneventful/pkg/connections/http_worker"
	"github.com/initialed85/uneventful/pkg/workers"
	"github.com/segmentio/ksuid"
	"io/ioutil"
	"net/http"
)

type Server struct {
	workers.Worker
	reader     *Reader
	caller     *Caller
	httpServer *http_worker.Worker
}

func NewServer() *Server {
	name := fmt.Sprintf("server_%v", domainName)

	s := Server{
		reader: NewReader(name),
		caller: NewCaller(name, ksuid.New()),
	}

	s.httpServer = http_worker.New(
		name,
		defaultHTTPServerPort,
		map[string]http.HandlerFunc{
			fmt.Sprintf("/%v/", domainName): s.handleWallet,
		},
	)

	s.Worker = workers.NewLazyWorker(
		name,
		s.setup,
		s.teardown,
	)

	return &s
}

func (s *Server) setup() (err error) {
	return workers.Setup(s.reader, s.caller, s.httpServer)
}

func (s *Server) teardown() (err error) {
	return workers.Teardown(s.httpServer, s.caller, s.reader)
}

func (s *Server) handleBalance(
	responseWriter http.ResponseWriter,
	request *http.Request,
	entityID ksuid.KSUID,
) {
	if request.Method != http.MethodGet {
		if handledErrorResponse(
			fmt.Errorf("HTTP GET supported only"),
			nil,
			responseWriter,
			request,
			500,
			s,
		) {
			return
		}
	}

	balance, err := s.reader.GetBalance(entityID)
	if handledErrorResponse(
		err,
		fmt.Errorf("failed to get balance from walletState: %v", err),
		responseWriter,
		request,
		500,
		s,
	) {
		return
	}

	_ = http_worker.HandleResponse(
		responseWriter,
		request,
		200,
		balance,
	)
}

func (s *Server) handleTransactions(
	responseWriter http.ResponseWriter,
	request *http.Request,
	entityID ksuid.KSUID,
) {
	if request.Method != http.MethodGet {
		if handledErrorResponse(
			fmt.Errorf("HTTP GET supported only"),
			nil,
			responseWriter,
			request,
			500,
			s,
		) {
			return
		}
	}

	transactions, err := s.reader.GetTransactions(entityID)
	if handledErrorResponse(
		err,
		fmt.Errorf("failed to get transactions from walletState: %v", err),
		responseWriter,
		request,
		500,
		s,
	) {
		return
	}

	_ = http_worker.HandleResponse(
		responseWriter,
		request,
		200,
		transactions,
	)
}

func (s *Server) handleCredit(
	responseWriter http.ResponseWriter,
	request *http.Request,
	entityID ksuid.KSUID,
) {
	if request.Method != http.MethodPost {
		if handledErrorResponse(
			fmt.Errorf("HTTP POST supported only"),
			nil,
			responseWriter,
			request,
			500,
			s,
		) {
			return
		}
	}

	data, err := ioutil.ReadAll(request.Body)
	if handledErrorResponse(
		err,
		fmt.Errorf("failed to get amount from request body: %v", err),
		responseWriter,
		request,
		400,
		s,
	) {
		return
	}

	defer func() {
		_ = request.Body.Close()
	}()

	amount := Amount{}

	err = json.Unmarshal(data, &amount)
	if handledErrorResponse(
		err,
		fmt.Errorf("failed to get amount from request body: %v", err),
		responseWriter,
		request,
		400,
		s,
	) {
		return
	}

	err = s.caller.Credit(entityID, amount.Amount)
	if handledErrorResponse(
		err,
		fmt.Errorf("failed to credit %#+v: %v", amount, err),
		responseWriter,
		request,
		400,
		s,
	) {
		return
	}

	_ = http_worker.HandleResponse(
		responseWriter,
		request,
		200,
		struct {
			Success bool
		}{
			Success: true,
		},
	)
}

func (s *Server) handleDebit(
	responseWriter http.ResponseWriter,
	request *http.Request,
	entityID ksuid.KSUID,
) {
	if request.Method != http.MethodPost {
		if handledErrorResponse(
			fmt.Errorf("HTTP POST supported only"),
			nil,
			responseWriter,
			request,
			500,
			s,
		) {
			return
		}
	}

	data, err := ioutil.ReadAll(request.Body)
	if handledErrorResponse(
		err,
		fmt.Errorf("failed to get amount from request body: %v", err),
		responseWriter,
		request,
		400,
		s,
	) {
		return
	}

	defer func() {
		_ = request.Body.Close()
	}()

	amount := Amount{}

	err = json.Unmarshal(data, &amount)
	if handledErrorResponse(
		err,
		fmt.Errorf("failed to get amount from request body: %v", err),
		responseWriter,
		request,
		400,
		s,
	) {
		return
	}

	err = s.caller.Debit(entityID, amount.Amount)
	if handledErrorResponse(
		err,
		fmt.Errorf("failed to debit %#+v: %v", amount, err),
		responseWriter,
		request,
		400,
		s,
	) {
		return
	}

	_ = http_worker.HandleResponse(
		responseWriter,
		request,
		200,
		struct {
			Success bool
		}{
			Success: true,
		},
	)
}

func (s *Server) handleWallet(
	responseWriter http.ResponseWriter,
	request *http.Request,
) {
	pathParts := http_worker.GetURLPathParts(request.URL)

	if len(pathParts) != 3 {
		if handledErrorResponse(
			fmt.Errorf("path must be 'wallet/[entity ksuid]/[endpoint]'"),
			nil,
			responseWriter,
			request,
			400,
			s,
		) {
			return
		}
	}

	entityID, err := ksuid.Parse(pathParts[1])
	if handledErrorResponse(
		err,
		fmt.Errorf("entity ksuid %#+v could not be parsed: %v", pathParts[1], err),
		responseWriter,
		request,
		400,
		s,
	) {
		return
	}

	var handler func(http.ResponseWriter, *http.Request, ksuid.KSUID)

	switch pathParts[2] {
	case "balance":
		handler = s.handleBalance
	case "transactions":
		handler = s.handleTransactions
	case "credit":
		handler = s.handleCredit
	case "debit":
		handler = s.handleDebit
	default:
		if handledErrorResponse(
			fmt.Errorf("endpoint must be one of 'balance', 'transactions', 'credit' or 'debit'"),
			nil,
			responseWriter,
			request,
			400,
			s,
		) {
			return
		}
	}

	handler(responseWriter, request, entityID)
}
