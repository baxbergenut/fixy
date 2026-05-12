package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"

	"fixy-backend/internal/models"
)

type UpstreamStatusError struct {
	StatusCode int
	Message    string
}

func (err *UpstreamStatusError) Error() string {
	return err.Message
}

type groqChatCompletionRequest struct {
	Model               string          `json:"model"`
	Messages            []groqMessage   `json:"messages"`
	ResponseFormat      groqResponseFmt `json:"response_format"`
	Temperature         float64         `json:"temperature"`
	MaxCompletionTokens int             `json:"max_completion_tokens,omitempty"`
	TopP                float64         `json:"top_p,omitempty"`
}

type groqMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type groqContentPart struct {
	Type     string        `json:"type"`
	Text     string        `json:"text,omitempty"`
	ImageURL *groqImageURL `json:"image_url,omitempty"`
}

type groqImageURL struct {
	URL string `json:"url"`
}

type groqResponseFmt struct {
	Type string `json:"type"`
}

type groqChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

const (
	groqInvoiceModel = "meta-llama/llama-4-scout-17b-16e-instruct"

	invoiceParsePrompt = `You are parsing a trucking fleet maintenance invoice.
Extract the following fields and return ONLY valid JSON, no prose:
{
  "vendor": "string or null",
  "expense_date": "YYYY-MM-DD or null",
  "truck_unit_number": "unit number like 071 or null",
  "driver_name": "string or null",
  "amount": number or null,
  "category": one of [PM Service, Oil change, Tire issue, Engine issue, Towing, Road Service, Body work, Leakage, Kris Shop, Truck Wash/Detailing, Electrical issue, Fluids/Truck Parts, Brakes/Drums/Rotors, Scale, Other],
  "description": "brief description of work done or null",
  "reference_number": "invoice or transaction number or null"
}
If a field cannot be determined, use null.`
	invoiceParseImagePrompt = "Extract the invoice fields from this image and return only JSON."
	invoiceParseTextPrompt  = "Extract the invoice fields from this text and return only JSON."
)

func ParseInvoice(ctx context.Context, token, mimeType string, data []byte) (models.InvoiceParseResult, error) {
	if strings.TrimSpace(token) == "" {
		return models.InvoiceParseResult{}, errors.New("GROQ_TOKEN is not configured")
	}

	if strings.TrimSpace(mimeType) == "" {
		mimeType = "application/octet-stream"
	}

	messages, err := buildGroqInvoiceMessages(mimeType, data)
	if err != nil {
		return models.InvoiceParseResult{}, err
	}

	requestBody := groqChatCompletionRequest{
		Model:    groqInvoiceModel,
		Messages: messages,
		ResponseFormat: groqResponseFmt{
			Type: "json_object",
		},
		Temperature:         0,
		MaxCompletionTokens: 1024,
		TopP:                1,
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return models.InvoiceParseResult{}, fmt.Errorf("marshal groq request: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.groq.com/openai/v1/chat/completions",
		bytes.NewReader(payload),
	)
	if err != nil {
		return models.InvoiceParseResult{}, fmt.Errorf("build groq request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 90 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return models.InvoiceParseResult{}, fmt.Errorf("call groq: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return models.InvoiceParseResult{}, fmt.Errorf("read groq response: %w", err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return models.InvoiceParseResult{}, &UpstreamStatusError{
			StatusCode: response.StatusCode,
			Message:    fmt.Sprintf("groq returned %s: %s", response.Status, strings.TrimSpace(string(body))),
		}
	}

	var parsed groqChatCompletionResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return models.InvoiceParseResult{}, fmt.Errorf("decode groq response: %w", err)
	}

	textResponse := groqResponseText(parsed)
	if textResponse == "" {
		return models.InvoiceParseResult{}, errors.New("groq response did not include parsed text")
	}

	return parseInvoiceResult(textResponse)
}

func buildGroqInvoiceMessages(mimeType string, data []byte) ([]groqMessage, error) {
	switch {
	case strings.EqualFold(mimeType, "application/pdf"):
		text, err := extractPlainTextFromPDF(data)
		if err != nil {
			return nil, err
		}

		text = strings.TrimSpace(text)
		if text == "" {
			return nil, errors.New("the uploaded PDF did not contain extractable text; upload an image instead")
		}

		return []groqMessage{
			{Role: "system", Content: invoiceParsePrompt},
			{Role: "user", Content: invoiceParseTextPrompt + "\n\nInvoice text:\n" + text},
		}, nil
	case strings.HasPrefix(strings.ToLower(mimeType), "image/"):
		encoded := base64.StdEncoding.EncodeToString(data)
		return []groqMessage{
			{Role: "system", Content: invoiceParsePrompt},
			{
				Role: "user",
				Content: []groqContentPart{
					{Type: "text", Text: invoiceParseImagePrompt},
					{
						Type:     "image_url",
						ImageURL: &groqImageURL{URL: "data:" + mimeType + ";base64," + encoded},
					},
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported invoice file type: %s", mimeType)
	}
}

func extractPlainTextFromPDF(data []byte) (string, error) {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("open pdf: %w", err)
	}

	plainText, err := reader.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("read pdf text: %w", err)
	}

	content, err := io.ReadAll(plainText)
	if err != nil {
		return "", fmt.Errorf("read pdf plain text: %w", err)
	}

	return string(content), nil
}

func groqResponseText(response groqChatCompletionResponse) string {
	if len(response.Choices) == 0 {
		return ""
	}

	return strings.TrimSpace(response.Choices[0].Message.Content)
}

func parseInvoiceResult(raw string) (models.InvoiceParseResult, error) {
	jsonText, err := extractJSONObject(raw)
	if err != nil {
		return models.InvoiceParseResult{}, err
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(jsonText), &payload); err != nil {
		return models.InvoiceParseResult{}, fmt.Errorf("decode invoice json: %w", err)
	}

	result := models.InvoiceParseResult{}
	if value, ok := stringFromAny(payload["vendor"]); ok {
		result.Vendor = &value
	}
	if value, ok := stringFromAny(payload["expense_date"]); ok {
		value = strings.TrimSpace(value)
		if value != "" {
			result.ExpenseDate = &value
		}
	}
	if value, ok := stringFromAny(payload["truck_unit_number"]); ok {
		value = strings.TrimSpace(value)
		if value != "" {
			result.TruckUnitNumber = &value
		}
	}
	if value, ok := stringFromAny(payload["driver_name"]); ok {
		value = strings.TrimSpace(value)
		if value != "" {
			result.DriverName = &value
		}
	}
	if value, ok := floatFromAny(payload["amount"]); ok {
		result.Amount = &value
	}
	category := NormalizeMaintenanceCategory(stringFromAnyOrEmpty(payload["category"]))
	result.Category = &category
	if value, ok := stringFromAny(payload["description"]); ok {
		value = strings.TrimSpace(value)
		if value != "" {
			result.Description = &value
		}
	}
	if value, ok := stringFromAny(payload["reference_number"]); ok {
		value = strings.TrimSpace(value)
		if value != "" {
			result.ReferenceNumber = &value
		}
	}

	return result, nil
}

func extractJSONObject(raw string) (string, error) {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start == -1 || end == -1 || end <= start {
		return "", errors.New("groq response did not contain json")
	}

	return raw[start : end+1], nil
}

func stringFromAny(value any) (string, bool) {
	switch typed := value.(type) {
	case string:
		return typed, true
	case fmt.Stringer:
		return typed.String(), true
	case json.Number:
		return typed.String(), true
	default:
		return "", false
	}
}

func stringFromAnyOrEmpty(value any) string {
	result, _ := stringFromAny(value)
	return result
}

func floatFromAny(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case json.Number:
		parsed, err := typed.Float64()
		if err != nil {
			return 0, false
		}
		return parsed, true
	case string:
		trimmed := strings.TrimSpace(strings.TrimPrefix(typed, "$"))
		trimmed = strings.ReplaceAll(trimmed, ",", "")
		if trimmed == "" {
			return 0, false
		}
		parsed, err := json.Number(trimmed).Float64()
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}
