package workers

import "log"

func Setup(workers ...Worker) (err error) {
	for _, worker := range workers {
		log.Printf("starting %v", worker.GetName())
		err = worker.Start()
		if err != nil {
			return err
		}
		log.Printf("started %v", worker.GetName())
	}

	return nil
}

func Teardown(workers ...Worker) (err error) {
	for _, worker := range workers {
		log.Printf("stopping %v", worker.GetName())
		err = worker.Stop()
		if err != nil {
			return err
		}
		log.Printf("stopped %v", worker.GetName())
	}

	return nil
}
