package repository_test

import (
	"context"
	"crm_backend/internal/repository"
	"crm_backend/pkg/database"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
)

func TestUserRepository_GetByEmail(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	db := &database.DB{Pool: mock}
	repo := repository.NewUserRepository(db)

	rows := pgxmock.NewRows([]string{"id", "name", "email", "password_hash", "role", "created_at", "updated_at"}).
		AddRow(1, "Admin", "admin@test.com", "hash", "admin", time.Now(), time.Now())

	mock.ExpectQuery("SELECT").WithArgs("admin@test.com").WillReturnRows(rows)

	user, err := repo.GetByEmail(context.Background(), "admin@test.com")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if user.Name != "Admin" {
		t.Errorf("Expected Admin, got %s", user.Name)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectations were not met: %s", err)
	}
}
