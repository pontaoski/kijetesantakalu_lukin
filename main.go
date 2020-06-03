package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alecthomas/repr"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var telegramBot Bot

func main() {
	fmt.Println("Initializing kijetesantakalu lukin...")

	bot, err := tgbotapi.NewBotAPI(config.Telegram.Token)
	telegramBot = NewBot(bot)
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
			if strings.HasPrefix(update.Message.Text, "repr.Println(chatData)") {
				chat := GetChat(update.Message.Chat.ID)
				telegramBot.queue <- tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("<pre><code>%s</code></pre>", repr.String(chat)))
			}
			const prefix = "chatData.subscriptions := []string{"
			const suffix = "}"
			if strings.HasPrefix(update.Message.Text, prefix) && strings.HasSuffix(update.Message.Text, suffix) {
				split := strings.Split(strings.TrimPrefix(strings.TrimSuffix(update.Message.Text, suffix), prefix), ",")
				var cleaned []string
				for _, str := range split {
					cleaned = append(cleaned, strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(str), "\""), "\""))
				}
				var groups []string
				var projects []string
				for _, cleanedStr := range cleaned {
					if strings.Contains(cleanedStr, "/") {
						projects = append(projects, cleanedStr)
					} else {
						groups = append(groups, cleanedStr)
					}
				}
				gData, err := json.Marshal(groups)
				if err != nil {
					telegramBot.queue <- tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, there was an error marshalling data: "+err.Error())
				}
				pData, err := json.Marshal(projects)
				if err != nil {
					telegramBot.queue <- tgbotapi.NewMessage(update.Message.Chat.ID, "Sorry, there was an error marshalling data: "+err.Error())
				}
				chat := GetChat(update.Message.Chat.ID)
				chat.UpdateGroupData(string(gData))
				chat.UpdateProjectData(string(pData))
				telegramBot.queue <- tgbotapi.NewMessage(update.Message.Chat.ID, "Subscriptions updated!\n")
			}
			if update.Message.ForwardFromChat != nil {
				println(update.Message.ForwardFromChat.ID)
			}
		}
	}
}
