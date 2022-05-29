package helpers

import "time"

func GetNow() time.Time {
	return time.Now().UTC()
}
