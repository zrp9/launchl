package core

import "encoding/json"

type Email struct {
	To   []string        `json:"to"`
	From string          `json:"from"`
	Data json.RawMessage `json:"data,omitempty"`
	//Data    map[string]any `json:"data,omitempty"`
	Subject string `json:"subject,omitempty"`
}
