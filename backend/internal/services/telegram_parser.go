package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type telegramMaintenanceParseResult struct {
	TruckUnitNumber *string  `json:"truck_unit_number,omitempty"`
	Amount          *float64 `json:"amount,omitempty"`
	ReferenceNumber *string  `json:"reference_number,omitempty"`
	Category        *string  `json:"category,omitempty"`
	Description     *string  `json:"description,omitempty"`
	WhoCovers       *string  `json:"who_covers,omitempty"`
}

const (
	telegramMaintenanceParsePrompt = `You are parsing a Telegram message from a trucking fleet accounting group.
Extract the following fields and return ONLY valid JSON, no prose:
{
  "truck_unit_number": "unit number like 022 or 2 or null",
  "amount": number or null,
  "reference_number": "report reference or transaction number or null",
  "category": one of [PM Service, Oil change, Tire issue, Engine issue, Towing, Road Service, Body work, Leakage, Kris Shop, Truck Wash/Detailing, Electrical issue, Fluids/Truck Parts, Brakes/Drums/Rotors, Scale, Other],
  "description": "brief normalized description of the work done or null",
  "who_covers": one of [Company, Owner O] or null
}
Rules:
- Ignore trailing @usernames or accountant sign-offs.
- If the message says owner operator, owner op, need to charge owner, or similar, use "Owner O".
- Otherwise use "Company".
- Preserve the truck unit exactly as written, including leading zeros.
- If a field cannot be determined, use null.`
	telegarmMaintenanceTextPrompt = "Extract the maintenance transaction fields from this Telegram message and return only JSON."
	telegarmMaintenanceModel      = groqInvoiceModel
)

var (
	telegramMoneyTransferCodePattern = regexp.MustCompile(`(?i)\bMoney Transfer code\s*[:#]?\s*([A-Za-z0-9-]+)`)
	telegramMoneyTransferLinePattern = regexp.MustCompile(`(?i)^\s*Money Transfer code\b.*$`)
	telegramReportReferencePattern   = regexp.MustCompile(`(?i)\bReport Reference\s*#?\s*([A-Za-z0-9-]+)`)
	telegramAmountPattern            = regexp.MustCompile(`(?i)\bAmount\s*\$?\s*([0-9][0-9,]*(?:\.[0-9]{2})?)`)
	telegramIssuedToPattern          = regexp.MustCompile(`(?i)\bIssued to\s*([A-Za-z0-9-]+)`)
	telegramNotesPattern             = regexp.MustCompile(`(?is)\bNotes\s*(.*)$`)
)

func splitTelegramMaintenanceBlocks(rawText string) []string {
	cleanedText := cleanTelegramMessageText(rawText)
	if cleanedText == "" {
		return nil
	}

	lines := strings.Split(cleanedText, "\n")
	blockStarts := make([]int, 0, len(lines))
	for index, line := range lines {
		if telegramMoneyTransferLinePattern.MatchString(line) {
			blockStarts = append(blockStarts, index)
		}
	}

	if len(blockStarts) == 0 {
		return []string{cleanedText}
	}

	blocks := make([]string, 0, len(blockStarts))
	for index, start := range blockStarts {
		end := len(lines)
		if index+1 < len(blockStarts) {
			end = blockStarts[index+1]
		}

		block := strings.TrimSpace(strings.Join(lines[start:end], "\n"))
		if block != "" {
			blocks = append(blocks, block)
		}
	}

	if len(blocks) == 0 {
		return []string{cleanedText}
	}

	return blocks
}

func parseTelegramMaintenance(ctx context.Context, token, rawText string) (telegramMaintenanceParseResult, error) {
	if strings.TrimSpace(token) == "" {
		return telegramMaintenanceParseResult{}, errors.New("GROQ_TOKEN is not configured")
	}

	cleanedText := cleanTelegramMessageText(rawText)
	if cleanedText == "" {
		return telegramMaintenanceParseResult{}, errors.New("telegram message text was empty")
	}

	requestBody := groqChatCompletionRequest{
		Model: telegarmMaintenanceModel,
		Messages: []groqMessage{
			{Role: "system", Content: telegramMaintenanceParsePrompt},
			{Role: "user", Content: telegarmMaintenanceTextPrompt + "\n\nTelegram message:\n" + cleanedText},
		},
		ResponseFormat: groqResponseFmt{Type: "json_object"},
		Temperature:    0,
		TopP:           1,
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return telegramMaintenanceParseResult{}, fmt.Errorf("marshal groq request: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.groq.com/openai/v1/chat/completions",
		bytes.NewReader(payload),
	)
	if err != nil {
		return telegramMaintenanceParseResult{}, fmt.Errorf("build groq request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 90 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return telegramMaintenanceParseResult{}, fmt.Errorf("call groq: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return telegramMaintenanceParseResult{}, fmt.Errorf("read groq response: %w", err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return telegramMaintenanceParseResult{}, &UpstreamStatusError{
			StatusCode: response.StatusCode,
			Message:    fmt.Sprintf("groq returned %s: %s", response.Status, strings.TrimSpace(string(body))),
		}
	}

	var parsed groqChatCompletionResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return telegramMaintenanceParseResult{}, fmt.Errorf("decode groq response: %w", err)
	}

	textResponse := groqResponseText(parsed)
	if textResponse == "" {
		return telegramMaintenanceParseResult{}, errors.New("groq response did not include parsed text")
	}

	result, err := parseTelegramMaintenanceResult(textResponse)
	if err != nil {
		return telegramMaintenanceParseResult{}, err
	}

	return enrichTelegramMaintenanceResult(result, cleanedText), nil
}

func parseTelegramMaintenanceResult(raw string) (telegramMaintenanceParseResult, error) {
	jsonText, err := extractJSONObject(raw)
	if err != nil {
		return telegramMaintenanceParseResult{}, err
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(jsonText), &payload); err != nil {
		return telegramMaintenanceParseResult{}, fmt.Errorf("decode telegram json: %w", err)
	}

	result := telegramMaintenanceParseResult{}
	if value, ok := stringFromAny(payload["truck_unit_number"]); ok {
		value = strings.TrimSpace(value)
		if value != "" {
			result.TruckUnitNumber = &value
		}
	}
	if value, ok := floatFromAny(payload["amount"]); ok {
		result.Amount = &value
	}
	if value, ok := stringFromAny(payload["reference_number"]); ok {
		value = strings.TrimSpace(value)
		if value != "" {
			result.ReferenceNumber = &value
		}
	}
	if value, ok := stringFromAny(payload["category"]); ok {
		value = NormalizeMaintenanceCategory(value)
		if value != "" {
			result.Category = &value
		}
	}
	if value, ok := stringFromAny(payload["description"]); ok {
		value = strings.TrimSpace(value)
		if value != "" {
			result.Description = &value
		}
	}
	if value, ok := stringFromAny(payload["who_covers"]); ok {
		value = normalizeWhoCovers(value)
		if value != "" {
			result.WhoCovers = &value
		}
	}

	return result, nil
}

func enrichTelegramMaintenanceResult(result telegramMaintenanceParseResult, cleanedText string) telegramMaintenanceParseResult {
	if result.TruckUnitNumber == nil {
		if value := firstTelegramPatternMatch(telegramIssuedToPattern, cleanedText); value != "" {
			result.TruckUnitNumber = &value
		}
	}
	if result.Amount == nil {
		if value := firstTelegramPatternMatch(telegramAmountPattern, cleanedText); value != "" {
			if parsedAmount, ok := floatFromAny(value); ok {
				result.Amount = &parsedAmount
			}
		}
	}
	if result.ReferenceNumber == nil {
		if value := firstTelegramPatternMatch(telegramReportReferencePattern, cleanedText); value != "" {
			result.ReferenceNumber = &value
		} else if value := firstTelegramPatternMatch(telegramMoneyTransferCodePattern, cleanedText); value != "" {
			result.ReferenceNumber = &value
		}
	}
	if result.Description == nil {
		if value := extractTelegramNotes(cleanedText); value != "" {
			result.Description = &value
		}
	}
	if result.Category == nil {
		category := inferTelegramCategory(cleanedText)
		result.Category = &category
	}
	if result.WhoCovers == nil {
		whoCovers := inferTelegramWhoCovers(cleanedText)
		result.WhoCovers = &whoCovers
	}

	return result
}

func cleanTelegramMessageText(text string) string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || isUsernameOnlyLine(trimmed) {
			continue
		}
		cleaned = append(cleaned, trimmed)
	}

	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

func extractTelegramNotes(text string) string {
	match := telegramNotesPattern.FindStringSubmatch(text)
	if len(match) < 2 {
		return ""
	}

	notes := strings.TrimSpace(match[1])
	notes = strings.TrimSpace(strings.TrimSuffix(notes, ","))
	notes = strings.TrimSpace(strings.TrimSuffix(notes, "."))
	return notes
}

func inferTelegramCategory(text string) string {
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "tire"):
		return "Tire issue"
	case strings.Contains(lower, "engine"):
		return "Engine issue"
	case strings.Contains(lower, "tow"):
		return "Towing"
	case strings.Contains(lower, "road service"):
		return "Road Service"
	case strings.Contains(lower, "body"):
		return "Body work"
	case strings.Contains(lower, "leak"):
		return "Leakage"
	case strings.Contains(lower, "electr"):
		return "Electrical issue"
	case strings.Contains(lower, "fluid") || strings.Contains(lower, "part"):
		return "Fluids/Truck Parts"
	case strings.Contains(lower, "brake"):
		return "Brakes/Drums/Rotors"
	case strings.Contains(lower, "scale"):
		return "Scale"
	case strings.Contains(lower, "wash") || strings.Contains(lower, "detail"):
		return "Truck Wash/Detailing"
	case strings.Contains(lower, "oil"):
		return "Oil change"
	default:
		return "Other"
	}
}

func inferTelegramWhoCovers(text string) string {
	lower := strings.ToLower(text)
	if strings.Contains(lower, "need to charge owner") || strings.Contains(lower, "charge owner") || strings.Contains(lower, "owner operator") || strings.Contains(lower, "owner op") || strings.Contains(lower, "owner operators") {
		return "Owner O"
	}

	return "Company"
}

func normalizeWhoCovers(value string) string {
	key := strings.ToUpper(strings.TrimSpace(value))
	key = strings.NewReplacer(",", " ", "-", " ", "_", " ", "/", " ").Replace(key)
	key = strings.Join(strings.Fields(key), " ")

	switch {
	case key == "COMPANY":
		return "Company"
	case key == "OWNER O", key == "OWNER OP", key == "OWNER OPERATOR", key == "OWNER OPERATORS":
		return "Owner O"
	default:
		if strings.Contains(key, "OWNER") {
			return "Owner O"
		}
		if strings.Contains(key, "COMPANY") {
			return "Company"
		}
		return strings.TrimSpace(value)
	}
}

func firstTelegramPatternMatch(pattern *regexp.Regexp, text string) string {
	match := pattern.FindStringSubmatch(text)
	if len(match) < 2 {
		return ""
	}

	return strings.TrimSpace(match[1])
}

func isUsernameOnlyLine(line string) bool {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return false
	}

	for _, field := range fields {
		if !strings.HasPrefix(field, "@") {
			return false
		}
	}

	return true
}
