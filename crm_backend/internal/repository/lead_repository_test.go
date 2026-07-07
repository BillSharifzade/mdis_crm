package repository_test

import (
	"context"
	"crm_backend/internal/repository"
	"crm_backend/pkg/database"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
)

func intPtr(i int) *int { return &i }

func TestLeadRepository_GetByID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	db := &database.DB{Pool: mock}
	repo := repository.NewLeadRepository(db)

	// Lead model: StatusID *int, ContactID *int, but TelegramID/WhatsAppID/VKID are plain string
	statusID := 1
	contactID := 2

	rows := pgxmock.NewRows([]string{"id", "first_name", "last_name", "email", "phone", "source_id", "program_id", "status_id", "assignee_id", "contact_id", "utm_source", "utm_medium", "utm_campaign", "telegram_id", "whatsapp_id", "vk_id", "social_url", "english_level", "program_name", "payment_status", "reminder_at", "reminder_note", "reminder_done", "work_company", "work_position", "enrolled_at", "created_at", "updated_at"}).
		AddRow(1, "John", "Doe", "john@doe.com", "123", (*int)(nil), (*int)(nil), &statusID, (*int)(nil), &contactID, "source", "medium", "campaign", "tg123", "", "", "", "", "", "", (*time.Time)(nil), "", false, "", "", (*time.Time)(nil), time.Now(), time.Now())

	mock.ExpectQuery("SELECT").WithArgs(1).WillReturnRows(rows)

	lead, err := repo.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if lead.FirstName != "John" {
		t.Errorf("Expected John, got %s", lead.FirstName)
	}
	if lead.TelegramID != "tg123" {
		t.Errorf("Expected tg123, got %s", lead.TelegramID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectations were not met: %s", err)
	}
}
