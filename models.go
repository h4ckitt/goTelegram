package goTelegram

import "reflect"

//Bot : Main Bot Struct
type Bot struct {
	Me         user
	APIURL     string
	Handler    reflect.Value
	HandlerSet bool
	Keyboard   keyboard
}

type user struct {
	ID        int    `json:"id"`
	Firstname string `json:"first_name"`
	Username  string `json:"username"`
}

//Update : Stores Data From Request
type Update struct {
	UpdateID      int           `json:"update_id"`
	Message       message       `json:"message"`
	CallbackQuery callbackQuery `json:"callback_query"`
	Command       string
}

type message struct {
	MessageID int    `json:"message_id"`
	Text      string `json:"Text"`
	Chat      chat   `json:"chat"`
	From      user   `json:"from"`
}

type chat struct {
	ID int `json:"id"`
}

type inlineKeyboard struct {
	Text string `json:"text"`
	Data string `json:"callback_data"`
}

type replyBody struct {
	ChatID      string      `json:"chat_id,omitempty"`
	Text        string      `json:"text,omitempty"`
	ParseMode   string      `json:"parse_mode,omitempty"`
	ReplyMarkup replyMarkup `json:"reply_markup,omitempty"`
}

type replyMarkup struct {
	InlineKeyboard [][]inlineKeyboard `json:"inline_keyboard,omitempty"`
}

type callbackQuery struct {
	ID      string  `json:"id"`
	From    user    `json:"from"`
	Data    string  `json:"data"`
	Message message `json:"message"`
}

type answerCallback struct {
	ID   string `json:"callback_query_id"`
	Text string `json:"text,omitempty"`
}

type editBody struct {
	MessageID   int         `json:"message_id"`
	Text        string      `json:"text"`
	ChatID      string      `json:"chat_id"`
	ReplyMarkup replyMarkup `json:"reply_markup,omitempty"`
}

type keyboard struct {
	Buttons  []inlineKeyboard
	Keyboard [][]inlineKeyboard
}
