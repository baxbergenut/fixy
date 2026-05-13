package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"fixy-backend/internal/db"
)

type TelegramEFSBot struct {
	database      *sql.DB
	groqToken     string
	telegramToken string
}

type telegramUpdatesResponse struct {
	OK          bool             `json:"ok"`
	Description string           `json:"description"`
	Result      []telegramUpdate `json:"result"`
}

type telegramUpdate struct {
	UpdateID int64            `json:"update_id"`
	Message  *telegramMessage `json:"message,omitempty"`
}

type telegramMessage struct {
	MessageID int           `json:"message_id"`
	From      *telegramUser `json:"from,omitempty"`
	Chat      telegramChat  `json:"chat"`
	Text      string        `json:"text,omitempty"`
}

type telegramUser struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
}

type telegramChat struct {
	ID       int64  `json:"id"`
	Type     string `json:"type"`
	Title    string `json:"title,omitempty"`
	UserName string `json:"username,omitempty"`
}

func NewTelegramEFSBot(database *sql.DB, groqToken, telegramToken string) *TelegramEFSBot {
	return &TelegramEFSBot{
		database:      database,
		groqToken:     strings.TrimSpace(groqToken),
		telegramToken: strings.TrimSpace(telegramToken),
	}
}

func (bot *TelegramEFSBot) Run(ctx context.Context) {
	if bot == nil {
		return
	}
	if bot.database == nil {
		log.Printf("telegram bot disabled: database not configured")
		return
	}
	if strings.TrimSpace(bot.telegramToken) == "" {
		log.Printf("telegram bot disabled: TELEGRAM_BOT_TOKEN is not configured")
		return
	}
	if strings.TrimSpace(bot.groqToken) == "" {
		log.Printf("telegram bot disabled: GROQ_TOKEN is not configured")
		return
	}

	log.Printf("telegram bot polling started")

	var offset int64
	for {
		if ctx.Err() != nil {
			return
		}

		updates, err := bot.fetchUpdates(ctx, offset)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("telegram bot poll error: %v", err)
			select {
			case <-time.After(5 * time.Second):
			case <-ctx.Done():
				return
			}
			continue
		}

		for _, update := range updates {
			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}

			if err := bot.handleUpdate(ctx, update); err != nil {
				log.Printf("telegram bot update %d: %v", update.UpdateID, err)
			}
		}
	}
}

func (bot *TelegramEFSBot) fetchUpdates(ctx context.Context, offset int64) ([]telegramUpdate, error) {
	form := url.Values{}
	form.Set("timeout", strconv.Itoa(30))
	if offset > 0 {
		form.Set("offset", strconv.FormatInt(offset, 10))
	}

	requestCtx, cancel := context.WithTimeout(ctx, 40*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(requestCtx, http.MethodPost, bot.telegramAPIURL("getUpdates"), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build telegram request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("call telegram: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read telegram response: %w", err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("telegram returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}

	var payload telegramUpdatesResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode telegram response: %w", err)
	}

	if !payload.OK {
		message := strings.TrimSpace(payload.Description)
		if message == "" {
			message = "telegram returned an error"
		}
		return nil, errors.New(message)
	}

	return payload.Result, nil
}

func (bot *TelegramEFSBot) handleUpdate(ctx context.Context, update telegramUpdate) error {
	if update.Message == nil {
		return nil
	}
	if update.Message.From == nil {
		return nil
	}
	if !isTelegramGroupChat(update.Message.Chat.Type) {
		return nil
	}

	messageText := cleanTelegramMessageText(update.Message.Text)
	if messageText == "" || !looksLikeEFSMessage(messageText) {
		return nil
	}

	blocks := splitTelegramMaintenanceBlocks(messageText)
	var insertedCount int
	var blockErrors []error
	for index, block := range blocks {
		if err := bot.insertMaintenanceBlock(ctx, update.Message, block); err != nil {
			blockErrors = append(blockErrors, fmt.Errorf("block %d: %w", index+1, err))
			continue
		}
		insertedCount++
	}

	if insertedCount == 0 && len(blockErrors) > 0 {
		return errors.Join(blockErrors...)
	}
	if len(blockErrors) > 0 {
		log.Printf("telegram bot inserted %d maintenance log(s) with %d block error(s)", insertedCount, len(blockErrors))
	}

	return nil
}

func (bot *TelegramEFSBot) insertMaintenanceBlock(ctx context.Context, message *telegramMessage, block string) error {
	parsed, err := parseTelegramMaintenance(ctx, bot.groqToken, block)
	if err != nil {
		return err
	}
	if parsed.Amount == nil {
		return errors.New("telegram block did not include an amount")
	}
	if parsed.TruckUnitNumber == nil || strings.TrimSpace(*parsed.TruckUnitNumber) == "" {
		return errors.New("telegram block did not include a truck unit")
	}

	truckID, err := bot.lookupTruckID(ctx, *parsed.TruckUnitNumber)
	if err != nil {
		return err
	}
	if strings.TrimSpace(truckID) == "" {
		return fmt.Errorf("truck unit %q was not found", strings.TrimSpace(*parsed.TruckUnitNumber))
	}

	description := ""
	if parsed.Description != nil {
		description = strings.TrimSpace(*parsed.Description)
	}
	if description == "" {
		description = block
	}

	referenceNumber := ""
	if parsed.ReferenceNumber != nil {
		referenceNumber = strings.TrimSpace(*parsed.ReferenceNumber)
	}

	whoCovers := inferTelegramWhoCovers(block)
	if parsed.WhoCovers != nil && strings.TrimSpace(*parsed.WhoCovers) != "" {
		whoCovers = normalizeWhoCovers(*parsed.WhoCovers)
	}

	category := "Other"
	if parsed.Category != nil && strings.TrimSpace(*parsed.Category) != "" {
		category = NormalizeMaintenanceCategory(*parsed.Category)
	}

	var sender *telegramUser
	if message != nil {
		sender = message.From
	}
	paidBy := telegramSenderName(sender)
	expenseDate := time.Now().Format("2006-01-02")
	telegramMessage := strings.TrimSpace(block)

	row := bot.database.QueryRowContext(ctx, db.InsertMaintenanceLogQuery,
		truckID,
		nil,
		expenseDate,
		nil,
		nil,
		*parsed.Amount,
		category,
		"EFS",
		description,
		nullableString(referenceNumber),
		nullableString(whoCovers),
		nullableString(paidBy),
		nullableString(telegramMessage),
		false,
		false,
		nil,
	)

	var createdID string
	if err := row.Scan(&createdID); err != nil {
		return fmt.Errorf("insert maintenance log: %w", err)
	}

	log.Printf("telegram bot inserted maintenance log %s for unit %s", createdID, strings.TrimSpace(*parsed.TruckUnitNumber))
	return nil
}

func (bot *TelegramEFSBot) lookupTruckID(ctx context.Context, unitNumber string) (string, error) {
	row := bot.database.QueryRowContext(ctx, db.GetTruckIDByUnitNumberQuery, strings.TrimSpace(unitNumber))

	var truckID string
	if err := row.Scan(&truckID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("lookup truck %q: %w", unitNumber, err)
	}

	return truckID, nil
}

func (bot *TelegramEFSBot) telegramAPIURL(method string) string {
	return "https://api.telegram.org/bot" + bot.telegramToken + "/" + method
}

func looksLikeEFSMessage(text string) bool {
	lower := strings.ToLower(text)
	return strings.Contains(lower, "money transfer") || strings.Contains(lower, "report reference") || strings.Contains(lower, "issued to") || strings.Contains(lower, "need to charge owner")
}

func telegramSenderName(user *telegramUser) string {
	if user == nil {
		return "Telegram user"
	}

	first := strings.TrimSpace(user.FirstName)
	last := strings.TrimSpace(user.LastName)
	if first != "" || last != "" {
		return strings.TrimSpace(strings.Join([]string{first, last}, " "))
	}

	username := strings.TrimSpace(user.Username)
	if username != "" {
		return "@" + username
	}

	return "Telegram user"
}

func nullableString(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	return trimmed
}

func isTelegramGroupChat(chatType string) bool {
	switch strings.ToLower(strings.TrimSpace(chatType)) {
	case "group", "supergroup":
		return true
	default:
		return false
	}
}
