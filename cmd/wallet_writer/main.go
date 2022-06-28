package main

import (
	"github.com/initialed85/uneventful/internal/helpers"
	"github.com/initialed85/uneventful/pkg/applications/wallet"
	"github.com/initialed85/uneventful/pkg/lifecycles"
	"log"
)

func main() {
	entityID, err := helpers.GetEntityIDFromEnvironmentVariable("")
	if err != nil {
		log.Fatal(err)
	}

	writer := wallet.NewWriter(entityID)

	lifecycles.Run(writer)
}
