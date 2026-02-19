package service

import (
	"bytes"
	"crypto/tls"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/smtp"

	"go.uber.org/zap"

	"github.com/brown/3d-print-shop/internal/config"
	"github.com/brown/3d-print-shop/internal/domain"
)

//go:embed email_templates/*.html
var emailTemplates embed.FS

type EmailService struct {
	cfg       config.SMTPConfig
	log       *zap.Logger
	templates *template.Template
}

func NewEmailService(cfg config.SMTPConfig, log *zap.Logger) (*EmailService, error) {
	tmpl, err := template.ParseFS(emailTemplates, "email_templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse email templates: %w", err)
	}

	return &EmailService{
		cfg:       cfg,
		log:       log,
		templates: tmpl,
	}, nil
}

type orderEmailData struct {
	OrderNumber    string
	CustomerName   string
	Status         string
	StatusRu       string
	Items          []orderItemData
	Subtotal       string
	Discount       string
	BonusDiscount  string
	DeliveryCost   string
	TotalPrice     string
	DeliveryMethod string
	PaymentMethod  string
	TrackingNumber string
	Year           int
}

type orderItemData struct {
	Name      string
	Quantity  int
	UnitPrice string
	Total     string
}

func (s *EmailService) SendOrderCreated(order *domain.Order) {
	if order.CustomerEmail == nil || *order.CustomerEmail == "" {
		return
	}

	data := s.buildOrderData(order)
	var buf bytes.Buffer
	if err := s.templates.ExecuteTemplate(&buf, "order_created.html", data); err != nil {
		s.log.Warn("failed to render order_created email", zap.Error(err))
		return
	}

	subject := fmt.Sprintf("Заказ #%s оформлен — АВАНГАРД", order.OrderNumber)
	if err := s.send(*order.CustomerEmail, subject, buf.String()); err != nil {
		s.log.Warn("failed to send order_created email",
			zap.Error(err),
			zap.String("to", *order.CustomerEmail),
			zap.String("order", order.OrderNumber),
		)
		return
	}

	s.log.Info("order_created email sent",
		zap.String("to", *order.CustomerEmail),
		zap.String("order", order.OrderNumber),
	)
}

func (s *EmailService) SendOrderStatusChanged(order *domain.Order) {
	if order.CustomerEmail == nil || *order.CustomerEmail == "" {
		return
	}

	data := s.buildOrderData(order)
	var buf bytes.Buffer
	if err := s.templates.ExecuteTemplate(&buf, "order_status.html", data); err != nil {
		s.log.Warn("failed to render order_status email", zap.Error(err))
		return
	}

	subject := fmt.Sprintf("Заказ #%s — %s — АВАНГАРД", order.OrderNumber, data.StatusRu)
	if err := s.send(*order.CustomerEmail, subject, buf.String()); err != nil {
		s.log.Warn("failed to send order_status email",
			zap.Error(err),
			zap.String("to", *order.CustomerEmail),
			zap.String("order", order.OrderNumber),
		)
		return
	}

	s.log.Info("order_status email sent",
		zap.String("to", *order.CustomerEmail),
		zap.String("order", order.OrderNumber),
		zap.String("status", order.Status),
	)
}

func (s *EmailService) SendVerificationCode(to, code string) {
	data := struct {
		Code string
		Year int
	}{Code: code, Year: 2026}

	var buf bytes.Buffer
	if err := s.templates.ExecuteTemplate(&buf, "verify_code.html", data); err != nil {
		s.log.Warn("failed to render verify_code email", zap.Error(err))
		return
	}

	subject := "Код подтверждения — АВАНГАРД"
	if err := s.send(to, subject, buf.String()); err != nil {
		s.log.Warn("failed to send verification email",
			zap.Error(err),
			zap.String("to", to),
		)
		return
	}

	s.log.Info("verification email sent", zap.String("to", to))
}

func (s *EmailService) buildOrderData(order *domain.Order) orderEmailData {
	data := orderEmailData{
		OrderNumber:    order.OrderNumber,
		CustomerName:   order.CustomerName,
		Status:         order.Status,
		StatusRu:       statusToRussian(order.Status),
		Subtotal:       formatPrice(order.Subtotal),
		Discount:       formatPrice(order.DiscountAmount),
		BonusDiscount:  formatPrice(order.BonusDiscount),
		DeliveryCost:   formatPrice(order.DeliveryCost),
		TotalPrice:     formatPrice(order.TotalPrice),
		DeliveryMethod: deliveryMethodRu(order.DeliveryMethod),
		PaymentMethod:  paymentMethodRu(order.PaymentMethod),
		Year:           2026,
	}

	if order.TrackingNumber != nil {
		data.TrackingNumber = *order.TrackingNumber
	}

	for _, item := range order.Items {
		data.Items = append(data.Items, orderItemData{
			Name:      item.Product.Name,
			Quantity:  item.Quantity,
			UnitPrice: formatPrice(item.UnitPrice),
			Total:     formatPrice(item.TotalPrice),
		})
	}

	return data
}

func (s *EmailService) send(to, subject, htmlBody string) error {
	from := s.cfg.FromEmail
	fromHeader := fmt.Sprintf("%s <%s>", s.cfg.FromName, from)

	msg := "MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		fmt.Sprintf("From: %s\r\n", fromHeader) +
		fmt.Sprintf("To: %s\r\n", to) +
		fmt.Sprintf("Subject: =?UTF-8?B?%s?=\r\n", base64Encode(subject)) +
		"\r\n" +
		htmlBody

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)

	tlsConfig := &tls.Config{
		ServerName: s.cfg.Host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	if _, err := w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}

	return client.Quit()
}

func statusToRussian(status string) string {
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

func deliveryMethodRu(method string) string {
	switch method {
	case "pickup":
		return "Самовывоз"
	case "courier":
		return "Курьером"
	case "post":
		return "Почтой"
	default:
		return method
	}
}

func paymentMethodRu(method string) string {
	switch method {
	case "card":
		return "Картой онлайн"
	case "cash":
		return "Наличными"
	case "transfer":
		return "Перевод"
	default:
		return method
	}
}

func formatPrice(price float64) string {
	if price == float64(int(price)) {
		return fmt.Sprintf("%.0f ₽", price)
	}
	return fmt.Sprintf("%.2f ₽", price)
}

func base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
