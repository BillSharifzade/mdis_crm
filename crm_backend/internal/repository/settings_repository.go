package repository

import (
	"context"
	"fmt"

	"crm_backend/internal/model"
	"crm_backend/pkg/database"
)

// SettingsRepository обслуживает CRUD над справочниками:
// programs (faculties), sources (integrations / откуда пришёл лид),
// pipeline_stages (этапы воронки).
// Если запись где-то используется (FK), мы не удаляем, а помечаем is_active=false.
type SettingsRepository struct {
	db *database.DB
}

func NewSettingsRepository(db *database.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// ── programs ───────────────────────────────────────────────────────────

// ActiveNames возвращает имена активных программ для inline-клавиатуры бота.
func (r *SettingsRepository) ActiveNames(ctx context.Context) ([]string, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT name FROM programs WHERE is_active = TRUE ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}

func (r *SettingsRepository) ListPrograms(ctx context.Context, onlyActive bool) ([]model.Program, error) {
	q := `SELECT id, name, COALESCE(faculty, ''), is_active FROM programs`
	if onlyActive {
		q += ` WHERE is_active = TRUE`
	}
	q += ` ORDER BY is_active DESC, name`
	rows, err := r.db.Pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Program{}
	for rows.Next() {
		var p model.Program
		if err := rows.Scan(&p.ID, &p.Name, &p.Faculty, &p.IsActive); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *SettingsRepository) CreateProgram(ctx context.Context, name, faculty string) (*model.Program, error) {
	var p model.Program
	err := r.db.Pool.QueryRow(ctx, `
		INSERT INTO programs (name, faculty, is_active) VALUES ($1, NULLIF($2,''), TRUE)
		RETURNING id, name, COALESCE(faculty,''), is_active
	`, name, faculty).Scan(&p.ID, &p.Name, &p.Faculty, &p.IsActive)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *SettingsRepository) UpdateProgram(ctx context.Context, id int, name, faculty string, isActive *bool) error {
	q := `UPDATE programs SET name = $1, faculty = NULLIF($2,''), updated_at = NOW()`
	args := []interface{}{name, faculty}
	if isActive != nil {
		q += `, is_active = $3 WHERE id = $4`
		args = append(args, *isActive, id)
	} else {
		q += ` WHERE id = $3`
		args = append(args, id)
	}
	_, err := r.db.Pool.Exec(ctx, q, args...)
	return err
}

func (r *SettingsRepository) DeleteProgram(ctx context.Context, id int) error {
	var refs int
	_ = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM leads WHERE program_id = $1`, id).Scan(&refs)
	if refs > 0 {
		_, err := r.db.Pool.Exec(ctx, `UPDATE programs SET is_active = FALSE, updated_at = NOW() WHERE id = $1`, id)
		if err != nil {
			return err
		}
		return fmt.Errorf("program is in use, archived instead (refs=%d)", refs)
	}
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM programs WHERE id = $1`, id)
	return err
}

// ── sources (integrations / lead sources) ──────────────────────────────

func (r *SettingsRepository) ListSources(ctx context.Context, onlyActive bool) ([]model.Source, error) {
	q := `SELECT id, name, is_active FROM sources`
	if onlyActive {
		q += ` WHERE is_active = TRUE`
	}
	q += ` ORDER BY is_active DESC, name`
	rows, err := r.db.Pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Source{}
	for rows.Next() {
		var s model.Source
		if err := rows.Scan(&s.ID, &s.Name, &s.IsActive); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func (r *SettingsRepository) CreateSource(ctx context.Context, name string) (*model.Source, error) {
	var s model.Source
	err := r.db.Pool.QueryRow(ctx, `
		INSERT INTO sources (name, is_active) VALUES ($1, TRUE)
		ON CONFLICT (name) DO UPDATE SET is_active = TRUE
		RETURNING id, name, is_active
	`, name).Scan(&s.ID, &s.Name, &s.IsActive)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SettingsRepository) UpdateSource(ctx context.Context, id int, name string, isActive *bool) error {
	q := `UPDATE sources SET name = $1, updated_at = NOW()`
	args := []interface{}{name}
	if isActive != nil {
		q += `, is_active = $2 WHERE id = $3`
		args = append(args, *isActive, id)
	} else {
		q += ` WHERE id = $2`
		args = append(args, id)
	}
	_, err := r.db.Pool.Exec(ctx, q, args...)
	return err
}

func (r *SettingsRepository) DeleteSource(ctx context.Context, id int) error {
	var refs int
	_ = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM leads WHERE source_id = $1`, id).Scan(&refs)
	if refs > 0 {
		_, err := r.db.Pool.Exec(ctx, `UPDATE sources SET is_active = FALSE, updated_at = NOW() WHERE id = $1`, id)
		if err != nil {
			return err
		}
		return fmt.Errorf("source is in use, archived instead (refs=%d)", refs)
	}
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM sources WHERE id = $1`, id)
	return err
}

// ── pipeline stages ────────────────────────────────────────────────────

func (r *SettingsRepository) ListStages(ctx context.Context, onlyActive bool) ([]model.PipelineStage, error) {
	q := `SELECT id, name, "order", is_final, is_active FROM pipeline_stages`
	if onlyActive {
		q += ` WHERE is_active = TRUE`
	}
	q += ` ORDER BY "order"`
	rows, err := r.db.Pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.PipelineStage{}
	for rows.Next() {
		var s model.PipelineStage
		if err := rows.Scan(&s.ID, &s.Name, &s.Order, &s.IsFinal, &s.IsActive); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func (r *SettingsRepository) CreateStage(ctx context.Context, name string, isFinal bool) (*model.PipelineStage, error) {
	var maxOrder int
	_ = r.db.Pool.QueryRow(ctx, `SELECT COALESCE(MAX("order"),0) FROM pipeline_stages`).Scan(&maxOrder)
	var s model.PipelineStage
	err := r.db.Pool.QueryRow(ctx, `
		INSERT INTO pipeline_stages (name, "order", is_final, is_active) VALUES ($1, $2, $3, TRUE)
		RETURNING id, name, "order", is_final, is_active
	`, name, maxOrder+1, isFinal).Scan(&s.ID, &s.Name, &s.Order, &s.IsFinal, &s.IsActive)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SettingsRepository) UpdateStage(ctx context.Context, id int, name string, order *int, isFinal, isActive *bool) error {
	q := `UPDATE pipeline_stages SET name = $1, updated_at = NOW()`
	args := []interface{}{name}
	idx := 2
	if order != nil {
		q += fmt.Sprintf(`, "order" = $%d`, idx)
		args = append(args, *order)
		idx++
	}
	if isFinal != nil {
		q += fmt.Sprintf(`, is_final = $%d`, idx)
		args = append(args, *isFinal)
		idx++
	}
	if isActive != nil {
		q += fmt.Sprintf(`, is_active = $%d`, idx)
		args = append(args, *isActive)
		idx++
	}
	q += fmt.Sprintf(` WHERE id = $%d`, idx)
	args = append(args, id)
	_, err := r.db.Pool.Exec(ctx, q, args...)
	return err
}

// ReorderStages принимает [{id, order}, …] и применяет всё одной транзакцией.
func (r *SettingsRepository) ReorderStages(ctx context.Context, ordering []model.PipelineStage) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, s := range ordering {
		if _, err := tx.Exec(ctx, `UPDATE pipeline_stages SET "order" = $1, updated_at = NOW() WHERE id = $2`, s.Order, s.ID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *SettingsRepository) DeleteStage(ctx context.Context, id int) error {
	var refs int
	_ = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM leads WHERE status_id = $1`, id).Scan(&refs)
	if refs > 0 {
		_, err := r.db.Pool.Exec(ctx, `UPDATE pipeline_stages SET is_active = FALSE, updated_at = NOW() WHERE id = $1`, id)
		if err != nil {
			return err
		}
		return fmt.Errorf("stage is in use, archived instead (refs=%d)", refs)
	}
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM pipeline_stages WHERE id = $1`, id)
	return err
}
