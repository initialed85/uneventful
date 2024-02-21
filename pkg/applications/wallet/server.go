package wallet

import (
	"fmt"

	"github.com/initialed85/uneventful/pkg/domains"
	"github.com/segmentio/ksuid"
)

type Server struct {
	domains.Server
}

func NewServer() *Server {
	name := fmt.Sprintf("server_%v", domainName)

	s := Server{Server: domains.NewServer(name, domainName, NewReader(name), NewCaller(name, ksuid.New()))}

	return &s
}
