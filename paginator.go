package main

import (
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var telegramMutex = &sync.Mutex{}
var telegramPaginators map[int]*telegramPaginator = make(map[int]*telegramPaginator)

type telegramPaginator struct {
	Pages []tgbotapi.MessageConfig
	Index int

	message  *tgbotapi.Message
	bot      *Bot
	lastused time.Time
}

func NewPaginator(bot *Bot) telegramPaginator {
	return telegramPaginator{
		bot: bot,
	}
}

func (p *telegramPaginator) AddPage(msg tgbotapi.MessageConfig) {
	p.Pages = append(p.Pages, msg)
}

var paginatorKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			"Previous",
			"previous",
		),
		tgbotapi.NewInlineKeyboardButtonData(
			"Next",
			"next",
		),
	),
)

func (p *telegramPaginator) Send() {
	p.Index = 0
	send := p.Pages[p.Index]
	send.ReplyMarkup = paginatorKeyboard
	msg, err := p.bot.Send(send)
	if err == nil {
		p.message = &msg
		p.lastused = time.Now()
		telegramPaginators[msg.MessageID] = p
	}
}

func init() {
	go telegramCleaner()
}

func telegramCleaner() {
	for {
		time.Sleep(5 * time.Minute)
		var rmkeys []int
		for key, cmd := range telegramPaginators {
			if time.Now().Sub(cmd.lastused) >= 10*time.Minute {
				rmkeys = append(rmkeys, key)
			}
		}
		for _, key := range rmkeys {
			telegramMutex.Lock()
			delete(telegramPaginators, key)
			telegramMutex.Unlock()
		}
	}
}

func (p *telegramPaginator) Prev() {
	p.Index--
	if p.Index < 0 {
		p.Index = len(p.Pages) - 1
	}
	send := p.Pages[p.Index]
	send.ReplyMarkup = paginatorKeyboard
	edit := tgbotapi.NewEditMessageText(p.message.Chat.ID, p.message.MessageID, "")
	edit.Text = p.Pages[p.Index].Text
	edit.ParseMode = p.Pages[p.Index].ParseMode
	kb := paginatorKeyboard
	edit.ReplyMarkup = &kb
	msg, err := p.bot.Send(edit)
	if err != nil {
		p.message = &msg
		p.lastused = time.Now()
	}
}

func (p *telegramPaginator) Next() {
	p.Index++
	if p.Index+1 > len(p.Pages) {
		p.Index = 0
	}
	send := p.Pages[p.Index]
	send.ReplyMarkup = paginatorKeyboard
	edit := tgbotapi.NewEditMessageText(p.message.Chat.ID, p.message.MessageID, "")
	edit.Text = p.Pages[p.Index].Text
	edit.ParseMode = p.Pages[p.Index].ParseMode
	kb := paginatorKeyboard
	edit.ReplyMarkup = &kb
	msg, err := p.bot.Send(edit)
	if err != nil {
		p.message = &msg
		p.lastused = time.Now()
	}
}

func TelegramPaginatorHandler(messageID int, direction string) {
	if val, ok := telegramPaginators[messageID]; ok {
		if direction == "previous" {
			val.Prev()
		} else {
			val.Next()
		}
	}
}
