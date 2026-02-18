// Package bitrix provides a client for the Bitrix24 REST API (webhook-based auth).
// Docs: https://dev.1c-bitrix.ru/rest_help/
package bitrix

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is a lightweight Bitrix24 REST API client using outgoing webhook authentication.
// Set up a Bitrix24 outgoing webhook:
//   Settings → Developer resources → Other → Inbound webhook
//   Grant access to: CRM
//   The resulting URL looks like: https://{portal}/rest/{userID}/{token}/
type Client struct {
	baseURL    string // e.g. "https://company.bitrix24.ru/rest/1/abc123/"
	httpClient *http.Client
}

// NewClient creates a Bitrix24 REST client.
// portal: e.g. "company.bitrix24.ru"
// userID: numeric user ID from the webhook URL
// token:  webhook access token
func NewClient(portal string, userID int, token string) *Client {
	return &Client{
		baseURL: fmt.Sprintf("https://%s/rest/%d/%s/", portal, userID, token),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Deal represents a Bitrix24 CRM deal.
type Deal struct {
	ID            string  `json:"ID"`
	Title         string  `json:"TITLE"`
	StageID       string  `json:"STAGE_ID"`
	Opportunity   string  `json:"OPPORTUNITY"`  // decimal as string
	CurrencyID    string  `json:"CURRENCY_ID"`
	Comments      string  `json:"COMMENTS"`
	// Custom field: our internal order number (create as UF_CRM_ORDER_NUMBER in Bitrix)
	OrderNumber   string  `json:"UF_CRM_ORDER_NUMBER"`
}

// CreateDealInput — fields for crm.deal.add
type CreateDealInput struct {
	Title       string
	StageID     string
	Opportunity float64
	Comments    string
	OrderNumber string // stored in UF_CRM_ORDER_NUMBER
}

// UpdateDealInput — fields for crm.deal.update (zero values are skipped)
type UpdateDealInput struct {
	StageID     string
	Opportunity float64
	Comments    string
}

// bitrixResponse is the envelope for all Bitrix REST responses.
type bitrixResponse struct {
	Result interface{} `json:"result"`
	Error  string      `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// CreateDeal calls crm.deal.add and returns the new deal ID.
func (c *Client) CreateDeal(ctx context.Context, input CreateDealInput) (string, error) {
	params := url.Values{}
	params.Set("fields[TITLE]", input.Title)
	params.Set("fields[STAGE_ID]", input.StageID)
	params.Set("fields[OPPORTUNITY]", fmt.Sprintf("%.2f", input.Opportunity))
	params.Set("fields[CURRENCY_ID]", "RUB")
	if input.Comments != "" {
		params.Set("fields[COMMENTS]", input.Comments)
	}
	if input.OrderNumber != "" {
		params.Set("fields[UF_CRM_ORDER_NUMBER]", input.OrderNumber)
	}

	var result interface{}
	if err := c.call(ctx, "crm.deal.add", params, &result); err != nil {
		return "", err
	}

	// Result is the new deal ID (float64 from JSON)
	switch v := result.(type) {
	case float64:
		return fmt.Sprintf("%.0f", v), nil
	case string:
		return v, nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// UpdateDeal calls crm.deal.update.
func (c *Client) UpdateDeal(ctx context.Context, dealID string, input UpdateDealInput) error {
	params := url.Values{}
	params.Set("id", dealID)
	if input.StageID != "" {
		params.Set("fields[STAGE_ID]", input.StageID)
	}
	if input.Opportunity > 0 {
		params.Set("fields[OPPORTUNITY]", fmt.Sprintf("%.2f", input.Opportunity))
	}
	if input.Comments != "" {
		params.Set("fields[COMMENTS]", input.Comments)
	}

	var result interface{}
	return c.call(ctx, "crm.deal.update", params, &result)
}

// GetDeal calls crm.deal.get and returns the deal.
func (c *Client) GetDeal(ctx context.Context, dealID string) (*Deal, error) {
	params := url.Values{}
	params.Set("id", dealID)

	var deal Deal
	if err := c.call(ctx, "crm.deal.get", params, &deal); err != nil {
		return nil, err
	}
	return &deal, nil
}

// call executes a Bitrix REST method via POST form and decodes the result into out.
func (c *Client) call(ctx context.Context, method string, params url.Values, out interface{}) error {
	reqURL := c.baseURL + method + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("bitrix: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("bitrix: http: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("bitrix: read body: %w", err)
	}

	var envelope struct {
		Result json.RawMessage `json:"result"`
		Error  string          `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("bitrix: decode response: %w", err)
	}
	if envelope.Error != "" {
		return fmt.Errorf("bitrix API error %s: %s", envelope.Error, envelope.ErrorDescription)
	}

	if out != nil && len(envelope.Result) > 0 {
		if err := json.Unmarshal(envelope.Result, out); err != nil {
			return fmt.Errorf("bitrix: decode result: %w", err)
		}
	}

	return nil
}
