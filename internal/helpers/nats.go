package helpers

import (
	"github.com/initialed85/uneventful/internal/constants"
	"github.com/nats-io/nats.go"
)

func GetNatsConn() (natsConn *nats.Conn, err error) {
	natsURL, err := GetEnvironmentVariable("NATS_URL", false, constants.DefaultNatsURL)
	if err != nil {
		return nil, err
	}

	natsConn, err = nats.Connect(natsURL, nats.MaxReconnects(0))
	if err != nil {
		return nil, err
	}

	return natsConn, nil
}
