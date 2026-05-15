package repository

import (
	"context"
	"fmt"

	"crm_backend/internal/model"
	"crm_backend/pkg/database"
)

type ProgramRepository struct {
	db *database.DB
}

func NewProgramRepository(db *database.DB) *ProgramRepository {
	return &ProgramRepository{db: db}
}

func (r *ProgramRepository) ListAll(ctx context.Context) ([]model.Program, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, name, COALESCE(faculty, ''), is_active
		FROM programs
		WHERE is_active = TRUE
		ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list programs: %w", err)
	}
	defer rows.Close()

	var out []model.Program
	for rows.Next() {
		var p model.Program
		if err := rows.Scan(&p.ID, &p.Name, &p.Faculty, &p.IsActive); err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *ProgramRepository) FindIDByName(ctx context.Context, name string) (int, bool, error) {
	var id int
	err := r.db.Pool.QueryRow(ctx, `SELECT id FROM programs WHERE name = $1 LIMIT 1`, name).Scan(&id)
	if err != nil {
		// pgx returns ErrNoRows when no match — treat as "not found"
		if err.Error() == "no rows in result set" {
			return 0, false, nil
		}
		return 0, false, err
	}
	return id, true, nil
}
