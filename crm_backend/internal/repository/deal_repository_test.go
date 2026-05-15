package repository_test

import (
	"context"
	"crm_backend/internal/repository"
	"crm_backend/pkg/database"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
)

func TestDealRepository_Create(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	db := &database.DB{Pool: mock}
	repo := repository.NewDealRepository(db)

	mock.ExpectQuery("INSERT INTO deals").
		WithArgs(1, 1, 1, pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(1, time.Now(), time.Now()))

	deal, err := repo.Create(context.Background(), 1, 1, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if deal.ID != 1 {
		t.Errorf("Expected ID 1, got %d", deal.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectations were not met: %s", err)
	}
}
