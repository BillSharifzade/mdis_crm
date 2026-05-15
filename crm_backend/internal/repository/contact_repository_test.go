package repository_test

import (
	"context"
	"crm_backend/internal/repository"
	"crm_backend/pkg/database"
	"fmt"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
)

func strPtr(s string) *string { return &s }

func TestContactRepository_FindOrCreate(t *testing.T) {
	t.Run("Find Existing", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatal(err)
		}
		defer mock.Close()

		db := &database.DB{Pool: mock}
		repo := repository.NewContactRepository(db)

		email := "john@test.com"
		phone := "123"
		tgID := "tg1"

		rows := pgxmock.NewRows([]string{"id", "first_name", "last_name", "email", "phone", "telegram_id", "whatsapp_id", "vk_id", "created_at", "updated_at"}).
			AddRow(1, "John", "Doe", &email, &phone, &tgID, (*string)(nil), (*string)(nil), time.Now(), time.Now())

		mock.ExpectQuery("SELECT").
			WithArgs("john@test.com", "123", "tg1", "", "").
			WillReturnRows(rows)

		contact, err := repo.FindOrCreate(context.Background(), "John", "Doe", "john@test.com", "123", "tg1", "", "")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if contact.ID != 1 {
			t.Errorf("Expected ID 1, got %d", contact.ID)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Expectations were not met: %s", err)
		}
	})

	t.Run("Create New", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		if err != nil {
			t.Fatal(err)
		}
		defer mock.Close()

		db := &database.DB{Pool: mock}
		repo := repository.NewContactRepository(db)

		mock.ExpectQuery("SELECT").
			WithArgs("new@test.com", "", "", "", "").
			WillReturnError(fmt.Errorf("no rows"))

		mock.ExpectQuery("INSERT INTO contacts").
			WithArgs("New", "User", pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).
				AddRow(2, time.Now(), time.Now()))

		contact, err := repo.FindOrCreate(context.Background(), "New", "User", "new@test.com", "", "", "", "")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if contact.ID != 2 {
			t.Errorf("Expected ID 2, got %d", contact.ID)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Expectations were not met: %s", err)
		}
	})
}
