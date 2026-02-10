package telegram

import (
	"fmt"
	"strings"
	"time"

	"github.com/brown/3d-print-shop/internal/domain"
)

const welcomeText = `<b>Добро пожаловать в АВАНГАРД!</b>

Мы создаём уникальные 3D-печатные изделия премиального качества.

Выберите действие:`

const helpText = `<b>Команды бота:</b>

/start — Главное меню
/orders — Мои заказы
/help — Помощь и контакты

<b>Контакты:</b>
Telegram: @avangard3d
Email: info@avangard3d.ru
Сайт: avangard-print.ru`

const ordersNotLinkedText = `У вас пока нет привязанных заказов.

Скоро вы сможете просматривать заказы прямо в боте!

А пока вы можете отслеживать заказ по номеру на сайте.`

const noOrdersText = `У вас пока нет заказов.

Откройте каталог, чтобы сделать первый заказ!`

func statusEmoji(status string) string {
	switch status {
	case "new":
		return "\U0001F195"
	case "confirmed":
		return "\u2705"
	case "processing":
		return "\u2699\uFE0F"
	case "shipped":
		return "\U0001F4E6"
	case "delivered":
		return "\U0001F389"
	case "cancelled":
		return "\u274C"
	default:
		return "\u2753"
	}
}

func statusText(status string) string {
	switch status {
	case "new":
		return "Новый"
	case "confirmed":
		return "Подтверждён"
	case "processing":
		return "В обработке"
	case "shipped":
		return "Отправлен"
	case "delivered":
		return "Доставлен"
	case "cancelled":
		return "Отменён"
	default:
		return status
	}
}

func formatPrice(price float64) string {
	return fmt.Sprintf("%.0f ₽", price)
}

func formatDate(t time.Time) string {
	return t.Format("02.01.2006")
}

func formatOrdersList(orders []domain.Order) string {
	var sb strings.Builder
	sb.WriteString("<b>\U0001F4CB Ваши заказы:</b>\n")

	limit := 5
	if len(orders) < limit {
		limit = len(orders)
	}

	for i := 0; i < limit; i++ {
		o := orders[i]
		sb.WriteString(fmt.Sprintf(
			"\n%s <b>%s</b> — %s\n   Сумма: %s\n   Создан: %s\n",
			statusEmoji(o.Status),
			o.OrderNumber,
			statusText(o.Status),
			formatPrice(o.TotalPrice),
			formatDate(o.CreatedAt),
		))
		if o.TrackingNumber != nil && *o.TrackingNumber != "" {
			sb.WriteString(fmt.Sprintf("   Трек: %s\n", *o.TrackingNumber))
		}
	}

	if len(orders) > 5 {
		sb.WriteString(fmt.Sprintf("\nПоказаны последние 5 из %d заказов.", len(orders)))
	}

	return sb.String()
}
