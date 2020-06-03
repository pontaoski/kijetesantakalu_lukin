package main

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

type EmbedHeader struct {
	Icon string
	Text string
	URL  string
}

type EmbedField struct {
	Title  string
	Body   string
	Inline bool
}

type Keyboard struct {
	Title string
	URL   string
}

func keyboardsToTG(kbs []Keyboard) (ret tgbotapi.InlineKeyboardMarkup) {
	for idx, kb := range kbs {
		ret.InlineKeyboard = append(ret.InlineKeyboard, []tgbotapi.InlineKeyboardButton{
			{
				Text: kb.Title,
				URL:  &kbs[idx].URL,
			},
		})
	}
	return
}

//Embed ...
type Embed struct {
	Header EmbedHeader
	Title  EmbedHeader
	Footer EmbedHeader

	Fields []EmbedField
	Body   string
	Colour int
}

type EmbedList struct {
	ItemTypeName string
	Embeds       []Embed
}

// Constants for message embed character limits
const (
	EmbedLimitTitle       = 256
	EmbedLimitDescription = 2048
	EmbedLimitFieldValue  = 1024
	EmbedLimitFieldName   = 256
	EmbedLimitField       = 25
	EmbedLimitFooter      = 2048
	EmbedLimit            = 4000
)

// Truncate truncates any embed value over the character limit.
func (e *Embed) Truncate() {
	if len(e.Body) > EmbedLimitDescription {
		e.Body = e.Body[:EmbedLimitDescription]
	}
	for _, v := range e.Fields {
		if len(v.Title) > EmbedLimitFieldName {
			v.Title = v.Title[:EmbedLimitFieldName]
		}

		if len(v.Body) > EmbedLimitFieldValue {
			v.Body = v.Body[:EmbedLimitFieldValue]
		}
	}
	if len(e.Title.Text) > EmbedLimitTitle {
		e.Title.Text = e.Title.Text[:EmbedLimitTitle]
	}
	if len(e.Footer.Text) > EmbedLimitFooter {
		e.Footer.Text = e.Footer.Text[:EmbedLimitFooter]
	}
}
