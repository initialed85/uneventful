package calls

import (
	"encoding/json"
)

type Response struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func NewResponseFromError(
	err error,
) *Response {
	errString := ""
	if err != nil {
		errString = err.Error()
	}

	r := Response{
		Success: err == nil,
		Error:   errString,
	}

	return &r
}

func ResponseFromJSON(data []byte) (*Response, error) {
	r := Response{}

	err := json.Unmarshal(data, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (r *Response) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}
