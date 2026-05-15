package model

import "time"

type Interaction struct {
	ID              int        `json:"id"`
	LeadID          *int       `json:"lead_id,omitempty"`
	DealID          *int       `json:"deal_id,omitempty"`
	Type            string     `json:"type"`
	Direction       string     `json:"direction"` // inbound | outbound
	Content         string     `json:"content"`
	DurationSecond  *int       `json:"duration_seconds,omitempty"`
	DurationMinutes *int       `json:"duration_minutes,omitempty"`
	Outcome         string     `json:"outcome,omitempty"` // answered | no_answer (для type=call)
	CreatedBy       *int       `json:"created_by,omitempty"`
	IsTask          bool       `json:"is_task"`
	IsDone          bool       `json:"is_done"`
	DueDate         *time.Time `json:"due_date,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type CreateInteractionRequest struct {
	LeadID          *int       `json:"lead_id,omitempty"`
	DealID          *int       `json:"deal_id,omitempty"`
	Type            string     `json:"type"`
	Direction       string     `json:"direction,omitempty"`
	Content         string     `json:"content"`
	DurationSecond  *int       `json:"duration_seconds,omitempty"`
	DurationMinutes *int       `json:"duration_minutes,omitempty"`
	Outcome         string     `json:"outcome,omitempty"`
	CreatedBy       *int       `json:"created_by,omitempty"`
	IsTask          bool       `json:"is_task"`
	DueDate         *time.Time `json:"due_date,omitempty"`
}
