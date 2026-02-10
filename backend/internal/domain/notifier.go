package domain

import "context"

// OrderNotifier sends notifications about order events.
type OrderNotifier interface {
	NotifyOrderCreated(ctx context.Context, order *Order) error
	NotifyOrderStatusChanged(ctx context.Context, order *Order) error
}
