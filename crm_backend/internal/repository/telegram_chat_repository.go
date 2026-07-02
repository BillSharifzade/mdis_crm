package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"crm_backend/internal/model"
	"crm_backend/pkg/database"

	"github.com/jackc/pgx/v5"
)

type TelegramChatRepository struct {
	db *database.DB
}

func NewTelegramChatRepository(db *database.DB) *TelegramChatRepository {
	return &TelegramChatRepository{db: db}
}

func (r *TelegramChatRepository) GetByChatID(ctx context.Context, chatID int64) (*model.TelegramChat, error) {
	row := r.db.Pool.QueryRow(ctx, `
		SELECT id, chat_id, lead_id, bot_state, bot_active,
		       COALESCE(tg_username, ''), COALESCE(tg_first_name, ''),
		       COALESCE(collected_name, ''), COALESCE(collected_program, ''), COALESCE(collected_english, ''), COALESCE(collected_phone, ''),
		       created_at, updated_at
		FROM telegram_chats WHERE chat_id = $1
	`, chatID)
	var c model.TelegramChat
	err := row.Scan(&c.ID, &c.ChatID, &c.LeadID, &c.BotState, &c.BotActive,
		&c.TGUsername, &c.TGFirstName,
		&c.CollectedName, &c.CollectedProgram, &c.CollectedEnglish, &c.CollectedPhone,
		&c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("telegram_chats get: %w", err)
	}
	return &c, nil
}

func (r *TelegramChatRepository) GetByLeadID(ctx context.Context, leadID int) (*model.TelegramChat, error) {
	row := r.db.Pool.QueryRow(ctx, `
		SELECT id, chat_id, lead_id, bot_state, bot_active,
		       COALESCE(tg_username, ''), COALESCE(tg_first_name, ''),
		       COALESCE(collected_name, ''), COALESCE(collected_program, ''), COALESCE(collected_english, ''), COALESCE(collected_phone, ''),
		       created_at, updated_at
		FROM telegram_chats WHERE lead_id = $1
		ORDER BY id DESC LIMIT 1
	`, leadID)
	var c model.TelegramChat
	err := row.Scan(&c.ID, &c.ChatID, &c.LeadID, &c.BotState, &c.BotActive,
		&c.TGUsername, &c.TGFirstName,
		&c.CollectedName, &c.CollectedProgram, &c.CollectedEnglish, &c.CollectedPhone,
		&c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("telegram_chats get by lead: %w", err)
	}
	return &c, nil
}

func (r *TelegramChatRepository) Create(ctx context.Context, chatID int64, username, firstName string) (*model.TelegramChat, error) {
	now := time.Now()
	var c model.TelegramChat
	err := r.db.Pool.QueryRow(ctx, `
		INSERT INTO telegram_chats (chat_id, bot_state, bot_active, tg_username, tg_first_name, created_at, updated_at)
		VALUES ($1, $2, TRUE, $3, $4, $5, $5)
		RETURNING id, chat_id, lead_id, bot_state, bot_active,
		          COALESCE(tg_username, ''), COALESCE(tg_first_name, ''),
		          COALESCE(collected_name, ''), COALESCE(collected_program, ''), COALESCE(collected_phone, ''),
		          created_at, updated_at
	`, chatID, model.BotStateGreet, username, firstName, now).Scan(
		&c.ID, &c.ChatID, &c.LeadID, &c.BotState, &c.BotActive,
		&c.TGUsername, &c.TGFirstName,
		&c.CollectedName, &c.CollectedProgram, &c.CollectedEnglish, &c.CollectedPhone,
		&c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("telegram_chats create: %w", err)
	}
	return &c, nil
}

func (r *TelegramChatRepository) UpdateState(ctx context.Context, id int, state string) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE telegram_chats SET bot_state=$1, updated_at=NOW() WHERE id=$2`, state, id)
	return err
}

func (r *TelegramChatRepository) SetBotActive(ctx context.Context, id int, active bool) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE telegram_chats SET bot_active=$1, updated_at=NOW() WHERE id=$2`, active, id)
	return err
}

func (r *TelegramChatRepository) SetLeadID(ctx context.Context, id int, leadID int) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE telegram_chats SET lead_id=$1, updated_at=NOW() WHERE id=$2`, leadID, id)
	return err
}

// UnreadByLead возвращает lead_id → unread_count.
// «Непрочитано» = inbound messenger-сообщения, пришедшие после последнего outbound
// от менеджера (created_by IS NOT NULL). Если outbound вообще не было — считаем
// все inbound непрочитанными.
func (r *TelegramChatRepository) UnreadByLead(ctx context.Context) (map[int]int, error) {
	q := `
	WITH last_mgr_reply AS (
		SELECT lead_id, MAX(created_at) AS ts
		FROM interactions
		WHERE type = 'messenger' AND direction = 'outbound' AND created_by IS NOT NULL AND lead_id IS NOT NULL
		GROUP BY lead_id
	)
	SELECT i.lead_id, COUNT(*)
	FROM interactions i
	LEFT JOIN last_mgr_reply r ON r.lead_id = i.lead_id
	WHERE i.type = 'messenger' AND i.direction = 'inbound' AND i.lead_id IS NOT NULL
	  AND (r.ts IS NULL OR i.created_at > r.ts)
	GROUP BY i.lead_id
	`
	rows, err := r.db.Pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int]int{}
	for rows.Next() {
		var leadID, cnt int
		if err := rows.Scan(&leadID, &cnt); err != nil {
			return nil, err
		}
		out[leadID] = cnt
	}
	return out, nil
}

// GetPollOffset читает сохранённый getUpdates-offset из app_state. Нужен, чтобы
// после РЕСТАРТА бэкенда бот не переигрывал уже обработанные апдейты заново
// (иначе Telegram переотдаёт неподтверждённый бэклог → дублирующиеся ответы,
// повторные «как вас зовут?» и рассинхрон состояния диалога).
func (r *TelegramChatRepository) GetPollOffset(ctx context.Context) (int, error) {
	var val string
	err := r.db.Pool.QueryRow(ctx,
		`SELECT value FROM app_state WHERE key = 'telegram_poll_offset'`).Scan(&val)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	n, convErr := strconv.Atoi(strings.TrimSpace(val))
	if convErr != nil {
		return 0, nil
	}
	return n, nil
}

// SetPollOffset сохраняет обработанный offset. Вызывается после каждого апдейта.
func (r *TelegramChatRepository) SetPollOffset(ctx context.Context, offset int) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO app_state (key, value, updated_at)
		VALUES ('telegram_poll_offset', $1, NOW())
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
	`, strconv.Itoa(offset))
	return err
}

func (r *TelegramChatRepository) SetCollected(ctx context.Context, id int, field, value string) error {
	allowed := map[string]bool{"collected_name": true, "collected_program": true, "collected_english": true, "collected_phone": true}
	if !allowed[field] {
		return fmt.Errorf("unknown field: %s", field)
	}
	q := fmt.Sprintf(`UPDATE telegram_chats SET %s=$1, updated_at=NOW() WHERE id=$2`, field)
	_, err := r.db.Pool.Exec(ctx, q, value, id)
	return err
}
