package main

import (
	"github.com/initialed85/uneventful/pkg/applications/wallet"
	"github.com/initialed85/uneventful/pkg/lifecycles"
)

func main() {
	server := wallet.NewServer()

	lifecycles.Run(server)
}
