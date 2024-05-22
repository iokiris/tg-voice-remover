package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	callbackListenersCount = 4
)

type BotHandler struct {
	broker *Broker
	bot    *tgbotapi.BotAPI
}

var (
	mutex   sync.RWMutex
	taskMap = make(map[string]Task)
)

var globalBotH *BotHandler

func NewBotHandler(botToken string, webhookURL string, broker *Broker) (*BotHandler, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, err
	}
	//bot.Debug = true
	webhookConfig, err := tgbotapi.NewWebhook(webhookURL)
	if err != nil {
		return nil, fmt.Errorf("cannot add webhook: %v", err)
	}
	_, err = bot.Request(webhookConfig)
	if err != nil {
		return nil, fmt.Errorf("webhook doesn't respond: %v", err)
	}
	fmt.Println("Webhook is connected")
	globalBotH = &BotHandler{
		bot:    bot,
		broker: broker,
	}
	return globalBotH, nil
}

func (h *BotHandler) Start() {
	for {
		err := h.Run()
		if err != nil {
			log.Printf("[MAIN]: bot crashed: %v", err)
			log.Println("[MAIN]: restarting...")
		}
	}
}

func (h *BotHandler) Run() error {
	updates := h.bot.ListenForWebhook("/tg-hook")
	go func() {
		http.HandleFunc("/callback", callbackWorker)
		err := http.ListenAndServe("0.0.0.0:8000", nil)
		if err != nil {
			log.Panicln("Cannot start server: ", err)
		}
	}()
	for update := range updates {
		h.safeCall(h.bot, update)
	}
	return nil
}

func main() {
	broker := NewBroker()
	defer broker.Close()

	ngrokApiURL, err := getEnvWithRetry("NGROK_API_URL", 5, 10)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("NGROK_API: ", ngrokApiURL)
	ngrokURL, err := getNgrokPublicURL(ngrokApiURL)
	if err != nil {
		log.Fatal("Error gettin Ngrok URL: ", err)
	}
	webhookUrl := ngrokURL + "/tg-hook"
	fmt.Println("GO NGROK_URL:", webhookUrl)

	botToken, err := getEnvWithRetry("BOT_TOKEN", 1, 0)
	if err != nil {
		log.Fatal(err)
	}
	handler, err := NewBotHandler(
		botToken,
		webhookUrl,
		broker,
	)
	go StartWorker(handler.bot, broker)
	if err != nil {
		log.Fatalf("error when create bot handler")
	}
	handler.Start()
}

func createAndSendTask(broker *Broker, taskType string, data interface{}, bot *tgbotapi.BotAPI, chatID int64) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Этот файл не удается обработать.")
		_, err := bot.Send(msg)
		if err != nil {
			return err
		}
		return err
	}
	taskUUID := uuid.New().String()
	task := Task{
		ID:            taskUUID,
		Type:          taskType,
		Data:          jsonData,
		UnixStartTime: time.Now().Unix(),
	}
	encodedTask, err := json.Marshal(task)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "Не удалось создать задачу.")
		_, err := bot.Send(msg)
		if err != nil {
			return err
		}
		return err
	}
	return broker.PushMessage(encodedTask)
}

func (h *BotHandler) safeCall(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("[RECOVER]: Recovered in safeCall:", r)
		}
	}()

	h.onUpdate(bot, update)
}

func (h *BotHandler) onUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message != nil {
		if audio := update.Message.Audio; audio != nil {
			h.onAudio(update, audio)
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Отправь аудиофайл с которого нужно вырезать слова")
			_, err := bot.Send(msg)
			if err != nil {
				return
			}
		}
	}
}

func (h *BotHandler) onAudio(update tgbotapi.Update, audio *tgbotapi.Audio) {
	audioTask := AudioTask{
		ChatID:    update.Message.Chat.ID,
		AudioID:   audio.FileID,
		AudioName: audio.Title,
	}
	log.Println("New audioTask, title: ", audioTask.AudioName)
	if err := createAndSendTask(
		h.broker, "audio_process", audioTask, h.bot, update.Message.Chat.ID,
	); err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Этот файл не удается обработать.")
		_, err := h.bot.Send(msg)
		if err != nil {
			return
		}
		log.Println(err)
		return
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Добавил вас в очередь, ожидайте.")
	_, err := h.bot.Send(msg)
	if err != nil {
		return
	}
}

// delay in seconds
func getEnvWithRetry(envVar string, retries int, delay time.Duration) (string, error) {
	var value string
	for i := 0; i < retries; i++ {
		value = os.Getenv(envVar)
		if value != "" {
			return value, nil
		}
		log.Printf("Env-var %s not set, retrying after %dsec...", envVar, delay)
		time.Sleep(time.Second * delay)
	}
	return "", fmt.Errorf("environment variable %s not set after %d retries", envVar, retries)
}
