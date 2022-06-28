package lifecycles

import (
	"github.com/initialed85/uneventful/internal/helpers"
	"log"
)

func Run(worker Worker) {
	helpers.SetLogFormat()

	log.Printf("%#+v starting...", worker)

	err := worker.Start()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		log.Printf("%#+v stopping...", worker)

		_ = worker.Stop()

		log.Printf("%#+v stopped.", worker)
	}()

	log.Printf("%#+v started.", worker)

	helpers.WaitForSigInt()
}
