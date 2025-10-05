package domain

import (
	"encoding/json"
	"time"
)

type Field struct {
	Name  string
	Value string
}

type Message struct {
	ID     string
	Values []Field
}

type Event struct {
	MessageID  string          `json:"messageId"`
	JID        string          `json:"id"`
	Kind       string          `json:"kind"`
	Target     string          `json:"target"`
	Source     string          `json:"source"`
	RetryLimit int64           `json:"retryLimit"`
	Payload    json.RawMessage `json:"payload"`
}

type EventResult struct {
	JID        string
	MsgID      string
	Success    bool
	Error      string
	RetryLimit int64
	Attempts   int64
	Duration   time.Duration
}

type PendingInfo struct {
	Total    int64
	FirstID  string
	LatestID string
	Group    string
}

type StreamEntry map[string]string
