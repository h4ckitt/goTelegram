package goTelegram

// Bot : Main Bot Struct
type Bot struct {
	Me              user `json:"result"`
	APIURL          string
	handler         func(Update)
	handlerSet      bool
	keyboardManager *keyboardManager
}

type user struct {
	ID        int    `json:"id"`
	Firstname string `json:"first_name"`
	Username  string `json:"username"`
}

// Update : Stores Data From Request
type Update struct {
	UpdateID      int           `json:"update_id"`
	EditedMessage Message       `json:"edited_message"`
	Message       Message       `json:"message"`
	CallbackQuery callbackQuery `json:"callback_query"`
	Command       string
	Type          string
}

type result struct {
	File fileDets `json:"result"`
}

type fileDets struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	FileSize     int    `json:"file_size"`
	FilePath     string `json:"file_path"`
}

type Message struct {
	MessageID int         `json:"message_id"`
	Text      string      `json:"Text"`
	Chat      Chat        `json:"chat"`
	From      user        `json:"from"`
	File      document    `json:"document"`
	Photo     []photoSize `json:"photo"`
	Video     video       `json:"video"`
}

type document struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	FileName     string `json:"file_name"`
	FileSize     int    `json:"file_size"`
}

type photoSize struct {
	document
	Width  int `json:"width"`
	Height int `json:"height"`
}

type video struct {
	photoSize
	Duration int `json:"duration"`
}

type InputMedia struct {
	Type              string `json:"type"`
	Media             string `json:"media"`
	Caption           string `json:"caption,omitempty"`
	HasSpoiler        bool   `json:"has_spoiler,omitempty"`
	Width             int    `json:"width,omitempty"`
	Height            int    `json:"height,omitempty"`
	Duration          int    `json:"duration,omitempty"`
	SupportsStreaming bool   `json:"supports_streaming,omitempty"`
}

type mediaGroup struct {
	ChatID              string          `json:"chat_id"`
	Media               []InputMedia    `json:"media"`
	DisableNotification bool            `json:"disable_notification,omitempty"`
	ProtectContent      bool            `json:"disable_content_type_detection,omitempty"`
	ReplyParameters     replyParameters `json:"reply_to_message_id,omitempty"`
}

type MediaOptions struct {
	UseSpoiler       bool
	SendNotification bool `json:"disable_notification,omitempty"`
	ProtectContent   bool `json:"disable_content_type_detection,omitempty"`
}

type Chat struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}

// InlineKeyboard : Structure To Hold The Keyboard To Be Sent
type InlineKeyboard struct {
	Text string `json:"text"`
	Data string `json:"callback_data"`
}

type replyParameters struct {
	MessageID                int    `json:"message_id"`
	ChatID                   int    `json:"chat_id,omitempty"`
	AllowSendingWithoutReply bool   `json:"allow_sending_without_reply,omitempty"`
	Quote                    string `json:"quote,omitempty"`
	QuoteParseMode           string `json:"quote_parse_mode,omitempty"`
	QuotePosition            int    `json:"quote_position,omitempty"`
}

type replyBody struct {
	ChatID          string          `json:"chat_id,omitempty"`
	Text            string          `json:"text,omitempty"`
	ParseMode       string          `json:"parse_mode,omitempty"`
	ReplyMarkup     replyMarkup     `json:"reply_markup,omitempty"`
	ReplyParameters replyParameters `json:"reply_to_message_id,omitempty"`
}

type videoBody struct {
	ChatID         string      `json:"chat_id"`
	Video          interface{} `json:"video"`
	Caption        string      `json:"caption,omitempty"`
	HasSpoiler     bool        `json:"has_spoiler,omitempty"`
	ProtectContent bool        `json:"protect_content,omitempty"`
}

type replyMarkup struct {
	InlineKeyboard [][]InlineKeyboard `json:"inline_keyboard,omitempty"`
}

type callbackQuery struct {
	ID      string  `json:"id"`
	From    user    `json:"from"`
	Data    string  `json:"data"`
	Message Message `json:"message"`
}

type answerCallback struct {
	ID        string `json:"callback_query_id"`
	Text      string `json:"text,omitempty"`
	ShowAlert string `json:"show_alert"`
}

type editBody struct {
	MessageID   int         `json:"message_id"`
	Text        string      `json:"text"`
	ChatID      string      `json:"chat_id"`
	ReplyMarkup replyMarkup `json:"reply_markup,omitempty"`
}

type deleteBody struct {
	MessageID int    `json:"message_id"`
	ChatID    string `json:"chat_id"`
}

type keyboard struct {
	Buttons  []InlineKeyboard
	Keyboard [][]InlineKeyboard
}

type TResponse struct {
	Ok     bool   `json:"ok"`
	Result Result `json:"result"`
}

type Result struct {
	MessageId int  `json:"message_id"`
	From      user `json:"from"`
	Chat      Chat `json:"chat"`
}
