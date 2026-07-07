package repository

import (
	"context"
	"fmt"
	"time"

	"crm_backend/internal/model"
	"crm_backend/pkg/database"
)

type LeadRepository struct {
	db *database.DB
}

func NewLeadRepository(db *database.DB) *LeadRepository {
	return &LeadRepository{db: db}
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func (r *LeadRepository) Create(ctx context.Context, lead *model.CreateLeadRequest, sourceID int, statusID int) (*model.Lead, error) {
	query := `
		INSERT INTO leads (
			first_name, last_name, email, phone, program_id,
			utm_source, utm_medium, utm_campaign, source_id, status_id,
			assignee_id, social_url, english_level,
			payment_status, reminder_at, reminder_note, work_company, work_position,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13,
			$14, $15, $16, $17, $18,
			$19, $20
		) RETURNING id, created_at, updated_at
	`

	now := time.Now()
	if lead.CreatedAt != nil {
		now = *lead.CreatedAt
	}
	var newLead model.Lead

	err := r.db.Pool.QueryRow(ctx, query,
		lead.FirstName, lead.LastName, lead.Email, lead.Phone, lead.ProgramID,
		lead.UTMSource, lead.UTMMedium, lead.UTMCampaign, sourceID, statusID,
		lead.AssigneeID, nullStr(lead.SocialURL), nullStr(lead.EnglishLevel),
		lead.PaymentStatus, lead.ReminderAt, nullStr(lead.ReminderNote), nullStr(lead.WorkCompany), nullStr(lead.WorkPosition),
		now, now,
	).Scan(&newLead.ID, &newLead.CreatedAt, &newLead.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create lead: %w", err)
	}

	newLead.FirstName = lead.FirstName
	newLead.LastName = lead.LastName
	newLead.Email = lead.Email
	newLead.Phone = lead.Phone
	newLead.ProgramID = lead.ProgramID
	newLead.UTMSource = lead.UTMSource
	newLead.UTMMedium = lead.UTMMedium
	newLead.UTMCampaign = lead.UTMCampaign
	newLead.SourceID = &sourceID
	newLead.StatusID = &statusID
	newLead.AssigneeID = lead.AssigneeID
	newLead.SocialURL = lead.SocialURL
	newLead.EnglishLevel = lead.EnglishLevel
	newLead.PaymentStatus = lead.PaymentStatus
	newLead.ReminderAt = lead.ReminderAt
	newLead.ReminderNote = lead.ReminderNote
	newLead.WorkCompany = lead.WorkCompany
	newLead.WorkPosition = lead.WorkPosition

	return &newLead, nil
}

func (r *LeadRepository) GetLastAssignedManagerID(ctx context.Context) (int, error) {
	var idStr string
	err := r.db.Pool.QueryRow(ctx, "SELECT value FROM app_state WHERE key = 'last_assigned_user_id'").Scan(&idStr)
	if err != nil {
		return 0, nil // Default to 0 if not found
	}
	var id int
	fmt.Sscanf(idStr, "%d", &id)
	return id, nil
}

func (r *LeadRepository) UpdateLastAssignedManagerID(ctx context.Context, id int) error {
	_, err := r.db.Pool.Exec(ctx, "UPDATE app_state SET value = $1 WHERE key = 'last_assigned_user_id'", fmt.Sprintf("%d", id))
	return err
}

func (r *LeadRepository) GetByID(ctx context.Context, id int) (*model.Lead, error) {
	query := `
		SELECT
			l.id,
			COALESCE(l.first_name, ''), COALESCE(l.last_name, ''),
			COALESCE(l.email, ''), COALESCE(l.phone, ''),
			l.source_id, l.program_id, l.status_id, l.assignee_id, l.contact_id,
			COALESCE(l.utm_source, ''), COALESCE(l.utm_medium, ''), COALESCE(l.utm_campaign, ''),
			COALESCE(c.telegram_id, ''), COALESCE(c.whatsapp_id, ''), COALESCE(c.vk_id, ''),
			COALESCE(l.social_url, ''),
			COALESCE(l.english_level, ''),
			COALESCE(p.name, ''),
			COALESCE(l.payment_status, ''),
			l.reminder_at, COALESCE(l.reminder_note, ''), l.reminder_done,
			COALESCE(l.work_company, ''), COALESCE(l.work_position, ''),
			l.enrolled_at,
			l.created_at, l.updated_at
		FROM leads l
		LEFT JOIN contacts c ON l.contact_id = c.id
		LEFT JOIN programs p ON l.program_id = p.id
		WHERE l.id = $1
	`
	var lead model.Lead
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&lead.ID, &lead.FirstName, &lead.LastName, &lead.Email, &lead.Phone,
		&lead.SourceID, &lead.ProgramID, &lead.StatusID, &lead.AssigneeID, &lead.ContactID,
		&lead.UTMSource, &lead.UTMMedium, &lead.UTMCampaign,
		&lead.TelegramID, &lead.WhatsAppID, &lead.VKID,
		&lead.SocialURL,
		&lead.EnglishLevel,
		&lead.ProgramName,
		&lead.PaymentStatus,
		&lead.ReminderAt, &lead.ReminderNote, &lead.ReminderDone,
		&lead.WorkCompany, &lead.WorkPosition,
		&lead.EnrolledAt,
		&lead.CreatedAt, &lead.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get lead by id: %w", err)
	}
	return &lead, nil
}

func (r *LeadRepository) UpdateStatus(ctx context.Context, leadID int, statusID int) error {
	// При переходе в «Зачисление» (id=6) фиксируем момент зачисления один раз —
	// на нём строится архив зачисленных студентов по учебным годам (#7).
	query := `
		UPDATE leads
		SET status_id = $1,
		    updated_at = $2,
		    enrolled_at = CASE WHEN $1 = 6 THEN COALESCE(enrolled_at, $2) ELSE enrolled_at END
		WHERE id = $3`
	_, err := r.db.Pool.Exec(ctx, query, statusID, time.Now(), leadID)
	if err != nil {
		return fmt.Errorf("failed to update lead status: %w", err)
	}
	return nil
}

// ── Напоминания (#3) ──────────────────────────────────────────────────────

// DueReminder — компактная строка для рассылки уведомлений о напоминаниях.
type DueReminder struct {
	LeadID     int
	Name       string
	AssigneeID *int
	Email      string
	ReminderAt time.Time
	Note       string
}

// DueReminders возвращает лиды, у которых наступил срок напоминания и о нём
// ещё не уведомляли (reminder_notified = FALSE) и оно не закрыто.
func (r *LeadRepository) DueReminders(ctx context.Context) ([]DueReminder, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, TRIM(COALESCE(first_name,'') || ' ' || COALESCE(last_name,'')),
		       assignee_id, COALESCE(email,''), reminder_at, COALESCE(reminder_note,'')
		FROM leads
		WHERE reminder_at IS NOT NULL
		  AND reminder_done = FALSE
		  AND reminder_notified = FALSE
		  AND reminder_at <= NOW()
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DueReminder
	for rows.Next() {
		var d DueReminder
		if err := rows.Scan(&d.LeadID, &d.Name, &d.AssigneeID, &d.Email, &d.ReminderAt, &d.Note); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}

func (r *LeadRepository) MarkReminderNotified(ctx context.Context, leadID int) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE leads SET reminder_notified = TRUE WHERE id = $1`, leadID)
	return err
}

// CompleteReminder закрывает напоминание (кнопка «Выполнено» у менеджера).
func (r *LeadRepository) CompleteReminder(ctx context.Context, leadID int) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE leads SET reminder_done = TRUE WHERE id = $1`, leadID)
	return err
}

func (r *LeadRepository) Update(ctx context.Context, leadID int, req *model.UpdateLeadRequest) error {
	query := `
		UPDATE leads
		SET first_name = $1, last_name = $2, email = $3, phone = $4,
		    program_id = COALESCE($5, program_id),
		    assignee_id = COALESCE($6, assignee_id),
		    utm_source = COALESCE($7, utm_source),
		    social_url = COALESCE($8, social_url),
		    english_level = COALESCE($9, english_level),
		    payment_status = COALESCE($10, payment_status),
		    work_company = COALESCE($11, work_company),
		    work_position = COALESCE($12, work_position),
		    updated_at = $13
		WHERE id = $14
	`
	_, err := r.db.Pool.Exec(ctx, query,
		req.FirstName, req.LastName, req.Email, req.Phone,
		req.ProgramID, req.AssigneeID, req.UTMSource, req.SocialURL, req.EnglishLevel,
		req.PaymentStatus, req.WorkCompany, req.WorkPosition,
		time.Now(), leadID,
	)
	if err != nil {
		return fmt.Errorf("failed to update lead: %w", err)
	}

	// Напоминание (#3): либо снимаем, либо (пере)устанавливаем — тогда сбрасываем
	// флаги доставки/выполнения, чтобы новое напоминание сработало заново.
	if req.ClearReminder {
		_, _ = r.db.Pool.Exec(ctx, `UPDATE leads SET reminder_at = NULL, reminder_note = NULL, reminder_done = FALSE, reminder_notified = FALSE WHERE id = $1`, leadID)
	} else if req.ReminderAt != nil {
		var note interface{}
		if req.ReminderNote != nil {
			note = nullStr(*req.ReminderNote)
		}
		_, _ = r.db.Pool.Exec(ctx, `UPDATE leads SET reminder_at = $1, reminder_note = $2, reminder_done = FALSE, reminder_notified = FALSE WHERE id = $3`, *req.ReminderAt, note, leadID)
	}
	if req.ContactID != nil && req.Email != "" {
		_, _ = r.db.Pool.Exec(ctx, `UPDATE contacts SET first_name = $1, last_name = $2, email = $3, phone = $4, updated_at = $5 WHERE id = $6`,
			req.FirstName, req.LastName, req.Email, req.Phone, time.Now(), *req.ContactID)
	}
	return nil
}

func (r *LeadRepository) Delete(ctx context.Context, leadID int) error {
	_, _ = r.db.Pool.Exec(ctx, `DELETE FROM status_history WHERE lead_id = $1`, leadID)
	_, _ = r.db.Pool.Exec(ctx, `DELETE FROM interactions WHERE lead_id = $1`, leadID)
	_, _ = r.db.Pool.Exec(ctx, `DELETE FROM telegram_chats WHERE lead_id = $1`, leadID)
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM leads WHERE id = $1`, leadID)
	if err != nil {
		return fmt.Errorf("failed to delete lead: %w", err)
	}
	return nil
}

// List returns paginated leads with total count
func (r *LeadRepository) List(ctx context.Context, limit, offset int) ([]model.Lead, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	// Get total count
	var total int
	err := r.db.Pool.QueryRow(ctx, "SELECT count(*) FROM leads").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count leads: %w", err)
	}

	query := `
		SELECT
			l.id,
			COALESCE(l.first_name, ''), COALESCE(l.last_name, ''),
			COALESCE(l.email, ''), COALESCE(l.phone, ''),
			l.source_id, l.program_id, l.status_id, l.assignee_id, l.contact_id,
			COALESCE(l.utm_source, ''), COALESCE(l.utm_medium, ''), COALESCE(l.utm_campaign, ''),
			COALESCE(c.telegram_id, ''), COALESCE(c.whatsapp_id, ''), COALESCE(c.vk_id, ''),
			COALESCE(l.social_url, ''),
			COALESCE(l.english_level, ''),
			COALESCE(p.name, ''),
			COALESCE(l.payment_status, ''),
			l.reminder_at, COALESCE(l.reminder_note, ''), l.reminder_done,
			COALESCE(l.work_company, ''), COALESCE(l.work_position, ''),
			l.enrolled_at,
			l.created_at, l.updated_at
		FROM leads l
		LEFT JOIN contacts c ON l.contact_id = c.id
		LEFT JOIN programs p ON l.program_id = p.id
		ORDER BY l.created_at DESC LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list leads: %w", err)
	}
	defer rows.Close()

	var leads []model.Lead
	for rows.Next() {
		var lead model.Lead
		err := rows.Scan(
			&lead.ID, &lead.FirstName, &lead.LastName, &lead.Email, &lead.Phone,
			&lead.SourceID, &lead.ProgramID, &lead.StatusID, &lead.AssigneeID, &lead.ContactID,
			&lead.UTMSource, &lead.UTMMedium, &lead.UTMCampaign,
			&lead.TelegramID, &lead.WhatsAppID, &lead.VKID,
			&lead.SocialURL,
			&lead.EnglishLevel,
			&lead.ProgramName,
			&lead.PaymentStatus,
			&lead.ReminderAt, &lead.ReminderNote, &lead.ReminderDone,
			&lead.WorkCompany, &lead.WorkPosition,
			&lead.EnrolledAt,
			&lead.CreatedAt, &lead.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("row scan error: %w", err)
		}
		leads = append(leads, lead)
	}
	return leads, total, nil
}

func (r *LeadRepository) LinkToContact(ctx context.Context, leadID int, contactID int) error {
	query := `UPDATE leads SET contact_id = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Pool.Exec(ctx, query, contactID, time.Now(), leadID)
	if err != nil {
		return fmt.Errorf("failed to link contact to lead: %w", err)
	}
	return nil
}

func (r *LeadRepository) MergeLeads(ctx context.Context, targetLeadID, sourceLeadID int) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `UPDATE interactions SET lead_id = $1 WHERE lead_id = $2`, targetLeadID, sourceLeadID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `DELETE FROM leads WHERE id = $1`, sourceLeadID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// FindByContactID finds an existing lead linked to a contact.
// Used by integrations to avoid creating duplicate leads for the same person.
func (r *LeadRepository) FindByContactID(ctx context.Context, contactID int) (*model.Lead, error) {
	query := `
		SELECT 
			l.id, l.first_name, l.last_name, l.email, l.phone,
			l.source_id, l.program_id, l.status_id, l.assignee_id, l.contact_id,
			l.utm_source, l.utm_medium, l.utm_campaign,
			COALESCE(c.telegram_id, ''), COALESCE(c.whatsapp_id, ''), COALESCE(c.vk_id, ''),
			l.created_at, l.updated_at
		FROM leads l
		LEFT JOIN contacts c ON l.contact_id = c.id
		WHERE l.contact_id = $1
		ORDER BY l.created_at DESC
		LIMIT 1
	`
	var lead model.Lead
	err := r.db.Pool.QueryRow(ctx, query, contactID).Scan(
		&lead.ID, &lead.FirstName, &lead.LastName, &lead.Email, &lead.Phone,
		&lead.SourceID, &lead.ProgramID, &lead.StatusID, &lead.AssigneeID, &lead.ContactID,
		&lead.UTMSource, &lead.UTMMedium, &lead.UTMCampaign,
		&lead.TelegramID, &lead.WhatsAppID, &lead.VKID,
		&lead.CreatedAt, &lead.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &lead, nil
}
