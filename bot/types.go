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

type errWithText struct {
	Message string
	Err     error
}

func MessageError(err error, msg string) *errWithText {
	if msg == "" && err != nil {
		msg = "Неизвестная ошибка"
	}
	return &errWithText{
		Message: msg,
		Err:     err,
	}
}

func SMessageError(err error, msg string) *errWithText {
	if msg == "" && err != nil {
		msg = "Внутренняя ошибка сервиса"
	}
	return &errWithText{
		Message: msg,
		Err:     err,
	}
}
