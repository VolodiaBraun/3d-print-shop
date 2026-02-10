package telegram

import (
	"context"
	"errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/brown/3d-print-shop/internal/domain"
)

func (b *Bot) handleCommand(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	cmd := update.Message.Command()

	b.log.Info("command received",
		zap.String("command", cmd),
		zap.Int64("chatID", chatID),
		zap.String("username", update.Message.From.UserName),
	)

	switch cmd {
	case "start":
		b.handleStart(chatID)
	case "help":
		b.handleHelp(chatID)
	case "orders":
		b.handleOrders(chatID, update.Message.From.ID)
	default:
		b.send(chatID, "Неизвестная команда. Отправьте /help для списка команд.")
	}
}

func (b *Bot) handleCallback(update tgbotapi.Update) {
	query := update.CallbackQuery

	// Acknowledge callback to remove loading spinner
	callback := tgbotapi.NewCallback(query.ID, "")
	if _, err := b.api.Request(callback); err != nil {
		b.log.Warn("failed to answer callback", zap.Error(err))
	}

	chatID := query.Message.Chat.ID
	data := query.Data

	b.log.Info("callback received",
		zap.String("data", data),
		zap.Int64("chatID", chatID),
	)

	switch data {
	case "start":
		b.handleStart(chatID)
	case "orders":
		b.handleOrders(chatID, query.From.ID)
	case "help":
		b.handleHelp(chatID)
	default:
		b.send(chatID, "Неизвестное действие.")
	}
}

func (b *Bot) handleStart(chatID int64) {
	kb := b.mainMenuKeyboard()
	b.sendWithKeyboard(chatID, welcomeText, kb)
}

func (b *Bot) handleHelp(chatID int64) {
	b.sendWithKeyboard(chatID, helpText, backToMenuButton())
}

func (b *Bot) handleOrders(chatID int64, telegramID int64) {
	ctx := context.Background()

	user, err := b.userRepo.FindByTelegramID(ctx, telegramID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			b.sendWithKeyboard(chatID, ordersNotLinkedText, backToMenuButton())
			return
		}
		b.log.Error("failed to find user by telegram id", zap.Error(err))
		b.send(chatID, "Произошла ошибка. Попробуйте позже.")
		return
	}

	orders, err := b.orderService.ListByUserID(ctx, user.ID)
	if err != nil {
		b.log.Error("failed to list orders", zap.Error(err))
		b.send(chatID, "Не удалось загрузить заказы. Попробуйте позже.")
		return
	}

	if len(orders) == 0 {
		b.sendWithKeyboard(chatID, noOrdersText, backToMenuButton())
		return
	}

	text := formatOrdersList(orders)
	b.sendWithKeyboard(chatID, text, backToMenuButton())
}
