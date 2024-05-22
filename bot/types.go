package main

import (
	"encoding/json"
)

type Task struct {
	Type          string          `json:"type"`
	Data          json.RawMessage `json:"data"`
	Repeated      bool            `json:"repeated"`
	UnixStartTime int64           `json:"unix_start_time"`
	ID            string          `json:"id"`
}

type AudioTask struct {
	ChatID    int64  `json:"chatID"`
	AudioID   string `json:"audioID"`
	AudioName string `json:"audioName"`
}

type ProcessedTask struct {
	Status   string `json:"status"`
	FilePath string `json:"file_path,omitempty"`
}
