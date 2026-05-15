package repository

import (
	"context"
	"fmt"
	"time"

	"crm_backend/internal/model"
	"crm_backend/pkg/database"
)

type ContactRepository struct {
	db *database.DB
}

func NewContactRepository(db *database.DB) *ContactRepository {
	return &ContactRepository{db: db}
}

func (r *ContactRepository) FindOrCreate(ctx context.Context, firstName, lastName, email, phone, telegramID, whatsappID, vkID string) (*model.Contact, error) {
	findQuery := `
		SELECT id, first_name, last_name, email, phone, telegram_id, whatsapp_id, vk_id, created_at, updated_at 
		FROM contacts 
		WHERE (email = $1 AND email != '') 
		   OR (phone = $2 AND phone != '')
		   OR (telegram_id = $3 AND telegram_id != '')
		   OR (whatsapp_id = $4 AND whatsapp_id != '')
		   OR (vk_id = $5 AND vk_id != '')
		LIMIT 1
	`
	var contact model.Contact
	err := r.db.Pool.QueryRow(ctx, findQuery, email, phone, telegramID, whatsappID, vkID).Scan(
		&contact.ID, &contact.FirstName, &contact.LastName, &contact.Email, &contact.Phone,
		&contact.TelegramID, &contact.WhatsAppID, &contact.VKID,
		&contact.CreatedAt, &contact.UpdatedAt,
	)

	if err == nil {
		return &contact, nil
	}

	createQuery := `
		INSERT INTO contacts (first_name, last_name, email, phone, telegram_id, whatsapp_id, vk_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	now := time.Now()
	
	// Convert empty strings to nil for cleaner DB storage if needed, 
	// but here we just pass them and let scan handle pointers.
	
	err = r.db.Pool.QueryRow(ctx, createQuery,
		firstName, lastName, 
		sqlNullable(email), sqlNullable(phone), 
		sqlNullable(telegramID), sqlNullable(whatsappID), sqlNullable(vkID), 
		now, now,
	).Scan(&contact.ID, &contact.CreatedAt, &contact.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert contact: %w", err)
	}

	contact.FirstName = firstName
	contact.LastName = lastName
	contact.Email = ptr(email)
	contact.Phone = ptr(phone)
	contact.TelegramID = ptr(telegramID)
	contact.WhatsAppID = ptr(whatsappID)
	contact.VKID = ptr(vkID)

	return &contact, nil
}

func sqlNullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ptr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
