package service

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/brown/3d-print-shop/internal/bitrix"
	"github.com/brown/3d-print-shop/internal/domain"
)

// orderStatusToBitrixStage maps our order status → Bitrix24 STAGE_ID.
// Adjust pipeline prefix ("C1") to match your Bitrix24 pipeline.
var orderStatusToBitrixStage = map[string]string{
	"new":       "NEW",
	"confirmed": "PREPARATION",
	"in_progress": "EXECUTING",
	"ready":     "FINAL_INVOICE",
	"delivered": "WON",
	"cancelled": "LOSE",
}

// bitrixStageToOrderStatus maps Bitrix24 STAGE_ID → our order status.
// Handles both bare IDs ("NEW") and pipeline-prefixed IDs ("C1:NEW").
var bitrixStageToOrderStatus = map[string]string{
	"new":          "new",
	"preparation":  "confirmed",
	"executing":    "in_progress",
	"final_invoice": "ready",
	"won":          "delivered",
	"lose":         "cancelled",
}

// BitrixService handles bidirectional sync between our orders and Bitrix24 CRM.
type BitrixService struct {
	client    *bitrix.Client
	orderRepo domain.OrderRepository
	customOrderRepo domain.CustomOrderRepository
	log       *zap.Logger
}

func NewBitrixService(
	client *bitrix.Client,
	orderRepo domain.OrderRepository,
	customOrderRepo domain.CustomOrderRepository,
	log *zap.Logger,
) *BitrixService {
	return &BitrixService{
		client:          client,
		orderRepo:       orderRepo,
		customOrderRepo: customOrderRepo,
		log:             log,
	}
}

// SyncOrderToBitrix creates or updates a Bitrix24 deal for a custom order.
// If the order already has a BitrixDealID, the deal is updated; otherwise a new one is created.
func (s *BitrixService) SyncOrderToBitrix(ctx context.Context, order *domain.Order) error {
	if order.CustomDetails == nil {
		return nil
	}

	stageID := orderStatusToBitrixStage[order.Status]
	if stageID == "" {
		stageID = "NEW"
	}

	title := fmt.Sprintf("Индивидуальный заказ %s — %s", order.OrderNumber, order.CustomerName)
	comments := buildComments(order)

	if order.CustomDetails.BitrixDealID != nil && *order.CustomDetails.BitrixDealID != "" {
		// Update existing deal
		err := s.client.UpdateDeal(ctx, *order.CustomDetails.BitrixDealID, bitrix.UpdateDealInput{
			StageID:     stageID,
			Opportunity: order.TotalPrice,
			Comments:    comments,
		})
		if err != nil {
			return fmt.Errorf("update bitrix deal: %w", err)
		}
		s.log.Info("bitrix deal updated",
			zap.String("dealID", *order.CustomDetails.BitrixDealID),
			zap.String("orderNumber", order.OrderNumber),
		)
	} else {
		// Create new deal
		dealID, err := s.client.CreateDeal(ctx, bitrix.CreateDealInput{
			Title:       title,
			StageID:     stageID,
			Opportunity: order.TotalPrice,
			Comments:    comments,
			OrderNumber: order.OrderNumber,
		})
		if err != nil {
			return fmt.Errorf("create bitrix deal: %w", err)
		}

		// Persist deal ID back to custom_order_details
		details := order.CustomDetails
		details.BitrixDealID = &dealID
		details.BitrixStageID = &stageID
		if err := s.customOrderRepo.Update(ctx, details); err != nil {
			s.log.Warn("failed to save bitrix deal ID to DB",
				zap.String("dealID", dealID),
				zap.Error(err),
			)
		}

		s.log.Info("bitrix deal created",
			zap.String("dealID", dealID),
			zap.String("orderNumber", order.OrderNumber),
		)
	}

	return nil
}

// HandleWebhook processes an incoming Bitrix24 webhook event.
// Bitrix sends form-encoded POST with keys like:
//
//	event=ONCRMDEALUPDATE&data[FIELDS][ID]=123
//
// We fetch the full deal from Bitrix and update the order status accordingly.
func (s *BitrixService) HandleWebhook(ctx context.Context, event string, dealID string) error {
	if dealID == "" {
		return fmt.Errorf("missing deal ID in webhook")
	}

	deal, err := s.client.GetDeal(ctx, dealID)
	if err != nil {
		return fmt.Errorf("get deal %s from bitrix: %w", dealID, err)
	}

	newStatus := stageToStatus(deal.StageID)
	if newStatus == "" {
		s.log.Debug("ignoring unknown bitrix stage",
			zap.String("stageID", deal.StageID),
			zap.String("dealID", dealID),
		)
		return nil
	}

	// Find order by bitrix_deal_id
	orders, _, err := s.orderRepo.List(ctx, domain.OrderFilter{OrderType: "custom", Limit: 1000})
	if err != nil {
		return fmt.Errorf("list custom orders: %w", err)
	}

	for _, o := range orders {
		if o.CustomDetails == nil {
			continue
		}
		if o.CustomDetails.BitrixDealID == nil || *o.CustomDetails.BitrixDealID != dealID {
			continue
		}

		if o.Status == newStatus {
			s.log.Debug("bitrix webhook: status unchanged, skipping",
				zap.String("orderNumber", o.OrderNumber),
				zap.String("status", newStatus),
			)
			return nil
		}

		if err := s.orderRepo.UpdateStatus(ctx, o.ID, newStatus); err != nil {
			return fmt.Errorf("update order status: %w", err)
		}

		// Persist updated stage ID
		details := o.CustomDetails
		details.BitrixStageID = &deal.StageID
		_ = s.customOrderRepo.Update(ctx, details)

		s.log.Info("bitrix webhook: order status updated",
			zap.String("orderNumber", o.OrderNumber),
			zap.String("oldStatus", o.Status),
			zap.String("newStatus", newStatus),
			zap.String("dealID", dealID),
		)
		return nil
	}

	s.log.Warn("bitrix webhook: no order found for deal",
		zap.String("dealID", dealID),
		zap.String("event", event),
	)
	return nil
}

// buildComments assembles a comments string from order details.
func buildComments(order *domain.Order) string {
	var parts []string
	if order.CustomerPhone != "" {
		parts = append(parts, "Телефон: "+order.CustomerPhone)
	}
	if order.CustomerEmail != nil {
		parts = append(parts, "Email: "+*order.CustomerEmail)
	}
	if order.CustomDetails != nil && order.CustomDetails.ClientDescription != nil {
		parts = append(parts, "Описание: "+*order.CustomDetails.ClientDescription)
	}
	if order.CustomDetails != nil && order.CustomDetails.AdminNotes != nil {
		parts = append(parts, "Заметки администратора: "+*order.CustomDetails.AdminNotes)
	}
	return strings.Join(parts, "\n")
}

// stageToStatus converts a Bitrix24 STAGE_ID to our order status.
// Strips pipeline prefix if present ("C1:NEW" → "new").
func stageToStatus(stageID string) string {
	bare := stageID
	if idx := strings.Index(stageID, ":"); idx >= 0 {
		bare = stageID[idx+1:]
	}
	return bitrixStageToOrderStatus[strings.ToLower(bare)]
}
