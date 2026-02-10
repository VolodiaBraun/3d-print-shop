package telegram

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/brown/3d-print-shop/internal/domain"
)

// NotifyOrderCreated sends a notification to the customer when an order is created.
func (b *Bot) NotifyOrderCreated(ctx context.Context, order *domain.Order) error {
	chatID, err := b.resolveUserChatID(ctx, order)
	if err != nil {
		return err
	}
	if chatID == 0 {
		return nil
	}

	text := fmt.Sprintf(
		"\u2705 <b>Заказ %s оформлен!</b>\n\nСумма: %s\nСпособ доставки: %s\n\nМы скоро свяжемся с вами для подтверждения.",
		order.OrderNumber,
		formatPrice(order.TotalPrice),
		deliveryMethodText(order.DeliveryMethod),
	)

	b.send(chatID, text)
	b.log.Info("sent order created notification", zap.String("order", order.OrderNumber), zap.Int64("chatID", chatID))
	return nil
}

// NotifyOrderStatusChanged sends a notification when the order status changes.
func (b *Bot) NotifyOrderStatusChanged(ctx context.Context, order *domain.Order) error {
	chatID, err := b.resolveUserChatID(ctx, order)
	if err != nil {
		return err
	}
	if chatID == 0 {
		return nil
	}

	var text string
	switch order.Status {
	case "confirmed":
		text = fmt.Sprintf("\U0001F44D <b>Заказ %s подтверждён!</b>\n\nМы начинаем подготовку вашего заказа.", order.OrderNumber)
	case "processing":
		text = fmt.Sprintf("\u2699\uFE0F <b>Заказ %s собирается</b>\n\nВаш заказ в работе. Совсем скоро отправим!", order.OrderNumber)
	case "shipped":
		text = fmt.Sprintf("\U0001F69A <b>Заказ %s отправлен!</b>", order.OrderNumber)
		if order.TrackingNumber != nil && *order.TrackingNumber != "" {
			text += fmt.Sprintf("\n\nТрек-номер: <code>%s</code>", *order.TrackingNumber)
		}
	case "delivered":
		text = fmt.Sprintf("\U0001F389 <b>Заказ %s доставлен!</b>\n\nСпасибо за покупку! Будем рады видеть вас снова.", order.OrderNumber)
	case "cancelled":
		text = fmt.Sprintf("\u274C <b>Заказ %s отменён</b>\n\nЕсли у вас есть вопросы, свяжитесь с нами через /help.", order.OrderNumber)
	default:
		return nil
	}

	b.send(chatID, text)
	b.log.Info("sent order status notification", zap.String("order", order.OrderNumber), zap.String("status", order.Status), zap.Int64("chatID", chatID))
	return nil
}

// resolveUserChatID finds the Telegram chat ID for the order's user.
func (b *Bot) resolveUserChatID(ctx context.Context, order *domain.Order) (int64, error) {
	if order.UserID == nil {
		return 0, nil
	}

	user, err := b.userRepo.FindByID(ctx, *order.UserID)
	if err != nil {
		b.log.Warn("failed to find user for notification", zap.Int("userID", *order.UserID), zap.Error(err))
		return 0, err
	}

	if user.TelegramID == nil {
		return 0, nil
	}

	return *user.TelegramID, nil
}

func deliveryMethodText(method string) string {
	switch method {
	case "pickup":
		return "Самовывоз"
	case "courier":
		return "Курьер"
	default:
		return method
	}
}
