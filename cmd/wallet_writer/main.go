package main

import (
	"github.com/initialed85/uneventful/internal/helpers"
	"github.com/initialed85/uneventful/pkg/domains/wallet"
	"github.com/initialed85/uneventful/pkg/workers"
	"log"
)

func main() {
	entityID, err := helpers.GetEntityIDFromEnvironmentVariable("")
	if err != nil {
		log.Fatal(err)
	}

	writer := wallet.NewWriter(entityID)

	workers.Run(writer)
}
