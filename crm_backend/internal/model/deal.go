package model

import "time"

type Deal struct {
	ID            int       `json:"id"`
	ContactID     int       `json:"contact_id"`
	StageID       int       `json:"stage_id"`
	AssigneeID    *int      `json:"assignee_id,omitempty"`
	SourceID      *int      `json:"source_id,omitempty"`
	RefusalReason string     `json:"refusal_reason,omitempty"`
	ExamDate      *time.Time `json:"exam_date,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
