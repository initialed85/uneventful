package states

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/initialed85/uneventful/internal/helpers"
	"github.com/jackc/pgtype"
	"github.com/segmentio/ksuid"
)

type State struct {
	VersionID uint64          `json:"version_id"`
	Timestamp time.Time       `json:"timestamp"`
	Name      string          `json:"name"`
	EntityID  ksuid.KSUID     `json:"entity_id"`
	Data      json.RawMessage `json:"data"`
}

func FromJSON(data []byte) (*State, error) {
	s := State{}

	err := json.Unmarshal(data, &s)

	return &s, err
}

func New(name string, entityID ksuid.KSUID, data json.RawMessage) *State {
	s := State{Timestamp: helpers.GetNow(), Name: name, EntityID: entityID, Data: data}

	return &s
}

func (s *State) String() string {
	return fmt.Sprintf("State{version=%v, size=%vB}", s.VersionID, len(s.Data))
}

func (s *State) ToJSON() ([]byte, error) {
	return json.Marshal(&s)
}

func (s *State) FromJSON(data []byte) error {
	state, err := FromJSON(data)
	if err != nil {
		return err
	}

	s.VersionID = state.VersionID
	s.Timestamp = state.Timestamp
	s.Name = state.Name
	s.EntityID = state.EntityID
	s.Data = state.Data

	return nil
}

func (s *State) ToDatabaseState() (*DatabaseState, error) {
	jsonData, err := s.Data.MarshalJSON()
	if err != nil {
		return nil, err
	}

	jsonbData := pgtype.JSONB{}

	err = jsonbData.Scan(jsonData)
	if err != nil {
		return nil, err
	}

	return &DatabaseState{VersionID: s.VersionID, Timestamp: s.Timestamp, Name: s.Name, EntityID: s.EntityID.String(), Data: jsonbData}, nil
}
