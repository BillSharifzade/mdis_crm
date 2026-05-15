package repository_test

import (
	"context"
	"crm_backend/internal/model"
	"crm_backend/internal/repository"
	"crm_backend/pkg/database"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
)

func TestInteractionRepository_Create(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	db := &database.DB{Pool: mock}
	repo := repository.NewInteractionRepository(db)

	leadID := 1
	req := &model.CreateInteractionRequest{
		LeadID:  &leadID,
		Type:    "call",
		Content: "test interaction",
		IsTask:  true,
		DueDate: nil,
	}

	mock.ExpectQuery("INSERT INTO interactions").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), "call", "inbound", "test interaction",
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
			true, pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at"}).AddRow(1, time.Now()))

	res, err := repo.Create(context.Background(), req)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if res.ID != 1 {
		t.Errorf("Expected ID 1, got %d", res.ID)
	}
	if !res.IsTask {
		t.Error("Expected IsTask to be true")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectations were not met: %s", err)
	}
}
