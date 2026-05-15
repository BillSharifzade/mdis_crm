package model

import "time"

type StatusHistory struct {
	ID            int       `json:"id"`
	LeadID        int       `json:"lead_id"`
	OldStatusID   *int      `json:"old_status_id,omitempty"`
	NewStatusID   int       `json:"new_status_id"`
	ChangedBy     *int      `json:"changed_by,omitempty"`
	RefusalReason string    `json:"refusal_reason,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // admin, admissions, guest
}

type UpdateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"` // optional — empty means keep current
	Role     string `json:"role"`
}
