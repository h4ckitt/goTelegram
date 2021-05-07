package goTelegram

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

//NewBot : Create A New Bot
func NewBot(s string) (Bot, error) {

	var newBot Bot

	newBot.APIURL = "https://api.telegram.org/bot" + s

	newBot.Keyboard = keyboard{Keyboard: [][]InlineKeyboard{}}

	resp, err := http.Get(newBot.APIURL + "/getMe")

	if err != nil {
		log.Println("Fetch Bot Details Failed, Check Internet Connection")
		log.Println(err)
		return newBot, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Invalid Token Provided")
		return newBot, errors.New("Invalid Bot Token Provided")
	}

	err = json.NewDecoder(resp.Body).Decode(&newBot)

	if err != nil {
		log.Println("Couldn't Marshal Response")
		log.Println(err)
		return newBot, err
	}

	return newBot, nil

	//return Bot{APIURL: "https://api.telegram.org/bot" + s, Keyboard: keyboard{Keyboard: [][]InlineKeyboard{}}}
}

//AddButton : Add Buttons For InlineKeyboard
func (b *Bot) AddButton(text, callback string) {
	b.Keyboard.Buttons = append(b.Keyboard.Buttons, InlineKeyboard{Text: text, Data: callback})
}

//MakeKeyboard : Create Final Keyboard To Be Sent
func (b *Bot) MakeKeyboard(maxCol int) {

	buttons := b.Keyboard.Buttons

	b.Keyboard.Buttons = nil

	if maxCol < 1 {
		log.Println("Maximum Number Of Columns Cannot Be Less Than 1")
		return
	}

	for index, button := range buttons {

		if (index+1)%maxCol != 0 {
			b.Keyboard.Buttons = append(b.Keyboard.Buttons, button)
		} else {
			b.Keyboard.Buttons = append(b.Keyboard.Buttons, button)
			b.Keyboard.Keyboard = append(b.Keyboard.Keyboard, b.Keyboard.Buttons)
			b.Keyboard.Buttons = nil
		}
	}

	if len(b.Keyboard.Buttons) > 0 {
		b.Keyboard.Keyboard = append(b.Keyboard.Keyboard, b.Keyboard.Buttons)
	}
}

//DeleteKeyboard : Delete Current Keyboard
func (b *Bot) DeleteKeyboard() {
	b.Keyboard.Keyboard = nil
	b.Keyboard.Buttons = nil
}

//SetHandler : Set Function To Be Run When New Updates Are Received
func (b *Bot) SetHandler(fn interface{}) {
	b.HandlerSet = false
	b.Handler = reflect.ValueOf(fn)
	if b.Handler.Kind() != reflect.Func {
		log.Println("Argument Is Not Of Type Function")
		return
	}

	b.HandlerSet = true
}

//UpdateHandler : Handles New Updates From Telegram
func (b *Bot) UpdateHandler(w http.ResponseWriter, r *http.Request) {

	if b.HandlerSet {

		var update Update
		var text []string

		err := json.NewDecoder(r.Body).Decode(&update)

		if err != nil {
			log.Println("Couldn't Parse Incoming Message")
			return
		}

		if len(update.EditedMessage.Text) > 0 {
			update.Type = "edited_text"
			text = strings.Fields(update.EditedMessage.Text)
		} else if len(update.Message.Text) > 0 {
			update.Type = "text"
			text = strings.Fields(update.Message.Text)
		} else if update.CallbackQuery.ID != "" {

			update.Type = "callback"
		}

		if len(text) > 0 {

			if strings.HasPrefix(text[0], "/") {

				update.Command = text[0]

				if strings.HasSuffix(text[0], b.Me.Username) {
					update.Command = strings.Split(text[0], "@")[0]
				}

			}
		}

		rarg := make([]reflect.Value, 1)

		rarg[0] = reflect.ValueOf(update)

		go b.Handler.Call(rarg)
	} else {
		log.Println("Please Set A Function To Be Called Upon New Updates")
		return
	}

}

//AnswerCallback : Answer Call Back Query From InlineKeyboard
func (b *Bot) AnswerCallback(callbackID string) {
	link := b.APIURL + "/answerCallbackQuery"

	answer := answerCallback{
		ID: callbackID,
	}

	jsonBody, err := json.Marshal(answer)

	if err != nil {
		log.Println(err)
		return
	}

	resp, err := http.Post(link, "application/json", bytes.NewBuffer(jsonBody))

	if err != nil {
		log.Println("Couldn't Answer CallBack Successfully, Check Internet Source")
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		//	body, _ := ioutil.ReadAll(resp.Body)
		log.Println("Couldn't Answer CallBack Successfully, Status Code Not OK")
		//	log.Println(string(body))
	}
}

//SendMessage : Send A Message To A User
func (b *Bot) SendMessage(s string, c chat) {

	link := b.APIURL + "/sendMessage"

	reply := replyBody{
		ChatID: strconv.Itoa(c.ID),
		Text:   s,
	}

	if len(b.Keyboard.Keyboard) > 0 {
		reply.ReplyMarkup.InlineKeyboard = b.Keyboard.Keyboard
	}

	jsonBody, err := json.Marshal(reply)

	if err != nil {
		log.Println("Couldn't Marshal Response")
		return
	}

	resp, err := http.Post(link, "application/json", bytes.NewBuffer(jsonBody))

	if err != nil {
		log.Println("Couldn't Make Request Successfully, Please Check Internet Source")
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		//	body, _ := ioutil.ReadAll(resp.Body)
		log.Println("Message Wasn't Sent Successfully, Please Try Again")
		//	log.Println(string(body))
	}
}

//EditMessage : Edit An Existing Message
func (b *Bot) EditMessage(message message, text string) {
	link := b.APIURL + "/editMessageText"

	updatedText := editBody{
		ChatID:    strconv.Itoa(message.Chat.ID),
		MessageID: message.MessageID,
		Text:      text,
	}

	if len(b.Keyboard.Keyboard) > 0 {
		updatedText.ReplyMarkup.InlineKeyboard = b.Keyboard.Keyboard
	}

	jsonBody, err := json.Marshal(updatedText)

	if err != nil {
		log.Println("There Was An Error Marshalling The Message")
		log.Println(err)
		return
	}

	resp, err := http.Post(link, "application/json", bytes.NewBuffer(jsonBody))

	if err != nil {
		log.Println("Couldn't Communicate With Telegram Servers, Please Check Internet Source")
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Message Wasn't Edited Successfully, Please Try Again")
	}
}

//DeleteMessage : Delete The Specified Message
func (b *Bot) DeleteMessage(message message) {
	link := b.APIURL + "/deleteMessage"

	deletion := deleteBody{
		MessageID: message.MessageID,
		ChatID:    strconv.Itoa(message.Chat.ID),
	}

	jsonBody, err := json.Marshal(deletion)

	if err != nil {
		log.Println("There Was An Erro Marshalling The Object")
		log.Println(err)
		return
	}

	resp, err := http.Post(link, "application/json", bytes.NewBuffer(jsonBody))

	if err != nil {
		log.Println("Couldn't Complete Message Deletion Request, Please Check Internet Source")
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("Message Couldn't Be Deleted Successfully")
		log.Println(string(body))
	}
}
