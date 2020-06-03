package main

import (
	"encoding/json"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/time/rate"
)

var telegramBot Bot

func main() {
	fmt.Println("Initializing kijetesantakalu lukin...")

	bot, err := tgbotapi.NewBotAPI(config.Telegram.Token)
	telegramBot = Bot{bot, make(chan tgbotapi.MessageConfig, 1024), rate.NewLimiter(5, 5)}
	if err != nil {
		fmt.Println("Error creating Telegram session: ", err.Error())
		return
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := telegramBot.GetUpdatesChan(u)

	fmt.Println("kijetesantakalu lukin is now running.")
	go telegramBot.QueueLoop()
	defer db.Close()
	go GitLab()
	for update := range updates {
		if update.CallbackQuery != nil && update.CallbackQuery.Message != nil {
			TelegramPaginatorHandler(update.CallbackQuery.Message.MessageID, update.CallbackQuery.Data)
		}
		if update.Message != nil {
			if strings.HasPrefix(update.Message.Text, "let subscriptions = [") && strings.HasSuffix(update.Message.Text, "]") {
				split := strings.Split(strings.TrimPrefix(strings.TrimSuffix(update.Message.Text, "]"), "let subscriptions = ["), ",")
				var cleaned []string
				for _, str := range split {
					cleaned = append(cleaned, strings.TrimSpace(str))
				}
				data, err := json.Marshal(cleaned)
				if err != nil {
					telegramBot.queue <- tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, there was an error marshalling data: "+err.Error())
				}
				chat := GetChat(update.Message.Chat.ID)
				chat.UpdateData(string(data))
				telegramBot.queue <- tgbotapi.NewMessage(update.Message.Chat.ID, "Subscriptions updated!\n")
			}
			if update.Message.ForwardFromChat != nil {
				println(update.Message.ForwardFromChat.ID)
			}
		}
	}
}
