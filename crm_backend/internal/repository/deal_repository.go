package repository

import (
	"context"
	"fmt"
	"time"

	"crm_backend/internal/model"
	"crm_backend/pkg/database"
)

type DealRepository struct {
	db *database.DB
}

func NewDealRepository(db *database.DB) *DealRepository {
	return &DealRepository{db: db}
}

func (r *DealRepository) Create(ctx context.Context, contactID, stageID, sourceID int) (*model.Deal, error) {
	query := `
		INSERT INTO deals (contact_id, stage_id, source_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	var deal model.Deal
	
	err := r.db.Pool.QueryRow(ctx, query,
		contactID, stageID, sourceID, now, now,
	).Scan(&deal.ID, &deal.CreatedAt, &deal.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create deal: %w", err)
	}

	deal.ContactID = contactID
	deal.StageID = stageID
	sourceIDPtr := sourceID
	deal.SourceID = &sourceIDPtr

	return &deal, nil
}

func (r *DealRepository) UpdateStatusByContactID(ctx context.Context, contactID int, stageID int, refusalReason string, examDate *time.Time) error {
	query := `UPDATE deals SET stage_id = $1, refusal_reason = $2, exam_date = $3, updated_at = $4 WHERE contact_id = $5`
	_, err := r.db.Pool.Exec(ctx, query, stageID, refusalReason, examDate, time.Now(), contactID)
	if err != nil {
		return fmt.Errorf("failed to update deal status: %w", err)
	}
	return nil
}
