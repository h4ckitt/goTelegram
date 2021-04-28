package bot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func newBot(s string) Bot {
	return Bot{ApiUrl: "https://api.telegram.org/bot" + s}
}

func newInlineKeyboard() Keyboard {
	return Keyboard{Keyboard: [][]InlineKeyboard{}}
}

func (k *Keyboard) addButton(text, callback string) {
	k.Buttons = append(k.Buttons, InlineKeyboard{Text: text, Data: callback})
}

func (k *Keyboard) makeKeyboardRow() {
	k.Keyboard = append(k.Keyboard, k.Buttons)
	k.Buttons = nil
}

func (k *Keyboard) deleteKeyboard() {
	k.Buttons = nil
	k.Keyboard = nil
}

func (b Bot) answerCallback(callbackID string) {
	link := b.ApiUrl + "/answerCallbackQuery"

	answer := AnswerCallback{
		Id: callbackID,
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
		log.Println("Couldn't Answer CallBack Successfully, Status Code Not OK")
	}
}

func (b Bot) sendMessage(s string, c Chat, k Keyboard) {

	link := b.ApiUrl + "/sendMessage"

	reply := ReplyBody{
		ChatID: strconv.Itoa(c.ID),
		Text:   s,
	}

	if len(k.Keyboard) > 0 {
		reply.ReplyMarkup.InlineKeyboard = k.Keyboard
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
		log.Println("Message Wasn't Sent Successfully, Please Try Again")
	}
}

func (b Bot) editMessage(message Message, text string, k Keyboard) {
	link := b.ApiUrl + "/editMessageText"

	updatedText := EditBody{
		ChatID:    strconv.Itoa(message.Chat.ID),
		MessageID: message.MessageID,
		Text:      text,
	}

	if len(k.Keyboard) > 0 {
		updatedText.ReplyMarkup.InlineKeyboard = k.Keyboard
	}

	jsonBody, err := json.Marshal(updatedText)

	if err != nil {
		log.Println("There Was An Error Marshalling The Message")
		log.Println(err)
		return
	}

	resp, err := http.Post(link, "application/json", bytes.NewBuffer(jsonBody))

	if err != nil {
		log.println("Couldn't Communicate With Telegram Servers, Please Check Internet Source")
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Message Wasn't Edited Successfully, Please Try Again")
		log.Println(err)
	}
}
