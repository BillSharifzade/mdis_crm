package repository

import (
	"context"
	"fmt"
	"time"

	"crm_backend/internal/model"
	"crm_backend/pkg/database"
)

type InteractionRepository struct {
	db *database.DB
}

func NewInteractionRepository(db *database.DB) *InteractionRepository {
	return &InteractionRepository{db: db}
}

func (r *InteractionRepository) Create(ctx context.Context, req *model.CreateInteractionRequest) (*model.Interaction, error) {
	direction := req.Direction
	if direction == "" {
		direction = "inbound"
	}
	query := `
		INSERT INTO interactions (
			lead_id, deal_id, type, direction, content, duration_seconds, duration_minutes, outcome, created_by, is_task, due_date, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, NULLIF($8,''), $9, $10, $11, $12
		) RETURNING id, created_at
	`
	now := time.Now()
	var interaction model.Interaction

	err := r.db.Pool.QueryRow(ctx, query,
		req.LeadID, req.DealID, req.Type, direction, req.Content, req.DurationSecond, req.DurationMinutes, req.Outcome,
		req.CreatedBy, req.IsTask, req.DueDate, now,
	).Scan(&interaction.ID, &interaction.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert interaction: %w", err)
	}

	interaction.LeadID = req.LeadID
	interaction.DealID = req.DealID
	interaction.Type = req.Type
	interaction.Direction = direction
	interaction.Content = req.Content
	interaction.DurationSecond = req.DurationSecond
	interaction.DurationMinutes = req.DurationMinutes
	interaction.Outcome = req.Outcome
	interaction.CreatedBy = req.CreatedBy
	interaction.IsTask = req.IsTask
	interaction.DueDate = req.DueDate

	return &interaction, nil
}

func (r *InteractionRepository) ListByLead(ctx context.Context, leadID int) ([]model.Interaction, error) {
	query := `
		SELECT id, lead_id, deal_id, type, direction, content, duration_seconds, duration_minutes,
		       COALESCE(outcome, ''), created_by, is_task, is_done, due_date, created_at
		FROM interactions
		WHERE lead_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, leadID)
	if err != nil {
		return nil, fmt.Errorf("failed to query interactions: %w", err)
	}
	defer rows.Close()

	var interactions []model.Interaction
	for rows.Next() {
		var i model.Interaction
		if err := rows.Scan(
			&i.ID, &i.LeadID, &i.DealID, &i.Type, &i.Direction, &i.Content, &i.DurationSecond, &i.DurationMinutes,
			&i.Outcome, &i.CreatedBy, &i.IsTask, &i.IsDone, &i.DueDate, &i.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}
		interactions = append(interactions, i)
	}
	return interactions, nil
}
