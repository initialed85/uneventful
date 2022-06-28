package history

import (
	"github.com/initialed85/uneventful/pkg/models"
	"github.com/segmentio/ksuid"
)

type Writer struct {
	models.Writer
	name string
}

func NewWriter(entityID ksuid.KSUID) *Writer {
	name := domainName

	w := Writer{name: name}

	w.Writer = models.NewWriterWithOverrides(
		name,
		entityID,
		func() (interface{}, error) {
			return nil, nil
		},
		"event.>",
		name,
		true,
		true,
		false,
	)

	return &w
}
