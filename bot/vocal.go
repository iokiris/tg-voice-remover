package main

import (
	"bytes"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"mime/multipart"
	"net/http"
)

func removeVocals(input []byte, task AudioTask) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", "audio.mp3")
	if err != nil {
		log.Println("Error creating form file:", err)
		return err
	}
	_, err = part.Write(input)
	if err != nil {
		log.Println("Error writing to form file:", err)
		return err
	}
	err = writer.Close()
	if err != nil {
		log.Println("Error closing writer:", err)
		return err
	}

	url := fmt.Sprintf("http://fastapi-tvr:8000/process-audio/?audio_name=%s&chat_id=%d&callback=%s",
		task.AudioName,
		task.ChatID,
		"/callback")

	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		log.Println("Error creating request:", err)
		return err
	}
	if err != nil {
		log.Println("Error creating request:", err)
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg := tgbotapi.NewMessage(task.ChatID, fmt.Sprintf("Ваше аудио *%s* не было отправлено в обработку из-за ошибки.", task.AudioName))
		msg.ParseMode = "MarkdownV2"
		_, err = globalBotH.bot.Send(msg)
		if err != nil {
			return err
		}
		log.Printf("No-OK http status: %s", resp.Status)
		return fmt.Errorf("NO-OK http status: %s", resp.Status)
	}
	msg := tgbotapi.NewMessage(task.ChatID, fmt.Sprintf("Ваше аудио *%s* подано в обработку.", task.AudioName))
	msg.ParseMode = "MarkdownV2"
	_, err = globalBotH.bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}
