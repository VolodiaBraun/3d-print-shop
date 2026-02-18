package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/brown/3d-print-shop/internal/domain"
	"github.com/brown/3d-print-shop/internal/storage"
)

const (
	maxModelFileSize  = 50 << 20 // 50 MB
	maxModelFilesPerOrder = 5
)

var allowedModelExtensions = map[string]string{
	// 3D model formats
	".stl":  "model/stl",
	".obj":  "model/obj",
	".3mf":  "model/3mf",
	".step": "application/step",
	".stp":  "application/step",
	".zip":  "application/zip",
	// Image / reference formats
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
	".heic": "image/heic",
	".heif": "image/heif",
}

var ErrTooManyFiles      = errors.New("превышено максимальное количество файлов (5)")
var ErrFileTooLarge      = errors.New("файл слишком большой (максимум 50 MB)")
var ErrUnsupportedFormat = errors.New("неподдерживаемый формат файла (STL, OBJ, 3MF, STEP, ZIP, JPG, PNG, WEBP)")

// SubmitCustomOrderInput — клиент или фронт отправляет заявку на индивидуальный заказ.
type SubmitCustomOrderInput struct {
	CustomerName        string          `json:"customerName" binding:"required"`
	CustomerPhone       string          `json:"customerPhone" binding:"required"`
	CustomerEmail       *string         `json:"customerEmail"`
	TelegramID          *int64          `json:"telegramId"`
	ClientDescription   *string         `json:"clientDescription"`
	FileURLs            json.RawMessage `json:"fileUrls"`       // опционально
	PrintSettings       json.RawMessage `json:"printSettings"`  // опционально
	PaymentMethod       string          `json:"paymentMethod" binding:"required,oneof=card cash"`
	DeliveryMethod      string          `json:"deliveryMethod" binding:"required,oneof=pickup courier pickup_point"`
	DeliveryAddress     *string         `json:"deliveryAddress"`
	PickupPointID       *int            `json:"pickupPointId"`
	Notes               *string         `json:"notes"`
}

// CreateCustomOrderByAdminInput — администратор создаёт заказ вручную (заводит из Bitrix или самостоятельно).
type CreateCustomOrderByAdminInput struct {
	CustomerName        string          `json:"customerName" binding:"required"`
	CustomerPhone       string          `json:"customerPhone" binding:"required"`
	CustomerEmail       *string         `json:"customerEmail"`
	UserID              *int            `json:"userId"`
	ClientDescription   *string         `json:"clientDescription"`
	AdminNotes          *string         `json:"adminNotes"`
	FileURLs            json.RawMessage `json:"fileUrls"`
	PrintSettings       json.RawMessage `json:"printSettings"`
	PaymentMethod       string          `json:"paymentMethod" binding:"required,oneof=card cash"`
	DeliveryMethod      string          `json:"deliveryMethod" binding:"required,oneof=pickup courier pickup_point"`
	DeliveryAddress     *string         `json:"deliveryAddress"`
	PickupPointID       *int            `json:"pickupPointId"`
	// Items: позиции (услуги), которые администратор добавляет вручную.
	Items               []CustomOrderItemInput `json:"items"`
	// Если поле SubtotalOverride задано > 0, subtotal берётся из него, а не считается из items.
	SubtotalOverride    float64         `json:"subtotalOverride"`
	Notes               *string         `json:"notes"`
	BitrixDealID        *string         `json:"bitrixDealId"`
	BitrixStageID       *string         `json:"bitrixStageId"`
}

type CustomOrderItemInput struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	Quantity    int     `json:"quantity" binding:"required,gt=0"`
	UnitPrice   float64 `json:"unitPrice" binding:"required,gt=0"`
}

// UpdateCustomOrderAdminInput — обновление заметок / цены / статуса admin-панелью.
type UpdateCustomOrderAdminInput struct {
	AdminNotes    *string         `json:"adminNotes"`
	PrintSettings json.RawMessage `json:"printSettings"`
	FileURLs      json.RawMessage `json:"fileUrls"`
	TotalPrice    *float64        `json:"totalPrice"` // если нужно скорректировать итог вручную
	BitrixDealID  *string         `json:"bitrixDealId"`
	BitrixStageID *string         `json:"bitrixStageId"`
}

type CustomOrderService struct {
	orderRepo       domain.OrderRepository
	customOrderRepo domain.CustomOrderRepository
	userRepo        domain.UserRepository
	paymentService  *PaymentService
	bitrixService   *BitrixService
	notifier        domain.OrderNotifier
	emailService    *EmailService
	s3              *storage.S3Client
	db              *gorm.DB
	log             *zap.Logger
}

func NewCustomOrderService(
	orderRepo domain.OrderRepository,
	customOrderRepo domain.CustomOrderRepository,
	userRepo domain.UserRepository,
	db *gorm.DB,
	log *zap.Logger,
) *CustomOrderService {
	return &CustomOrderService{
		orderRepo:       orderRepo,
		customOrderRepo: customOrderRepo,
		userRepo:        userRepo,
		db:              db,
		log:             log,
	}
}

func (s *CustomOrderService) SetPaymentService(ps *PaymentService) {
	s.paymentService = ps
}

func (s *CustomOrderService) SetNotifier(n domain.OrderNotifier) {
	s.notifier = n
}

func (s *CustomOrderService) SetEmailService(es *EmailService) {
	s.emailService = es
}

func (s *CustomOrderService) SetS3Client(s3 *storage.S3Client) {
	s.s3 = s3
}

func (s *CustomOrderService) SetBitrixService(bs *BitrixService) {
	s.bitrixService = bs
}

// SubmitRequest — клиент оставляет заявку. Заказ создаётся со статусом "new", цена = 0 (неизвестна).
func (s *CustomOrderService) SubmitRequest(ctx context.Context, input SubmitCustomOrderInput) (*domain.Order, error) {
	// Resolve or create user
	var userID *int
	if input.TelegramID != nil && *input.TelegramID != 0 {
		user, err := s.userRepo.FindByTelegramID(ctx, *input.TelegramID)
		if err == nil {
			userID = &user.ID
		}
	}

	orderNumber, err := s.orderRepo.NextOrderNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("generate order number: %w", err)
	}

	fileURLs := defaultJSONArray(input.FileURLs)
	printSettings := defaultJSONObject(input.PrintSettings)

	order := &domain.Order{
		OrderNumber:     orderNumber,
		UserID:          userID,
		OrderType:       "custom",
		Status:          "new",
		Subtotal:        0,
		TotalPrice:      0,
		DeliveryMethod:  input.DeliveryMethod,
		DeliveryAddress: input.DeliveryAddress,
		PaymentMethod:   input.PaymentMethod,
		CustomerName:    input.CustomerName,
		CustomerPhone:   input.CustomerPhone,
		CustomerEmail:   input.CustomerEmail,
		PickupPointID:   input.PickupPointID,
		Notes:           input.Notes,
	}

	details := &domain.CustomOrderDetails{
		ClientDescription: input.ClientDescription,
		FileURLs:          fileURLs,
		PrintSettings:     printSettings,
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return fmt.Errorf("create order: %w", err)
		}
		details.OrderID = order.ID
		if err := tx.Create(details).Error; err != nil {
			return fmt.Errorf("create custom details: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	created, err := s.orderRepo.FindByID(ctx, order.ID)
	if err != nil {
		created = order
	}

	s.log.Info("custom order submitted by client",
		zap.String("orderNumber", order.OrderNumber),
	)

	go s.sendNewOrderNotifications(created)

	return created, nil
}

// CreateByAdmin — администратор создаёт заказ с уже известными позициями и ценами.
func (s *CustomOrderService) CreateByAdmin(ctx context.Context, input CreateCustomOrderByAdminInput) (*domain.Order, error) {
	var subtotal float64
	var orderItems []domain.OrderItem

	if input.SubtotalOverride > 0 {
		subtotal = math.Round(input.SubtotalOverride*100) / 100
	} else {
		for _, it := range input.Items {
			itemTotal := math.Round(it.UnitPrice*float64(it.Quantity)*100) / 100
			subtotal += itemTotal
			desc := it.Description
			orderItems = append(orderItems, domain.OrderItem{
				CustomItemName:        &it.Name,
				CustomItemDescription: desc,
				Quantity:              it.Quantity,
				UnitPrice:             it.UnitPrice,
				TotalPrice:            itemTotal,
			})
		}
		subtotal = math.Round(subtotal*100) / 100
	}

	orderNumber, err := s.orderRepo.NextOrderNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("generate order number: %w", err)
	}

	fileURLs := defaultJSONArray(input.FileURLs)
	printSettings := defaultJSONObject(input.PrintSettings)

	order := &domain.Order{
		OrderNumber:     orderNumber,
		UserID:          input.UserID,
		OrderType:       "custom",
		Status:          "confirmed",
		Subtotal:        subtotal,
		TotalPrice:      subtotal,
		DeliveryMethod:  input.DeliveryMethod,
		DeliveryAddress: input.DeliveryAddress,
		PaymentMethod:   input.PaymentMethod,
		CustomerName:    input.CustomerName,
		CustomerPhone:   input.CustomerPhone,
		CustomerEmail:   input.CustomerEmail,
		PickupPointID:   input.PickupPointID,
		Notes:           input.Notes,
		Items:           orderItems,
	}

	details := &domain.CustomOrderDetails{
		ClientDescription: input.ClientDescription,
		AdminNotes:        input.AdminNotes,
		FileURLs:          fileURLs,
		PrintSettings:     printSettings,
		BitrixDealID:      input.BitrixDealID,
		BitrixStageID:     input.BitrixStageID,
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return fmt.Errorf("create order: %w", err)
		}
		details.OrderID = order.ID
		if err := tx.Create(details).Error; err != nil {
			return fmt.Errorf("create custom details: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	created, err := s.orderRepo.FindByID(ctx, order.ID)
	if err != nil {
		created = order
	}

	// Если метод оплаты — карта, генерируем ссылку сразу
	if input.PaymentMethod == "card" && s.paymentService != nil {
		if _, payErr := s.paymentService.InitiatePayment(ctx, created); payErr != nil {
			s.log.Warn("failed to initiate payment for custom order", zap.Error(payErr))
		} else if reloaded, reloadErr := s.orderRepo.FindByID(ctx, order.ID); reloadErr == nil {
			created = reloaded
		}
	}

	s.log.Info("custom order created by admin",
		zap.String("orderNumber", order.OrderNumber),
		zap.Float64("total", order.TotalPrice),
	)

	go s.sendNewOrderNotifications(created)

	return created, nil
}

// ConfirmRequest — администратор подтверждает заявку клиента и выставляет цену.
func (s *CustomOrderService) ConfirmRequest(ctx context.Context, orderID int, totalPrice float64, adminNotes *string) (*domain.Order, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.OrderType != "custom" {
		return nil, domain.ErrOrderNotCustom
	}
	if order.Status != "new" {
		return nil, domain.ErrCustomOrderAlreadyConfirmed
	}

	totalPrice = math.Round(totalPrice*100) / 100

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domain.Order{}).
			Where("id = ?", orderID).
			Updates(map[string]interface{}{
				"status":      "confirmed",
				"subtotal":    totalPrice,
				"total_price": totalPrice,
			}).Error; err != nil {
			return err
		}
		if adminNotes != nil {
			if err := tx.Model(&domain.CustomOrderDetails{}).
				Where("order_id = ?", orderID).
				Update("admin_notes", *adminNotes).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	updated, _ := s.orderRepo.FindByID(ctx, orderID)

	// Если оплата картой, создаём ссылку
	if order.PaymentMethod == "card" && s.paymentService != nil {
		if _, payErr := s.paymentService.InitiatePayment(ctx, updated); payErr != nil {
			s.log.Warn("failed to initiate payment after confirm", zap.Error(payErr))
		} else if reloaded, err := s.orderRepo.FindByID(ctx, orderID); err == nil {
			updated = reloaded
		}
	}

	go func() {
		bgCtx := context.Background()
		if s.notifier != nil {
			_ = s.notifier.NotifyOrderStatusChanged(bgCtx, updated)
		}
		if s.emailService != nil {
			s.emailService.SendOrderStatusChanged(updated)
		}
		if s.bitrixService != nil {
			if err := s.bitrixService.SyncOrderToBitrix(bgCtx, updated); err != nil {
				s.log.Warn("failed to sync confirmed order to bitrix", zap.Error(err))
			}
		}
	}()

	return updated, nil
}

// SendPaymentLink — пересоздаёт и отправляет ссылку на оплату (для cash → card или если ссылка устарела).
func (s *CustomOrderService) SendPaymentLink(ctx context.Context, orderID int) (*domain.Order, error) {
	if s.paymentService == nil {
		return nil, fmt.Errorf("payment service not configured")
	}
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.OrderType != "custom" {
		return nil, domain.ErrOrderNotCustom
	}
	if _, err := s.paymentService.InitiatePayment(ctx, order); err != nil {
		return nil, fmt.Errorf("initiate payment: %w", err)
	}
	return s.orderRepo.FindByID(ctx, orderID)
}

// MarkPaidManually — администратор вручную помечает заказ оплаченным (наличные / перевод).
func (s *CustomOrderService) MarkPaidManually(ctx context.Context, orderID int) error {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.OrderType != "custom" {
		return domain.ErrOrderNotCustom
	}
	return s.db.WithContext(ctx).
		Model(&domain.Order{}).
		Where("id = ?", orderID).
		Update("is_paid", true).Error
}

// UpdateAdminDetails — обновляет admin_notes, print_settings, file_urls, bitrix поля, total_price.
func (s *CustomOrderService) UpdateAdminDetails(ctx context.Context, orderID int, input UpdateCustomOrderAdminInput) (*domain.Order, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.OrderType != "custom" {
		return nil, domain.ErrOrderNotCustom
	}

	details, err := s.customOrderRepo.FindByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if input.AdminNotes != nil {
		details.AdminNotes = input.AdminNotes
	}
	if len(input.PrintSettings) > 0 {
		details.PrintSettings = input.PrintSettings
	}
	if len(input.FileURLs) > 0 {
		details.FileURLs = input.FileURLs
	}
	if input.BitrixDealID != nil {
		details.BitrixDealID = input.BitrixDealID
	}
	if input.BitrixStageID != nil {
		details.BitrixStageID = input.BitrixStageID
	}

	if err := s.customOrderRepo.Update(ctx, details); err != nil {
		return nil, fmt.Errorf("update custom details: %w", err)
	}

	if input.TotalPrice != nil && *input.TotalPrice >= 0 {
		rounded := math.Round(*input.TotalPrice*100) / 100
		if err := s.db.WithContext(ctx).
			Model(&domain.Order{}).
			Where("id = ?", orderID).
			Updates(map[string]interface{}{
				"subtotal":    rounded,
				"total_price": rounded,
			}).Error; err != nil {
			return nil, fmt.Errorf("update order price: %w", err)
		}
	}

	return s.orderRepo.FindByID(ctx, orderID)
}

// GetByID возвращает любой заказ по ID с полным preload (включая CustomDetails).
func (s *CustomOrderService) GetByID(ctx context.Context, orderID int) (*domain.Order, error) {
	return s.orderRepo.FindByID(ctx, orderID)
}

// ListCustomOrders — список только custom-заказов.
func (s *CustomOrderService) ListCustomOrders(ctx context.Context, filter domain.OrderFilter) ([]domain.Order, int64, error) {
	filter.OrderType = "custom"
	return s.orderRepo.List(ctx, filter)
}

// sendNewOrderNotifications — асинхронные уведомления при создании.
func (s *CustomOrderService) sendNewOrderNotifications(order *domain.Order) {
	bgCtx := context.Background()
	if s.notifier != nil {
		_ = s.notifier.NotifyOrderCreated(bgCtx, order)
		_ = s.notifier.NotifyAdminNewOrder(bgCtx, order)
	}
	if s.emailService != nil {
		s.emailService.SendOrderCreated(order)
	}
	if s.bitrixService != nil {
		if err := s.bitrixService.SyncOrderToBitrix(bgCtx, order); err != nil {
			s.log.Warn("failed to sync new order to bitrix", zap.Error(err), zap.String("orderNumber", order.OrderNumber))
		}
	}
}

// UploadModelFile загружает 3D-файл в S3 и добавляет URL в file_urls заказа.
// Возвращает публичный URL загруженного файла.
func (s *CustomOrderService) UploadModelFile(ctx context.Context, orderID int, fileName string, file io.Reader, fileSize int64) (string, error) {
	if s.s3 == nil {
		return "", fmt.Errorf("file storage not configured")
	}

	// Validate extension
	ext := strings.ToLower(filepath.Ext(fileName))
	contentType, ok := allowedModelExtensions[ext]
	if !ok {
		return "", ErrUnsupportedFormat
	}

	// Validate size
	if fileSize > maxModelFileSize {
		return "", ErrFileTooLarge
	}

	// Load current file list and check count
	details, err := s.customOrderRepo.FindByOrderID(ctx, orderID)
	if err != nil {
		return "", err
	}

	var urls []string
	if err := json.Unmarshal(details.FileURLs, &urls); err != nil {
		urls = []string{}
	}
	if len(urls) >= maxModelFilesPerOrder {
		return "", ErrTooManyFiles
	}

	// Upload to S3
	key := fmt.Sprintf("custom-orders/%d/%s%s", orderID, uuid.New().String(), ext)
	publicURL, err := s.s3.UploadFromReader(ctx, key, file, contentType)
	if err != nil {
		return "", fmt.Errorf("upload to s3: %w", err)
	}

	// Append URL and save
	urls = append(urls, publicURL)
	newFileURLs, _ := json.Marshal(urls)
	details.FileURLs = json.RawMessage(newFileURLs)
	if err := s.customOrderRepo.Update(ctx, details); err != nil {
		// Best-effort cleanup of S3 object
		_ = s.s3.Delete(ctx, key)
		return "", fmt.Errorf("save file url: %w", err)
	}

	s.log.Info("model file uploaded",
		zap.Int("orderID", orderID),
		zap.String("key", key),
	)

	return publicURL, nil
}

// DeleteModelFile удаляет 3D-файл из S3 и убирает URL из file_urls.
func (s *CustomOrderService) DeleteModelFile(ctx context.Context, orderID int, fileURL string) error {
	if s.s3 == nil {
		return fmt.Errorf("file storage not configured")
	}

	details, err := s.customOrderRepo.FindByOrderID(ctx, orderID)
	if err != nil {
		return err
	}

	var urls []string
	if err := json.Unmarshal(details.FileURLs, &urls); err != nil {
		return fmt.Errorf("parse file urls: %w", err)
	}

	// Remove the target URL from the list
	newURLs := make([]string, 0, len(urls))
	found := false
	for _, u := range urls {
		if u == fileURL {
			found = true
		} else {
			newURLs = append(newURLs, u)
		}
	}
	if !found {
		return fmt.Errorf("file not found in order")
	}

	// Extract S3 key from URL and delete from storage
	// S3 key starts after publicURL prefix; find "custom-orders/" in the URL.
	keyIdx := strings.Index(fileURL, "custom-orders/")
	if keyIdx >= 0 {
		key := fileURL[keyIdx:]
		if delErr := s.s3.Delete(ctx, key); delErr != nil {
			s.log.Warn("failed to delete model file from s3", zap.String("key", key), zap.Error(delErr))
		}
	}

	newFileURLs, _ := json.Marshal(newURLs)
	details.FileURLs = json.RawMessage(newFileURLs)
	return s.customOrderRepo.Update(ctx, details)
}

// defaultJSONArray возвращает пустой JSON-массив если input пустой.
func defaultJSONArray(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage("[]")
	}
	return raw
}

// defaultJSONObject возвращает пустой JSON-объект если input пустой.
func defaultJSONObject(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage("{}")
	}
	return raw
}
