package telegram

import (
	"encoding/json"
	"fmt"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/brown/3d-print-shop/internal/config"
	"github.com/brown/3d-print-shop/internal/domain"
	"github.com/brown/3d-print-shop/internal/service"
)

// Bot handles Telegram bot interactions via webhook.
type Bot struct {
	api          *tgbotapi.BotAPI
	orderService *service.OrderService
	userRepo     domain.UserRepository
	webAppURL    string
	adminChatID  int64
	log          *zap.Logger
}

// New creates a new Telegram bot, sets the webhook, and registers bot commands.
func New(
	cfg config.TelegramConfig,
	orderService *service.OrderService,
	userRepo domain.UserRepository,
	log *zap.Logger,
) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("create bot api: %w", err)
	}

	bot := &Bot{
		api:          api,
		orderService: orderService,
		userRepo:     userRepo,
		webAppURL:    cfg.WebAppURL,
		adminChatID:  cfg.AdminChatID,
		log:          log.Named("telegram"),
	}

	// Set webhook
	if cfg.WebhookURL != "" {
		wh, _ := tgbotapi.NewWebhook(cfg.WebhookURL)
		if _, err := api.Request(wh); err != nil {
			return nil, fmt.Errorf("set webhook: %w", err)
		}
		log.Info("telegram webhook set", zap.String("url", cfg.WebhookURL))
	}

	// Register bot commands menu
	commands := tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{Command: "start", Description: "Главное меню"},
		tgbotapi.BotCommand{Command: "orders", Description: "Мои заказы"},
		tgbotapi.BotCommand{Command: "help", Description: "Помощь и контакты"},
	)
	if _, err := api.Request(commands); err != nil {
		log.Warn("failed to set bot commands", zap.Error(err))
	}

	// Set menu button to open Web App
	bot.setMenuButton()

	return bot, nil
}

// Username returns the bot's Telegram username.
func (b *Bot) Username() string {
	return b.api.Self.UserName
}

// RegisterWebhook adds the webhook POST route to the Gin router.
func (b *Bot) RegisterWebhook(router *gin.Engine) {
	router.POST("/webhook/telegram", b.handleWebhook)
}

// Stop removes the webhook on shutdown.
func (b *Bot) Stop() {
	if _, err := b.api.Request(tgbotapi.DeleteWebhookConfig{}); err != nil {
		b.log.Warn("failed to delete webhook", zap.Error(err))
	}
	b.log.Info("telegram bot stopped")
}

func (b *Bot) handleWebhook(c *gin.Context) {
	var update tgbotapi.Update
	if err := c.ShouldBindJSON(&update); err != nil {
		b.log.Warn("invalid telegram update", zap.Error(err))
		c.Status(http.StatusOK)
		return
	}

	b.processUpdate(update)
	c.Status(http.StatusOK)
}

func (b *Bot) processUpdate(update tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			b.log.Error("panic in telegram handler", zap.Any("recover", r))
		}
	}()

	if update.Message != nil && update.Message.IsCommand() {
		b.handleCommand(update)
	} else if update.CallbackQuery != nil {
		b.handleCallback(update)
	}
}

func (b *Bot) send(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	if _, err := b.api.Send(msg); err != nil {
		b.log.Error("failed to send message", zap.Int64("chatID", chatID), zap.Error(err))
	}
}

func (b *Bot) sendWithKeyboard(chatID int64, text string, keyboard inlineKeyboard) {
	// Use raw params to support web_app buttons not available in tgbotapi v5.5.1
	replyMarkup, _ := json.Marshal(keyboard)

	params := make(tgbotapi.Params)
	params.AddNonZero64("chat_id", chatID)
	params["text"] = text
	params["parse_mode"] = "HTML"
	params["reply_markup"] = string(replyMarkup)

	if _, err := b.api.MakeRequest("sendMessage", params); err != nil {
		b.log.Error("failed to send message", zap.Int64("chatID", chatID), zap.Error(err))
	}
}
