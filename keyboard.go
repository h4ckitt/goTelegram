package goTelegram

type keyboardManager struct {
	keyboards map[int]*chatKeyboard
}

type chatKeyboard struct {
	maxColumns int
	keyboard   keyboard
}

func newKeyboardManager() *keyboardManager {
	return &keyboardManager{make(map[int]*chatKeyboard)}
}

func (k *keyboardManager) CreateKeyboard(chatID, maxCol int) {
	_, exists := k.keyboards[chatID]

	if !exists {
		k.keyboards[chatID] = &chatKeyboard{
			maxColumns: maxCol,
			keyboard:   keyboard{Keyboard: [][]InlineKeyboard{}},
		}
	}
}

func (k *keyboardManager) HasKeyboard(chatID int) bool {
	chatKbd, exists := k.keyboards[chatID]

	if exists {
		return len(chatKbd.keyboard.Buttons) > 0
	}

	return false
}

func (k *keyboardManager) DeleteKeyboard(chatID int) {
	delete(k.keyboards, chatID)
}

func (k *keyboardManager) AddButton(chatID int, text, callBack string) {
	chatKbd, exists := k.keyboards[chatID]

	if exists {
		chatKbd.keyboard.Buttons = append(chatKbd.keyboard.Buttons, InlineKeyboard{Text: text, Data: callBack})
	}
}

func (k *keyboardManager) ClearKeyboard(chatID int) {
	chatKbd, exists := k.keyboards[chatID]

	if exists {
		chatKbd.keyboard.Keyboard = nil
		chatKbd.keyboard.Buttons = nil
	}
}

func (k *keyboardManager) ReturnKeyboard(chatID int) [][]InlineKeyboard {
	chatKbd, exists := k.keyboards[chatID]

	if len(chatKbd.keyboard.Keyboard) > 0 {
		return chatKbd.keyboard.Keyboard
	}

	if !exists {
		return nil
	}

	buttons := make([]InlineKeyboard, 0)

	for index, button := range chatKbd.keyboard.Buttons {
		if (index+1)%chatKbd.maxColumns == 0 {
			buttons = append(buttons, button)
			chatKbd.keyboard.Keyboard = append(chatKbd.keyboard.Keyboard, buttons)
			buttons = nil
		} else {
			buttons = append(buttons, button)
		}
	}

	if len(buttons) > 0 {
		chatKbd.keyboard.Keyboard = append(chatKbd.keyboard.Keyboard, buttons)
	}

	return chatKbd.keyboard.Keyboard
}
