package model

import "time"

// TelegramChat хранит состояние диалога с лидом через Telegram-бота.
// bot_state — текущий шаг сценария; bot_active=false означает, что менеджер перехватил чат.
type TelegramChat struct {
	ID               int       `json:"id"`
	ChatID           int64     `json:"chat_id"`
	LeadID           *int      `json:"lead_id,omitempty"`
	BotState         string    `json:"bot_state"`
	BotActive        bool      `json:"bot_active"`
	TGUsername       string    `json:"tg_username,omitempty"`
	TGFirstName      string    `json:"tg_first_name,omitempty"`
	CollectedName    string    `json:"collected_name,omitempty"`
	CollectedProgram string    `json:"collected_program,omitempty"`
	CollectedEnglish string    `json:"collected_english,omitempty"`
	CollectedPhone   string    `json:"collected_phone,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Состояния сценария бота
const (
	BotStateGreet      = "greet"
	BotStateMenu       = "menu" // Главное меню с FAQ + "Подать заявку"
	BotStateAskName    = "ask_name"
	BotStateAskProgram = "ask_program"
	BotStateAskEnglish = "ask_english"
	BotStateAskPhone   = "ask_phone"
	BotStateManager    = "manager"
)

type SendTelegramRequest struct {
	Text string `json:"text"`
}
