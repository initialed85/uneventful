package helpers

import "time"

func GetErrorChannel() chan error {
	return make(chan error)
}

func WaitForError(errors chan error, timeout time.Duration) error {
	timer := time.NewTimer(timeout)

wait:
	for {
		select {
		case <-timer.C:
			break wait
		case err := <-errors:
			return err
		}
	}

	return nil
}
