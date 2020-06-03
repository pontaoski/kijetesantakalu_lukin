package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/time/rate"
)

type Bot struct {
	*tgbotapi.BotAPI
	queue   chan tgbotapi.MessageConfig
	limiter *rate.Limiter
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

func (b *Bot) QueueLoop() {
	for {
		select {
		case msg := <-b.queue:
			b.limiter.Wait(context.Background())
			b.Send(msg)
		}
	}
}

func (b *Bot) SendMessage(chatID int64, content interface{}, keyboards ...Keyboard) {
	time.Sleep(time.Second / 4)
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
