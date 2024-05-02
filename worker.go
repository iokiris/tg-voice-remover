package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log"
	"net/http"
	"time"
)

func StartWorker(bot *tgbotapi.BotAPI, broker *Broker) {
	for {
		messages, err := broker.ReceiveMessage()
		if err != nil {
			if err != nil {
				log.Printf("Failed to receive messages: %v, sleeping 5s...", err)
				time.Sleep(time.Second * 5)
				continue
			}
		}
		for message := range messages {
			var task Task
			if err := json.Unmarshal(message.Body, &task); err != nil {
				log.Println("Error decoding task:", err)
				continue
			}
			switch task.Type {

			case "audio_process":
				var audioTask AudioTask
				if err := json.Unmarshal(task.Data, &audioTask); err != nil {
					log.Println("Error decoding audio")
					continue
				}
				if err := processAudio(bot, audioTask, task.UnixStartTime); err != nil {
					continue
				}
			}
		}
	}
}

func downloadAudio(bot *tgbotapi.BotAPI, fileID string) ([]byte, error) {
	fileConfig := tgbotapi.FileConfig{
		FileID: fileID,
	}
	fileInfo, err := bot.GetFile(fileConfig)
	if err != nil {
		return nil, err
	}
	url, err := bot.GetFileDirectURL(fileInfo.FileID)
	if err != nil {
		return nil, err
	}
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("Error closing file")
		}
	}(response.Body)
	content, err := io.ReadAll(response.Body)
	return content, nil
}

func processAudio(bot *tgbotapi.BotAPI, task AudioTask, startTime int64) error {
	content, err := downloadAudio(bot, task.AudioID)
	if err != nil {
		log.Println("Download error")
		return err
	}
	log.Println("Processing audio: ", task.AudioName)
	instrumental, err := removeVocals(content)
	if err != nil {
		log.Println("removeVocals error")
		return err
	}
	b := bytes.NewReader(instrumental)
	msg := tgbotapi.NewAudio(task.ChatID, tgbotapi.FileReader{
		Name:   "@voiceeraser_bot - " + task.AudioName,
		Reader: b,
	})
	msg.Caption = fmt.Sprintf("Время обработки: %dс.", time.Now().Unix()-startTime)
	_, err = bot.Send(msg)
	if err != nil {

		return err
	}
	return nil
}
