package model

import "time"

type Contact struct {
	ID         int       `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name,omitempty"`
	Email      *string   `json:"email,omitempty"`
	Phone      *string   `json:"phone,omitempty"`
	TelegramID *string   `json:"telegram_id,omitempty"`
	WhatsAppID *string   `json:"whatsapp_id,omitempty"`
	VKID       *string   `json:"vk_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
