package history

import (
	"github.com/initialed85/uneventful/pkg/model"
	"github.com/initialed85/uneventful/pkg/model/calls"
	"github.com/initialed85/uneventful/pkg/model/events"
	"github.com/segmentio/ksuid"
)

type Writer struct {
	*model.Writer
	name string
}

func NewWriter(
	entityID ksuid.KSUID,
) *Writer {
	name := domainName

	w := Writer{
		name: name,
	}

	w.Writer = model.NewWriterWithOverrides(
		name,
		entityID,
		func(event *events.Event, request *calls.Request) error {
			return nil
		},
		"event.>",
		name,
		true,
		true,
		false,
	)

	return &w
}
