package bot

import "reflect"

type Bot struct {
	Me         User
	ApiUrl     string
	Handler    reflect.Value
	HandlerSet bool
}

type User struct {
	Id        int    `json:"id"`
	Firstname string `json:"first_name"`
	Username  string `json:"username"`
}

type Update struct {
	UpdateID      int           `json:"update_id"`
	Message       Message       `json:"message"`
	CallbackQuery CallbackQuery `json:"callback_query"`
}

type Message struct {
	MessageID int    `json:"message_id"`
	Text      string `json:"Text"`
	Chat      Chat   `json:"chat"`
	From      User   `json:"from"`
}

type Chat struct {
	ID int `json:"id"`
}

type InlineKeyboard struct {
	Text string `json:"text"`
	Data string `json:"callback_data"`
}

type ReplyBody struct {
	ChatID      string      `json:"chat_id,omitempty"`
	Text        string      `json:"text,omitempty"`
	ParseMode   string      `json:"parse_mode,omitempty"`
	ReplyMarkup ReplyMarkup `json:"reply_markup,omitempty"`
}

type ReplyMarkup struct {
	InlineKeyboard [][]InlineKeyboard `json:"inline_keyboard,omitempty"`
}

type CallbackQuery struct {
	Id      string  `json:"id"`
	From    User    `json:"from"`
	Data    string  `json:"data"`
	Message Message `json:"message"`
}

type AnswerCallback struct {
	Id   string `json:"callback_query_id"`
	Text string `json:"text,omitempty"`
}

type EditBody struct {
	MessageID   int         `json:"message_id"`
	Text        string      `json:"text"`
	ChatID      string      `json:"chat_id"`
	ReplyMarkup ReplyMarkup `json:"reply_markup,omitempty"`
}

type Keyboard struct {
	Buttons  []InlineKeyboard
	Keyboard [][]InlineKeyboard
}
