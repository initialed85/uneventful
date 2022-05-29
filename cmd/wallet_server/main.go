package main

import (
	"github.com/initialed85/uneventful/pkg/domains/wallet"
	"github.com/initialed85/uneventful/pkg/workers"
)

func main() {
	server := wallet.NewServer()

	workers.Run(server)
}
