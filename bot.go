package main

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/time/rate"
)

type Bot struct {
	*tgbotapi.BotAPI
	queue       chan tgbotapi.MessageConfig
	globalLimit *rate.Limiter
	limits      map[int64]*rate.Limiter
}

func telegramEmbed(d Embed) tgbotapi.MessageConfig {
	d.Truncate()
	var fields []string
	for _, field := range d.Fields {
		fields = append(fields, fmt.Sprintf("%s: %s", field.Title, field.Body))
	}
	msg := tgbotapi.NewMessage(0, "")
	msg.Text = fmt.Sprintf(`<i>%s</i>

<b>%s</b>
<i>%s</i>
%s

%s`, d.Header.Text, d.Title.Text, d.Body, strings.Join(fields, "\n"), d.Footer.Text)
	msg.ParseMode = tgbotapi.ModeHTML
	return msg
}

func NewBot(bot *tgbotapi.BotAPI) Bot {
	ret := Bot{}
	ret.BotAPI = bot
	ret.queue = make(chan tgbotapi.MessageConfig, 1024)
	ret.globalLimit = rate.NewLimiter(10, 1)
	ret.limits = make(map[int64]*rate.Limiter)
	return ret
}

func (b *Bot) GetLimiter(chatID int64) *rate.Limiter {
	if val, ok := b.limits[chatID]; ok {
		return val
	}
	b.limits[chatID] = rate.NewLimiter(0.25, 1)
	return b.limits[chatID]
}

func (b *Bot) QueueLoop() {
	for {
		select {
		case msg := <-b.queue:
			go func(msg tgbotapi.MessageConfig) {
				b.globalLimit.Wait(context.Background())
				b.GetLimiter(msg.ChatID).Wait(context.Background())
				_, err := b.Send(msg)
				if err != nil {
					println(err.Error())
				}
			}(msg)
		}
	}
}

func (b *Bot) SendMessage(chatID int64, content interface{}, keyboards ...Keyboard) {
	switch content.(type) {
	case string:
		b.queue <- tgbotapi.NewMessage(chatID, content.(string))
	case Embed:
		msg := telegramEmbed(content.(Embed))
		msg.ReplyMarkup = keyboardsToTG(keyboards)
		msg.ChatID = chatID
		b.queue <- msg
	case EmbedList:
		telegramMutex.Lock()
		defer telegramMutex.Unlock()
		paginator := NewPaginator(b)
		title := "Item"
		if content.(EmbedList).ItemTypeName != "" {
			title = content.(EmbedList).ItemTypeName
		}
		for idx, page := range content.(EmbedList).Embeds {
			page.Footer.Text = fmt.Sprintf("%s %d out of %d", title, idx+1, len(content.(EmbedList).Embeds))
			msg := telegramEmbed(page)
			msg.ChatID = chatID
			paginator.AddPage(msg)
		}
		paginator.Send()
	}
}
