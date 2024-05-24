package main

import (
	"bytes"
	"fmt"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"time"
)

func removeVocals(input []byte, task AudioTask) *errWithText {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", "audio.mp3")
	if err != nil {
		log.Println("Error creating form file:", err)
		return SMessageError(err, "")
	}
	_, err = part.Write(input)
	if err != nil {
		log.Println("Error writing to form file:")
		return SMessageError(err, "")
	}
	err = writer.Close()
	if err != nil {
		log.Println("Error closing writer:")
		return SMessageError(err, "")
	}

	url := fmt.Sprintf("http://fastapi-tvr:8000/process-audio/?audio_name=%s&chat_id=%d&callback=%s",
		task.AudioName,
		task.ChatID,
		"/callback")

	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		log.Println("Error creating post request:")
		return SMessageError(err, "")
	}
	if err != nil {
		log.Println("Error creating request:")
		return SMessageError(err, "")
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request with forms: ", err)
		if e, ok := err.(net.Error); ok && e.Timeout() {
			return MessageError(
				err,
				"Возможно, файл слишком большой.",
			)
		}
		return SMessageError(err, "Аудиофайл не был обработан по неизвестной причине")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if err != nil {
			return SMessageError(err, "Сервис обработки вернул ошибку.")
		}
		log.Printf("No-OK http status: %s", resp.Status)
		return SMessageError(err, "")
	}
	if err != nil {
		log.Println("cannot send message: ", err)
		return MessageError(err, "не удалось отправить сообщение об обработке")
	}
	return nil
}
