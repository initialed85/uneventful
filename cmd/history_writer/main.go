package main

import (
	"github.com/initialed85/uneventful/internal/helpers"
	"github.com/initialed85/uneventful/pkg/domains/history"
	"github.com/initialed85/uneventful/pkg/workers"
	"log"
)

func main() {
	entityID, err := helpers.GetEntityIDFromEnvironmentVariable("")
	if err != nil {
		log.Fatal(err)
	}

	writer := history.NewWriter(entityID)

	workers.Run(writer)
}
