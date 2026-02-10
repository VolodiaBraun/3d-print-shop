package telegram

import (
	"encoding/json"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// WebApp types not available in tgbotapi v5.5.1

type webAppInfo struct {
	URL string `json:"url"`
}

type inlineButton struct {
	Text         string      `json:"text"`
	URL          string      `json:"url,omitempty"`
	CallbackData string      `json:"callback_data,omitempty"`
	WebApp       *webAppInfo `json:"web_app,omitempty"`
}

type inlineKeyboard struct {
	InlineKeyboard [][]inlineButton `json:"inline_keyboard"`
}

func (b *Bot) setMenuButton() {
	menuBtn, _ := json.Marshal(map[string]interface{}{
		"type":    "web_app",
		"text":    "Магазин",
		"web_app": map[string]string{"url": b.webAppURL},
	})
	params := make(tgbotapi.Params)
	params["menu_button"] = string(menuBtn)
	if _, err := b.api.MakeRequest("setChatMenuButton", params); err != nil {
		b.log.Warn("failed to set menu button")
	}
}

func (b *Bot) mainMenuKeyboard() inlineKeyboard {
	return inlineKeyboard{
		InlineKeyboard: [][]inlineButton{
			{
				{Text: "\U0001F4E6 Каталог", WebApp: &webAppInfo{URL: b.webAppURL + "/catalog"}},
				{Text: "\U0001F6D2 Корзина", WebApp: &webAppInfo{URL: b.webAppURL + "/cart"}},
			},
			{
				{Text: "\U0001F4CB Мои заказы", CallbackData: "orders"},
				{Text: "\u2139\uFE0F Помощь", CallbackData: "help"},
			},
			{
				{Text: "\U0001F3E0 Открыть магазин", WebApp: &webAppInfo{URL: b.webAppURL}},
			},
		},
	}
}

func backToMenuKeyboard() inlineKeyboard {
	return inlineKeyboard{
		InlineKeyboard: [][]inlineButton{
			{
				{Text: "\u2B05 Главное меню", CallbackData: "start"},
			},
		},
	}
}
