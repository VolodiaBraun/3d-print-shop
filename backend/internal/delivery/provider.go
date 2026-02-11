package delivery

import "context"

// DeliveryOption represents one delivery option returned to the user.
type DeliveryOption struct {
	Type             string  `json:"type"`
	Name             string  `json:"name"`
	Cost             float64 `json:"cost"`
	OriginalCost     float64 `json:"originalCost"`
	EstimatedDaysMin int     `json:"estimatedDaysMin"`
	EstimatedDaysMax int     `json:"estimatedDaysMax"`
	IsFreeDelivery   bool    `json:"isFreeDelivery"`
	ProviderName     string  `json:"providerName"`
}

// CalculateInput contains data needed to calculate delivery cost.
type CalculateInput struct {
	City        string
	TotalWeight float64 // grams
	OrderTotal  float64
}

// ShipmentInput contains data for creating a shipment.
type ShipmentInput struct {
	OrderID        int
	OrderNumber    string
	RecipientName  string
	RecipientPhone string
	Address        string
	City           string
	PickupPointID  *int
	Items          []ShipmentItem
}

// ShipmentItem represents one item in a shipment.
type ShipmentItem struct {
	Name     string
	Quantity int
	Weight   float64 // grams
	Price    float64
}

// ShipmentResult is returned after creating a shipment.
type ShipmentResult struct {
	TrackingNumber string `json:"trackingNumber"`
	ProviderName   string `json:"providerName"`
}

// TrackingEvent represents one tracking status update.
type TrackingEvent struct {
	Status    string `json:"status"`
	Location  string `json:"location"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

// Provider is the abstract interface for delivery providers.
// Implement this for CDEK, Boxberry, or any other carrier.
type Provider interface {
	Name() string
	CalculateDelivery(ctx context.Context, input CalculateInput) ([]DeliveryOption, error)
	CreateShipment(ctx context.Context, input ShipmentInput) (*ShipmentResult, error)
	TrackShipment(ctx context.Context, trackingNumber string) ([]TrackingEvent, error)
}
