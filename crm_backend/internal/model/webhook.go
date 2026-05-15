package model

type TelegramUser struct {
	ID           int    `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
}

type TelegramChatInfo struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
	Type      string `json:"type"`
}

type TelegramMessage struct {
	MessageID int              `json:"message_id"`
	From      TelegramUser     `json:"from"`
	Chat      TelegramChatInfo `json:"chat"`
	Date      int              `json:"date"`
	Text      string           `json:"text"`
}

type TelegramCallbackQuery struct {
	ID      string           `json:"id"`
	From    TelegramUser     `json:"from"`
	Message *TelegramMessage `json:"message,omitempty"`
	Data    string           `json:"data"`
}

type TelegramWebhookRequest struct {
	UpdateID      int                    `json:"update_id"`
	Message       TelegramMessage        `json:"message"`
	CallbackQuery *TelegramCallbackQuery `json:"callback_query,omitempty"`
}

type TelephonyWebhookRequest struct {
	CallID    string `json:"call_id"`
	Caller    string `json:"caller"`
	Called    string `json:"called"`
	Status    string `json:"status"`
	Duration  int    `json:"duration"`
	Recording string `json:"recording_url"`
}
