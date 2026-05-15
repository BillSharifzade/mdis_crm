package model

import (
	"time"
)

type Lead struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name,omitempty"`
	Email     string    `json:"email,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	SourceID  *int      `json:"source_id,omitempty"`
	ProgramID *int      `json:"program_id,omitempty"`
	StatusID  *int      `json:"status_id,omitempty"`
	AssigneeID *int     `json:"assignee_id,omitempty"`
	ContactID *int      `json:"contact_id,omitempty"`
	UTMSource   string    `json:"utm_source,omitempty"`
	UTMMedium   string    `json:"utm_medium,omitempty"`
	UTMCampaign string    `json:"utm_campaign,omitempty"`
	TelegramID  string    `json:"telegram_id,omitempty"`
	WhatsAppID  string    `json:"whatsapp_id,omitempty"`
	VKID        string    `json:"vk_id,omitempty"`
	SocialURL   string    `json:"social_url,omitempty"`
	ProgramName string    `json:"program_name,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Program struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Faculty  string `json:"faculty,omitempty"`
	IsActive bool   `json:"is_active"`
}

type CreateLeadRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	UTMSource string `json:"utm_source"`
	UTMMedium string `json:"utm_medium"`
	UTMCampaign string `json:"utm_campaign"`
	ProgramID   *int   `json:"program_id"`
	SourceID    *int   `json:"source_id,omitempty"`
	StatusID    *int   `json:"status_id,omitempty"`
	TelegramID  string `json:"telegram_id,omitempty"`
	WhatsAppID  string `json:"whatsapp_id,omitempty"`
	VKID        string `json:"vk_id,omitempty" example:"vk123"`
	SocialURL   string `json:"social_url,omitempty"`
	AssigneeID  *int   `json:"assignee_id,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
}

type UpdateLeadRequest struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	ProgramID  *int   `json:"program_id,omitempty"`
	AssigneeID *int   `json:"assignee_id,omitempty"`
	UTMSource  string `json:"utm_source,omitempty"`
	SocialURL  *string `json:"social_url,omitempty"`
	ContactID  *int   `json:"-"`
}

type PipelineStage struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Order    int    `json:"order"`
	IsFinal  bool   `json:"is_final"`
	IsActive bool   `json:"is_active"`
}

type Source struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}
