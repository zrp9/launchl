package core

type Email struct {
	To      []string       `json:"to"`
	From    string         `json:"from"`
	Data    map[string]any `json:"data,omitempty"`
	Subject string         `json:"subject,omitempty"`
}
