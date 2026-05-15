package repository

import (
	"context"
	"fmt"

	"crm_backend/internal/model"
	"crm_backend/pkg/database"
)

// KPIRepository — CRUD над целевыми KPI + расчёт фактических показателей
// активности менеджера приёма.
type KPIRepository struct {
	db *database.DB
}

func NewKPIRepository(db *database.DB) *KPIRepository { return &KPIRepository{db: db} }

func (r *KPIRepository) ListTargets(ctx context.Context, userID int) ([]model.KPITarget, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, user_id, metric, target_count, period_days, created_at, updated_at
		FROM kpi_targets WHERE user_id = $1 ORDER BY metric
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.KPITarget{}
	for rows.Next() {
		var t model.KPITarget
		if err := rows.Scan(&t.ID, &t.UserID, &t.Metric, &t.TargetCount, &t.PeriodDays, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

// UpsertTarget — есть запись (user_id, metric) → апдейт, иначе вставка.
func (r *KPIRepository) UpsertTarget(ctx context.Context, userID int, metric string, target, periodDays int) error {
	if metric != "processed" && metric != "created" {
		return fmt.Errorf("unsupported metric: %s", metric)
	}
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO kpi_targets (user_id, metric, target_count, period_days)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, metric) DO UPDATE
		  SET target_count = EXCLUDED.target_count,
		      period_days  = EXCLUDED.period_days,
		      updated_at   = NOW()
	`, userID, metric, target, periodDays)
	return err
}

func (r *KPIRepository) DeleteTarget(ctx context.Context, id int) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM kpi_targets WHERE id = $1`, id)
	return err
}

// GetStats считает фактические показатели за окно `periodDays`.
//   processed: число уникальных leads, по которым у пользователя есть interactions за окно
//   created:   число leads с assignee_id = user за окно (round-robin или ручное создание)
//   calls/notes/messages: разбивка по type interactions, сделанных пользователем за окно.
func (r *KPIRepository) GetStats(ctx context.Context, userID int, periodDays int) (*model.KPIStats, error) {
	if periodDays <= 0 {
		periodDays = 30
	}
	q := `
	WITH window_ts AS (SELECT NOW() - ($2::int * INTERVAL '1 day') AS since)
	SELECT
		(SELECT COUNT(DISTINCT lead_id) FROM interactions, window_ts
			WHERE created_by = $1 AND created_at >= window_ts.since AND lead_id IS NOT NULL) AS processed,
		(SELECT COUNT(*) FROM leads, window_ts
			WHERE assignee_id = $1 AND created_at >= window_ts.since) AS created,
		(SELECT COUNT(*) FROM interactions, window_ts
			WHERE created_by = $1 AND type = 'call' AND created_at >= window_ts.since) AS calls,
		(SELECT COUNT(*) FROM interactions, window_ts
			WHERE created_by = $1 AND type IN ('note', 'manual_note') AND created_at >= window_ts.since) AS notes,
		(SELECT COUNT(*) FROM interactions, window_ts
			WHERE created_by = $1 AND type IN ('messenger','sms','email','message') AND created_at >= window_ts.since) AS msgs
	`
	var s model.KPIStats
	s.UserID = userID
	s.PeriodDays = periodDays
	if err := r.db.Pool.QueryRow(ctx, q, userID, periodDays).Scan(
		&s.Processed, &s.Created, &s.CallsCount, &s.NotesCount, &s.MessagesCount,
	); err != nil {
		return nil, err
	}
	// Подтягиваем цели
	rows, err := r.db.Pool.Query(ctx, `SELECT metric, target_count FROM kpi_targets WHERE user_id = $1`, userID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var m string
			var t int
			if err := rows.Scan(&m, &t); err == nil {
				switch m {
				case "processed":
					s.ProcessedGoal = t
				case "created":
					s.CreatedGoal = t
				}
			}
		}
	}
	return &s, nil
}
