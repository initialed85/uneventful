package events

import (
	"encoding/json"
	"fmt"
	"github.com/initialed85/uneventful/internal/helpers"
	"github.com/jackc/pgtype"
	"github.com/segmentio/ksuid"
	"time"
)

type Event struct {
	EventID       ksuid.KSUID     `json:"event_id"`
	CorrelationID ksuid.KSUID     `json:"correlation_id"`
	Timestamp     time.Time       `json:"timestamp"`
	SourceName    string          `json:"source_name"`
	SourceID      ksuid.KSUID     `json:"source_uuid"`
	TypeName      string          `json:"type_name"`
	Data          json.RawMessage `json:"data"`
}

func ToJSON(e *Event) (
	[]byte,
	error,
) {
	return json.Marshal(&e)
}

func (e *Event) String() string {
	return fmt.Sprintf(
		"Event{id=%s, type=%#+v, size=%vB}",
		e.EventID,
		e.TypeName,
		len(e.Data),
	)
}

func FromJSON(data []byte) (*Event, error) {
	e := Event{}

	err := json.Unmarshal(data, &e)

	return &e, err
}

func NewWithCorrelation(
	correlationID ksuid.KSUID,
	typeName string,
	data json.RawMessage,
) *Event {
	e := Event{
		EventID:       ksuid.New(),
		CorrelationID: correlationID,
		TypeName:      typeName,
		Data:          data,
		Timestamp:     helpers.GetNow(),
	}

	return &e
}

func NewWithoutCorrelation(
	typeName string,
	data json.RawMessage,
) *Event {
	e := Event{
		EventID:   ksuid.New(),
		Timestamp: helpers.GetNow(),
		TypeName:  typeName,
		Data:      data,
	}

	return &e
}

func (e *Event) SetSource(
	name string,
	id ksuid.KSUID,
) {
	e.SourceName = name
	e.SourceID = id
}

func (e *Event) ToJSON() (
	[]byte,
	error,
) {
	return ToJSON(e)
}

func (e *Event) FromJSON(data []byte) error {
	event, err := FromJSON(data)
	if err != nil {
		return err
	}

	e.EventID = event.EventID
	e.CorrelationID = event.CorrelationID
	e.Timestamp = event.Timestamp
	e.SourceName = event.SourceName
	e.SourceID = event.SourceID
	e.TypeName = event.TypeName
	e.Data = event.Data

	return nil
}

func (e *Event) ToConsumedDatabaseEvent() (
	*DatabaseEvent,
	error,
) {
	jsonData, err := e.Data.MarshalJSON()
	if err != nil {
		return nil, err
	}

	jsonbData := pgtype.JSONB{}

	err = jsonbData.Scan(jsonData)
	if err != nil {
		return nil, err
	}

	return &DatabaseEvent{
		EventID:       e.EventID.String(),
		CorrelationID: e.CorrelationID.String(),
		Timestamp:     e.Timestamp,
		SourceName:    e.SourceName,
		SourceID:      e.SourceID.String(),
		TypeName:      e.TypeName,
		Data:          jsonbData,
		IsHandled:     false,
	}, nil
}

func (e *Event) ToDatabaseEvent() (
	*DatabaseEvent,
	error,
) {
	jsonData, err := e.Data.MarshalJSON()
	if err != nil {
		return nil, err
	}

	jsonbData := pgtype.JSONB{}

	err = jsonbData.Scan(jsonData)
	if err != nil {
		return nil, err
	}

	return &DatabaseEvent{
		EventID:       e.EventID.String(),
		CorrelationID: e.CorrelationID.String(),
		Timestamp:     e.Timestamp,
		SourceName:    e.SourceName,
		SourceID:      e.SourceID.String(),
		TypeName:      e.TypeName,
		Data:          jsonbData,
		IsHandled:     false,
	}, nil
}
