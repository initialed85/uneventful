package helpers

import (
	"fmt"
	"github.com/segmentio/ksuid"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func GetEnvironmentVariable(
	key string,
	required bool,
	defaultValue string,
) (string, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		if required {
			return "", fmt.Errorf("%v environment variable not set", key)
		}
		value = defaultValue
	}

	return value, nil
}

func WaitForSigInt() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	<-c
}

func GetEntityIDFromString(
	entityID string,
) (ksuid.KSUID, error) {
	return ksuid.Parse(entityID)
}

func GetEntityIDFromEnvironmentVariable(
	prefix string,
) (ksuid.KSUID, error) {
	key := "ENTITY_ID"

	prefix = strings.TrimSpace(prefix)
	if prefix != "" {
		key = fmt.Sprintf("%v_%v", prefix, key)
	}

	entityIDString, err := GetEnvironmentVariable(key, true, "")
	if err != nil {
		return ksuid.KSUID{}, err
	}

	entityID, err := GetEntityIDFromString(entityIDString)
	if err != nil {
		return ksuid.KSUID{}, err
	}

	return entityID, nil
}
