package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
				go func() {
					var audioTask AudioTask
					log.Printf("Audiotask %s putted from queue\n", task.ID)
					if err := json.Unmarshal(task.Data, &audioTask); err != nil {
						log.Println("Error decoding audio")
						msg := tgbotapi.NewMessage(audioTask.ChatID,
							fmt.Sprintf("Не удается распознать аудио *%s*", audioTask.AudioName))
						msg.ParseMode = "MarkdownV2"
						_, err := bot.Send(msg)
						if err != nil {
							log.Println("Error sending error to client: ", err)
							return
						}
					}
					if err, text := processAudio(bot, audioTask); err != nil {
						sText := QMString(fmt.Sprintf(
							"Не удается отправить в обработку аудио *%s*. \n%s",
							audioTask.AudioName,
							text,
						),
						)
						msg := tgbotapi.NewMessage(audioTask.ChatID, sText)
						msg.ParseMode = "MarkdownV2"
						_, err := bot.Send(msg)
						if err != nil {
							log.Println("Error sending error to client: ", err)
							return
						}
					} else {
						msg := tgbotapi.NewMessage(audioTask.ChatID,
							fmt.Sprintf("Ваше аудио *%s* подано в обработку", audioTask.AudioName))
						msg.ParseMode = "MarkdownV2"
						_, err = bot.Send(msg)
					}
				}()
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

func processAudio(bot *tgbotapi.BotAPI, task AudioTask) (error, string) {
	content, err := downloadAudio(bot, task.AudioID)
	if err != nil {
		log.Println("Download error")
		return err, "Не удалось скачать файл"
	}
	log.Println("Processing audio: ", task.AudioName)
	pErr := removeVocals(content, task)
	log.Println(pErr)
	if pErr != nil {
		log.Println("removeVocals error")
		return pErr.Err, pErr.Message
	}
	return nil, ""
}

type CallbackRequest struct {
	Status      string `json:"status"`
	ProcessTime int    `json:"process_time,omitempty"`
	FileData    []byte `json:"file_data,omitempty"`
	ChatID      int64  `json:"chat_id,omitempty"`
	AudioName   string `json:"audio_name,omitempty"`
}

func callbackWorker(w http.ResponseWriter, r *http.Request) {
	var cbReq CallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&cbReq); err != nil {
		http.Error(w, fmt.Sprintf("failed to decode request body: %v", err), http.StatusBadRequest)
		return
	}

	switch cbReq.Status {
	case "ready":
		go func(audioName string, fileData []byte, chatID int64, processTime int) {
			err := sendAudioFile(audioName, fileData, chatID, processTime)
			if err != nil {
				log.Println("Error sending audio file:", err)
				return
			}
		}(cbReq.AudioName, cbReq.FileData, cbReq.ChatID, cbReq.ProcessTime)

	case "error":
		fmt.Printf("error processing audio '%s': ", cbReq.AudioName)

	default:
		http.Error(w, fmt.Sprintf("unknown status: %s", cbReq.Status), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func sendAudioFile(filename string, fileData []byte, chatID int64, ptime int) error {
	b := bytes.NewReader(fileData)
	msg := tgbotapi.NewAudio(chatID, tgbotapi.FileReader{
		Name:   "@voiceeraser_bot - " + filename,
		Reader: b,
	})
	msg.Caption = fmt.Sprintf("Время обработки: %dс.", ptime)
	_, err := globalBotH.bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}
