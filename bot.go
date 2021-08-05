package goTelegram

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
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
		return newBot, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Invalid Token Provided")
		return newBot, errors.New("invalid Bot Token Provided")
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
func (b *Bot) SetHandler(fn interface{}) bool {
	b.handlerSet = false
	b.handler = reflect.ValueOf(fn)
	if b.handler.Kind() != reflect.Func {
		log.Println("Argument Is Not Of Type Function")
		return false
	}

	b.handlerSet = true

	return true
}

//UpdateHandler : Handles New Updates From Telegram
func (b *Bot) UpdateHandler(w http.ResponseWriter, r *http.Request) {

	if b.handlerSet {

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
		} else if len(update.Message.File.FileName) > 0 {
			update.Type = "document"
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

		go b.handler.Call(rarg)
	} else {
		log.Println("Please Set A Function To Be Called Upon New Updates")
		return
	}

}

//AnswerCallback : Answer Call Back Query From InlineKeyboard
func (b *Bot) AnswerCallback(callbackID, text string, showAlert bool) error {
	link := b.APIURL + "/answerCallbackQuery"

	answer := answerCallback{
		ID: callbackID,
	}

	if text != "" {
		answer.Text = text
	}

	if showAlert {
		answer.ShowAlert = "true"
	} else {
		answer.ShowAlert = "false"
	}

	jsonBody, err := json.Marshal(answer)

	if err != nil {
		return err
	}

	resp, err := http.Post(link, "application/json", bytes.NewBuffer(jsonBody))

	if err != nil {
		log.Println("Couldn't Answer CallBack Successfully, Check Internet Source")
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("Couldn't Answer CallBack Successfully, Status Code Not OK")
		return errors.New(string(body))
	}

	return nil
}

//SendMessage : Send A Message To A User
func (b *Bot) SendMessage(s string, c Chat) (Message, error) {

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
		return Message{}, err
	}

	resp, err := http.Post(link, "application/json", bytes.NewBuffer(jsonBody))

	if err != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("Couldn't Make Request Successfully, Please Check Internet Source")
		return Message{}, errors.New(string(body))
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("Message Wasn't Sent Successfully, Please Try Again")
		return Message{}, errors.New(string(body))
	}

	var response TResponse
	var newMessage Message

	err = json.NewDecoder(resp.Body).Decode(&response)

	newMessage = Message{
		MessageID: response.Result.MessageId,
		Chat:      response.Result.Chat,
		From:      response.Result.From,
	}

	return newMessage, nil
}

func (b *Bot) ReplyMessage(s string, m Message) error {
	link := b.APIURL + "/sendMessage"

	reply := replyBody{
		ChatID: strconv.Itoa(m.Chat.ID),
		Text:   s,
		Reply:  m.MessageID,
	}

	if len(b.Keyboard.Keyboard) > 0 {
		reply.ReplyMarkup.InlineKeyboard = b.Keyboard.Keyboard
	}

	jsonBody, err := json.Marshal(reply)

	if err != nil {
		log.Println("Couldn't Marshal Response")
		return err
	}

	resp, err := http.Post(link, "application/json", bytes.NewBuffer(jsonBody))

	if err != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("Couldn't Make Request Successfully, Please Check Internet Source")
		return errors.New(string(body))
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("Message Wasn't Sent Successfully, Please Try Again")
		return errors.New(string(body))
	}

	return nil
}

func (b *Bot) GetFile(fileId, filename string) error {
	var res result

	//	link := strings.Join(strings.Split(b.APIURL, "bot")[0] + "/file/"
	splitLink := strings.Split(b.APIURL, "bot")
	link := b.APIURL + "/getFile"
	jsonBody, err := json.Marshal(struct {
		FileID string `json:"file_id"`
	}{
		fileId,
	})

	resp, err := http.Post(link, "application/json", bytes.NewBuffer(jsonBody))

	if err != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(body))
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&res)

	if err != nil {
		return err
	}

	file := res.File
	fmt.Println("File Path: ", file.FilePath)
	resp, err = http.Get(splitLink[0] + "/file/bot" + splitLink[1] + "/" + file.FilePath)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(body))
	}

	file_name, err := os.Create(filename)

	if err != nil {
		return err
	}

	defer file_name.Close()

	_, err = io.Copy(file_name, resp.Body)

	if err != nil {
		return err
	}

	return nil
}

//EditMessage : Edit An Existing Message
func (b *Bot) EditMessage(m Message, text string) (Message, error) {
	link := b.APIURL + "/editMessageText"

	updatedText := editBody{
		ChatID:    strconv.Itoa(m.Chat.ID),
		MessageID: m.MessageID,
		Text:      text,
	}

	if len(b.Keyboard.Keyboard) > 0 {
		updatedText.ReplyMarkup.InlineKeyboard = b.Keyboard.Keyboard
	}

	jsonBody, err := json.Marshal(updatedText)

	if err != nil {
		log.Println("There Was An Error Marshalling The Message")
		return Message{}, err
	}

	resp, err := http.Post(link, "application/json", bytes.NewBuffer(jsonBody))

	if err != nil {
		log.Println("Couldn't Communicate With Telegram Servers, Please Check Internet Source")
		log.Println(err)
		body, _ := ioutil.ReadAll(resp.Body)
		return Message{}, errors.New(string(body))
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("Message Wasn't Edited Successfully, Please Try Again")
		return Message{}, errors.New(string(body))
	}

	var response TResponse
	var newMessage Message

	err = json.NewDecoder(resp.Body).Decode(&response)

	newMessage = Message{
		MessageID: response.Result.MessageId,
		Chat:      response.Result.Chat,
		From:      response.Result.From,
	}

	defer resp.Body.Close()

	return newMessage, nil
}

//DeleteMessage : Delete The Specified Message
func (b *Bot) DeleteMessage(message Message) error {
	link := b.APIURL + "/deleteMessage"

	deletion := deleteBody{
		MessageID: message.MessageID,
		ChatID:    strconv.Itoa(message.Chat.ID),
	}

	jsonBody, err := json.Marshal(deletion)

	if err != nil {
		log.Println("There Was An Error Marshalling The Object")
		return err
	}

	resp, err := http.Post(link, "application/json", bytes.NewBuffer(jsonBody))

	if err != nil {
		log.Println("Couldn't Complete Message Deletion Request, Please Check Internet Source")
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(body))
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("Message Couldn't Be Deleted Successfully")
		return errors.New(string(body))
	}

	return nil
}

func (b *Bot) SendVideo(file string, caption string, c Chat) error {
	link := b.APIURL + "/sendVideo"

	if regexp.MustCompile("^(https?)").MatchString(file) {
		jsonBody, err := json.Marshal(videoBody{
			ChatID:    strconv.Itoa(c.ID),
			VideoLink: file,
			Caption:   caption,
		})

		if err != nil {
			log.Println("There Was An Error Marshalling The Object")
			return err
		}

		resp, err := http.Post(link, "application/json", bytes.NewBuffer(jsonBody))

		if err != nil {
			log.Println("Video Couldn't Be Sent Successfully, Please Check Internet Source")
			return err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Println("Video Not Sent Successfully, Check Error Logs For More Details")
			body, _ := ioutil.ReadAll(resp.Body)
			return errors.New(string(body))
		}

		return nil
	}

	vid, err := os.Open(file)

	if err != nil {
		log.Println("Couldn't Open Specified File For Reading")
		return err
	}

	defer vid.Close()

	body := new(bytes.Buffer)

	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("video", filepath.Base(file))

	io.Copy(part, vid)

	writer.WriteField("chat_id", strconv.Itoa(c.ID))

	if caption != "" {
		writer.WriteField("caption", caption)
	}

	writer.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("POST", link, body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)

	if err != nil {
		log.Println("Couldn't Send Video, Check Internet Connection")
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Video Not Sent Successfully, Check Error Logs For Details")
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(body))
	}
	return nil
}
