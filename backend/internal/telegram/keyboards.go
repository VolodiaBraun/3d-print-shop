package telegram

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func (b *Bot) mainMenuKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(
				"\U0001F4E6 Каталог",
				b.webAppURL+"/catalog",
			),
			tgbotapi.NewInlineKeyboardButtonURL(
				"\U0001F6D2 Корзина",
				b.webAppURL+"/cart",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("\U0001F4CB Мои заказы", "orders"),
			tgbotapi.NewInlineKeyboardButtonData("\u2139\uFE0F Помощь", "help"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(
				"\U0001F3E0 Открыть магазин",
				b.webAppURL,
			),
		),
	)
}

func backToMenuButton() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("\u2B05 Главное меню", "start"),
		),
	)
}
