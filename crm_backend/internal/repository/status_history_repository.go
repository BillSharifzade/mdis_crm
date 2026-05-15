package repository

import (
	"context"
	"fmt"
	"time"

	"crm_backend/internal/model"
	"crm_backend/pkg/database"
)

type StatusHistoryRepository struct {
	db *database.DB
}

func NewStatusHistoryRepository(db *database.DB) *StatusHistoryRepository {
	return &StatusHistoryRepository{db: db}
}

func (r *StatusHistoryRepository) Create(ctx context.Context, leadID int, oldStatusID *int, newStatusID int, changedBy *int, refusalReason string) error {
	query := `
		INSERT INTO status_history (lead_id, old_status_id, new_status_id, changed_by, refusal_reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Pool.Exec(ctx, query, leadID, oldStatusID, newStatusID, changedBy, refusalReason, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create status history: %w", err)
	}
	return nil
}

func (r *StatusHistoryRepository) GetByLeadID(ctx context.Context, leadID int) ([]model.StatusHistory, error) {
	query := `
		SELECT id, lead_id, old_status_id, new_status_id, changed_by, refusal_reason, created_at
		FROM status_history
		WHERE lead_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, leadID)
	if err != nil {
		return nil, fmt.Errorf("failed to query status history: %w", err)
	}
	defer rows.Close()

	var history []model.StatusHistory
	for rows.Next() {
		var h model.StatusHistory
		if err := rows.Scan(&h.ID, &h.LeadID, &h.OldStatusID, &h.NewStatusID, &h.ChangedBy, &h.RefusalReason, &h.CreatedAt); err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}
		history = append(history, h)
	}
	return history, nil
}

// HasRecentTask checks if a task was already created today for a given lead
func (r *StatusHistoryRepository) HasRecentTask(ctx context.Context, leadID int) (bool, error) {
	// We use the interactions table for this check
	query := `
		SELECT EXISTS(
			SELECT 1 FROM interactions
			WHERE lead_id = $1 AND is_task = TRUE AND created_at >= CURRENT_DATE
		)
	`
	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, leadID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
