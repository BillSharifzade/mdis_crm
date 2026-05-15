package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client — минимальный клиент Telegram Bot API для отправки сообщений
// и подписки на webhook. Достаточно для нужд CRM (sendMessage, setWebhook).
type Client struct {
	Token string
	HTTP  *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		Token: token,
		HTTP:  &http.Client{Timeout: 10 * time.Second},
	}
}

type sendMessagePayload struct {
	ChatID      int64       `json:"chat_id"`
	Text        string      `json:"text"`
	ParseMode   string      `json:"parse_mode,omitempty"`
	ReplyMarkup interface{} `json:"reply_markup,omitempty"`
}

// InlineButton — кнопка под сообщением. Можно слать с callback_data.
type InlineButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
	URL          string `json:"url,omitempty"`
}

type InlineKeyboard struct {
	InlineKeyboard [][]InlineButton `json:"inline_keyboard"`
}

type apiResponse struct {
	OK          bool            `json:"ok"`
	Description string          `json:"description,omitempty"`
	Result      json.RawMessage `json:"result,omitempty"`
}

func (c *Client) call(ctx context.Context, method string, body any) error {
	if c.Token == "" {
		return fmt.Errorf("telegram bot token is empty")
	}
	buf, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", c.Token, method)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("telegram api: %w", err)
	}
	defer resp.Body.Close()
	var r apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	if !r.OK {
		return fmt.Errorf("telegram api %s: %s", method, r.Description)
	}
	return nil
}

// SendMessage отправляет текстовое сообщение в чат с указанным chatID.
func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	return c.call(ctx, "sendMessage", sendMessagePayload{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "HTML",
	})
}

// SendMessageWithKeyboard отправляет сообщение с inline-кнопками под ним.
// Каждая внутренняя группа `[]InlineButton` рендерится одним рядом.
func (c *Client) SendMessageWithKeyboard(ctx context.Context, chatID int64, text string, rows [][]InlineButton) error {
	return c.call(ctx, "sendMessage", sendMessagePayload{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   "HTML",
		ReplyMarkup: InlineKeyboard{InlineKeyboard: rows},
	})
}

// AnswerCallbackQuery отзывает на клик кнопки (убирает "часики" у клиента).
func (c *Client) AnswerCallbackQuery(ctx context.Context, callbackID, text string) error {
	body := map[string]any{"callback_query_id": callbackID}
	if text != "" {
		body["text"] = text
	}
	return c.call(ctx, "answerCallbackQuery", body)
}

// SetWebhook регистрирует URL-вебхук для входящих сообщений.
// Вызывать вручную или при старте сервера, если задан PUBLIC_URL.
func (c *Client) SetWebhook(ctx context.Context, url string) error {
	return c.call(ctx, "setWebhook", map[string]any{
		"url":             url,
		"allowed_updates": []string{"message", "callback_query"},
	})
}

// DeleteWebhook сбрасывает зарегистрированный вебхук.
func (c *Client) DeleteWebhook(ctx context.Context) error {
	return c.call(ctx, "deleteWebhook", map[string]any{})
}
