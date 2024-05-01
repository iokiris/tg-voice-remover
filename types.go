package main

import (
	"encoding/json"
)

type Task struct {
	Type          string          `json:"type"`
	Data          json.RawMessage `json:"data"`
	Repeated      bool            `json:"repeated"`
	UnixStartTime int64           `json:"unix_start_time"`
}

type AudioTask struct {
	ChatID    int64  `json:"chatID"`
	AudioID   string `json:"audioID"`
	AudioName string `json:"audioName"`
}
