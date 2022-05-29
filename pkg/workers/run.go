package workers

import (
	"github.com/initialed85/uneventful/internal/helpers"
	"log"
)

func Run(worker Worker) {
	helpers.SetLogFormat()

	err := worker.Start()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		_ = worker.Stop()
	}()

	helpers.WaitForSigInt()
}
