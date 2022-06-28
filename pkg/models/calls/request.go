package calls

import "encoding/json"

type Request struct {
	Endpoint string          `json:"endpoint"`
	Data     json.RawMessage `json:"data"`
}

func RequestFromJSON(data []byte) (*Request, error) {
	r := Request{}

	err := json.Unmarshal(data, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (r *Request) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}
