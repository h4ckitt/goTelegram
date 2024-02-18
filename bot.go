package goTelegram

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// NewBot : Create A New Bot
func NewBot(s string) (Bot, error) {

	var newBot Bot

	newBot.APIURL = "https://api.telegram.org/bot" + s

	newBot.keyboardManager = newKeyboardManager()

	resp, err := http.Get(newBot.APIURL + "/getMe")

	if err != nil {
		log.Println("Fetch Bot Details Failed, Check Internet Connection")
		return newBot, err
	}

	defer func() { _ = resp.Body.Close() }()

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
}

func (b *Bot) CreateKeyboard(chatId int, maxColumns ...int) {
	maxCols := 3

	if len(maxColumns) > 0 {
		maxCols = maxColumns[0]
	}

	b.keyboardManager.CreateKeyboard(chatId, maxCols)
}

// AddButtons : Add Buttons For InlineKeyboard
func (b *Bot) AddButtons(chatID int, buttonData ...string) error {
	if len(buttonData)%2 != 0 {
		return errors.New("invalid Number Of Parameters Passed, It Should Be In Teh Format (buttonData, data, buttonData, data)")
	}

	for i := 0; i < len(buttonData); i += 2 {
		b.keyboardManager.AddButton(chatID, buttonData[i], buttonData[i+1])
	}

	return nil
}

// DeleteKeyboard : Delete Current Keyboard
func (b *Bot) DeleteKeyboard(chatID int) {
	b.keyboardManager.DeleteKeyboard(chatID)
}

// SetHandler : Set Function To Be Run When New Updates Are Received
func (b *Bot) SetHandler(fn func(Update)) bool {
	b.handler = fn
	b.handlerSet = true
	return true
}

// UpdateHandler : Handles New Updates From Telegram
func (b *Bot) UpdateHandler(_ http.ResponseWriter, r *http.Request) {

	if b.handlerSet {

		var update Update
		var text []string

		err := json.NewDecoder(r.Body).Decode(&update)

		if err != nil {
			log.Println("Couldn't Parse Incoming Message")
			return
		}

		switch {
		case len(update.EditedMessage.Text) > 0:
			update.Type = "edited_text"
			text = strings.Fields(update.EditedMessage.Text)

		case len(update.Message.Text) > 0:
			update.Type = "text"
			text = strings.Fields(update.Message.Text)

		case len(update.CallbackQuery.ID) > 0:
			update.Type = "callback"

		case len(update.Message.File.FileName) > 0:
			update.Type = "document"

		case len(update.Message.Photo) > 0 && len(update.Message.Video.FileID) > 0:
			update.Type = "media_group"

		case len(update.Message.Photo) > 0:
			update.Type = "photo"

		case len(update.Message.Video.FileID) > 0:
			update.Type = "video"

		default:
			update.Type = "unknown"
		}

		if len(text) > 0 {

			if strings.HasPrefix(text[0], "/") {

				update.Command = text[0]

				if strings.HasSuffix(text[0], b.Me.Username) {
					update.Command = strings.Split(text[0], "@")[0]
				}

			}
		}

		go b.handler(update)
	} else {
		log.Println("Please Set A Function To Be Called Upon New Updates")
		return
	}

}

// AnswerCallback : Answer Call Back Query From InlineKeyboard
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

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Println("Couldn't Answer CallBack Successfully, Status Code Not OK")
		return errors.New(string(body))
	}

	return nil
}

// SendMessage : Send A Message To A User
func (b *Bot) SendMessage(s string, c Chat) (Message, error) {

	link := b.APIURL + "/sendMessage"

	reply := replyBody{
		ChatID: strconv.Itoa(c.ID),
		Text:   s,
	}

	if b.keyboardManager.HasKeyboard(c.ID) {
		reply.ReplyMarkup.InlineKeyboard = b.keyboardManager.ReturnKeyboard(c.ID)
	}

	jsonBody, err := json.Marshal(reply)

	if err != nil { // _, err := bot.SendMessage("Hello "+update.Message.From.Firstname, update.Message.Chat)
		//
		// if err != nil {
		// 	log.Println(err)
		// }
		log.Println("Couldn't Marshal Response")
		return Message{}, err

	}

	resp, err := http.Post(link, "application/json", bytes.NewBuffer(jsonBody))

	if err != nil {
		body, _ := io.ReadAll(resp.Body)
		log.Println("Couldn't Make Request Successfully, Please Check Internet Source")
		return Message{}, errors.New(string(body))
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Println("Message Wasn't Sent Successfully, Please Try Again")
		return Message{}, errors.New(string(body))
	}

	var response TResponse
	var newMessage Message

	err = json.NewDecoder(resp.Body).Decode(&response)

	if err != nil {
		return Message{}, err
	}

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
		ReplyParameters: replyParameters{
			MessageID: m.MessageID,
		},
	}

	if b.keyboardManager.HasKeyboard(m.Chat.ID) {
		reply.ReplyMarkup.InlineKeyboard = b.keyboardManager.ReturnKeyboard(m.Chat.ID)
	}

	jsonBody, err := json.Marshal(reply)

	if err != nil {
		log.Println("Couldn't Marshal Response")
		return err
	}

	resp, err := http.Post(link, "application/json", bytes.NewBuffer(jsonBody))

	if err != nil {
		body, _ := io.ReadAll(resp.Body)
		log.Println("Couldn't Make Request Successfully, Please Check Internet Source")
		return errors.New(string(body))
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Println("Message Wasn't Sent Successfully, Please Try Again")
		return errors.New(string(body))
	}

	return nil
}

func (b *Bot) DownloadFile(fileId, filename string) error {
	splitLink := strings.Split(b.APIURL, "bot")

	fileDetails, err := b.fetchFileDetails(fileId)
	if err != nil {
		return err
	}

	file := fileDetails.File
	resp, err := http.Get(splitLink[0] + "/file/bot" + splitLink[1] + "/" + file.FilePath)

	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return errors.New(string(body))
	}

	fileName, err := os.Create(filename)

	if err != nil {
		return err
	}

	defer func() { _ = fileName.Close() }()

	_, err = io.Copy(fileName, resp.Body)

	if err != nil {
		return err
	}

	return nil
}

func (b *Bot) DownloadFileToMemory(fileId string) ([]byte, error) {
	buff := new(bytes.Buffer)
	splitLink := strings.Split(b.APIURL, "bot")

	fileDetails, err := b.fetchFileDetails(fileId)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/file/bot%s/%s", splitLink[0], splitLink[1], fileDetails.File.FilePath)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.New(string(body))
	}

	_, err = io.Copy(buff, resp.Body)

	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

// EditMessage : Edit An Existing Message
func (b *Bot) EditMessage(m Message, text string) (Message, error) {

	link := b.APIURL + "/editMessageText"

	updatedText := editBody{
		ChatID:    strconv.Itoa(m.Chat.ID),
		MessageID: m.MessageID,
		Text:      text,
	}

	if b.keyboardManager.HasKeyboard(m.Chat.ID) {
		updatedText.ReplyMarkup.InlineKeyboard = b.keyboardManager.ReturnKeyboard(m.Chat.ID)
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
		body, _ := io.ReadAll(resp.Body)
		return Message{}, errors.New(string(body))
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
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

	defer func() { _ = resp.Body.Close() }()

	return newMessage, nil
}

// DeleteMessage : Delete The Specified Message
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
		body, _ := io.ReadAll(resp.Body)
		return errors.New(string(body))
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Println("Message Wasn't Deleted Successfully, Please Try Again")
		return errors.New(string(body))
	}

	return nil
}

func (b *Bot) SendVideoFromMemory(data []byte, caption string, c Chat, options MediaOptions) error {
	link := b.APIURL + "/sendVideo"

	buffer := bytes.NewReader(data)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("video", "video.mp4")
	if err != nil {
		return err
	}

	_, err = io.Copy(part, buffer)
	if err != nil {
		return err
	}

	_ = writer.WriteField("chat_id", strconv.Itoa(c.ID))

	if caption != "" {
		_ = writer.WriteField("caption", caption)
	}

	if options.UseSpoiler {
		_ = writer.WriteField("has_spoiler", "true")
	}

	if options.ProtectContent {
		_ = writer.WriteField("protect_content", "true")
	}

	_ = writer.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("POST", link, body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)

	if err != nil {
		log.Println("Couldn't Send Photo, Check Internet Connection")
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		log.Println("Photo Not Sent Successfully, Check Error Logs For Details")
		body, _ := io.ReadAll(resp.Body)
		return errors.New(string(body))
	}
	return nil

}

func (b *Bot) SendVideo(file string, caption string, c Chat, options MediaOptions) error {

	link := b.APIURL + "/sendVideo"

	if regexp.MustCompile("^(https?)").MatchString(file) || file[:len(file)-4] != ".mp4" {
		body := videoBody{
			ChatID:  strconv.Itoa(c.ID),
			Video:   file,
			Caption: caption,
		}

		if options.UseSpoiler {
			body.HasSpoiler = true
		}

		if options.ProtectContent {
			body.ProtectContent = true
		}

		jsonBody, err := json.Marshal(body)

		if err != nil {
			log.Println("There Was An Error Marshalling The Object")
			return err
		}

		resp, err := http.Post(link, "application/json", bytes.NewBuffer(jsonBody))

		if err != nil {
			log.Println("Video Couldn't Be Sent Successfully, Please Check Internet Source")
			return err
		}

		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			log.Println("Video Not Sent Successfully, Check Error Logs For More Details")
			body, _ := io.ReadAll(resp.Body)
			return errors.New(string(body))
		}

		return nil
	}
	vid, err := os.Open(file)

	if err != nil {
		log.Println("Couldn't Open Specified File For Reading")
		return err
	}

	defer func() { _ = vid.Close() }()

	body := new(bytes.Buffer)

	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("video", filepath.Base(file))

	if err != nil {
		return err
	}

	_, err = io.Copy(part, vid)

	if err != nil {
		return err
	}

	_ = writer.WriteField("chat_id", strconv.Itoa(c.ID))

	if caption != "" {
		_ = writer.WriteField("caption", caption)
	}

	if options.UseSpoiler {
		_ = writer.WriteField("has_spoiler", "true")
	}

	if options.ProtectContent {
		_ = writer.WriteField("protect_content", "true")
	}

	_ = writer.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("POST", link, body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)

	if err != nil {
		log.Println("Couldn't Send Video, Check Internet Connection")
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		log.Println("Video Not Sent Successfully, Check Error Logs For Details")
		body, _ := io.ReadAll(resp.Body)
		return errors.New(string(body))
	}
	return nil

}

func (b *Bot) SendPhotoFromMemory(data []byte, caption string, c Chat, options MediaOptions) error {
	link := b.APIURL + "/sendPhoto"

	buffer := bytes.NewReader(data)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("photo", "photo.jpg")
	if err != nil {
		return err
	}

	_, err = io.Copy(part, buffer)
	if err != nil {
		return err
	}

	_ = writer.WriteField("chat_id", strconv.Itoa(c.ID))

	if caption != "" {
		_ = writer.WriteField("caption", caption)
	}

	if options.UseSpoiler {
		_ = writer.WriteField("has_spoiler", "true")
	}

	if options.ProtectContent {
		_ = writer.WriteField("protect_content", "true")
	}

	_ = writer.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("POST", link, body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)

	if err != nil {
		log.Println("Couldn't Send Photo, Check Internet Connection")
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		log.Println("Photo Not Sent Successfully, Check Error Logs For Details")
		body, _ := io.ReadAll(resp.Body)
		return errors.New(string(body))
	}
	return nil

}

func (b *Bot) SendPhoto(file string, caption string, c Chat, options MediaOptions) error {

	link := b.APIURL + "/sendPhoto"

	if regexp.MustCompile("^(https?)").MatchString(file) || file[:len(file)-4] != ".jpg" && file[:len(file)-4] != ".png" && file[:len(file)-4] != ".jpeg" {
		body := videoBody{
			ChatID:  strconv.Itoa(c.ID),
			Video:   file,
			Caption: caption,
		}

		if options.UseSpoiler {
			body.HasSpoiler = true
		}

		if options.ProtectContent {
			body.ProtectContent = true
		}

		jsonBody, err := json.Marshal(body)

		if err != nil {
			log.Println("There Was An Error Marshalling The Object")
			return err
		}

		resp, err := http.Post(link, "application/json", bytes.NewBuffer(jsonBody))

		if err != nil {
			log.Println("Photo Couldn't Be Sent Successfully, Please Check Internet Source")
			return err
		}

		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			log.Println("Photo Not Sent Successfully, Check Error Logs For More Details")
			body, _ := io.ReadAll(resp.Body)
			return errors.New(string(body))
		}

		return nil
	}

	photo, err := os.Open(file)

	if err != nil {
		log.Println("Couldn't Open Specified File For Reading")
		return err
	}

	defer func() { _ = photo.Close() }()

	body := new(bytes.Buffer)

	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("photo", filepath.Base(file))

	if err != nil {
		return err
	}

	_, err = io.Copy(part, photo)

	if err != nil {
		return err
	}

	_ = writer.WriteField("chat_id", strconv.Itoa(c.ID))

	if caption != "" {
		_ = writer.WriteField("caption", caption)
	}

	if options.UseSpoiler {
		_ = writer.WriteField("has_spoiler", "true")
	}

	if options.ProtectContent {
		_ = writer.WriteField("protect_content", "true")
	}

	_ = writer.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("POST", link, body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)

	if err != nil {
		log.Println("Couldn't Send Photo, Check Internet Connection")
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		log.Println("Photo Not Sent Successfully, Check Error Logs For Details")
		body, _ := io.ReadAll(resp.Body)
		return errors.New(string(body))
	}
	return nil
}

// SendMediaGroup : Send A Group Of Media Files
// It Only Works With Local Files For Now
func (b *Bot) SendMediaGroup(files []InputMedia, c Chat, options MediaOptions) error {

	link := b.APIURL + "/sendMediaGroup"

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for _, file := range files {
		if m, err := regexp.MatchString(`^(https?)`, file.Media); err != nil || !m {
			if err != nil {
				log.Println("There Was An Error Checking The Media Link")
				continue
			}

			photo, err := os.Open(file.Media)

			if err != nil {
				log.Println("Couldn't Open Specified File For Reading")
				continue
			}

			file.Media = "attach://" + filepath.Base(file.Media)
			part, err := writer.CreateFormFile(filepath.Base(file.Media), file.Media)
			if err != nil {
				log.Println("There Was An Error Creating The Form File")
				continue
			}

			_, err = io.Copy(part, photo)

			if err != nil {
				log.Println("There Was An Error Copying The File")
				continue
			}

			_ = photo.Close()
		}
	}

	jsonBody, err := json.Marshal(files)

	if err != nil {
		log.Println("There Was An Error Marshalling The Object")
		return err
	}

	_ = writer.WriteField("chat_id", strconv.Itoa(c.ID))
	_ = writer.WriteField("disable_notification", strconv.FormatBool(!options.SendNotification))
	_ = writer.WriteField("disable_content_type_detection", strconv.FormatBool(options.ProtectContent))
	_ = writer.WriteField("media", string(jsonBody))

	/*mg := mediaGroup{
		ChatID:              strconv.Itoa(c.ID),
		Media:               files,
		DisableNotification: !options.SendNotification,
		ProtectContent:      options.ProtectContent,
	}*/

	client := &http.Client{}
	req, _ := http.NewRequest("POST", link, body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)

	if err != nil {
		log.Println("Media Group Couldn't Be Sent Successfully, Please Check Internet Source")
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		log.Println("Media Group Not Sent Successfully, Check Error Logs For More Details")
		body, _ := io.ReadAll(resp.Body)
		return errors.New(string(body))
	}

	return nil
}

func (b *Bot) fetchFileDetails(fileId string) (*result, error) {
	var res result

	link := b.APIURL + "/getFile"
	jsonBody, err := json.Marshal(struct {
		FileID string `json:"file_id"`
	}{
		fileId,
	})

	resp, err := http.Post(link, "application/json", bytes.NewBuffer(jsonBody))

	if err != nil {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.New(string(body))
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.New(string(body))
	}

	defer func() { _ = resp.Body.Close() }()

	err = json.NewDecoder(resp.Body).Decode(&res)

	if err != nil {
		return nil, err
	}

	return &res, nil

}
