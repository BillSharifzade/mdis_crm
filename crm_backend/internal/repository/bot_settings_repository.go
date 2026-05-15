package repository

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"crm_backend/pkg/database"
)

// BotSettingsRepository — key/value хранилище текстов телеграм-бота.
// Кэш в памяти (TTL 30s) — бот читает на каждом сообщении.
type BotSettingsRepository struct {
	db    *database.DB
	mu    sync.RWMutex
	cache map[string]string
	exp   time.Time
}

func NewBotSettingsRepository(db *database.DB) *BotSettingsRepository {
	return &BotSettingsRepository{db: db, cache: map[string]string{}}
}

const botSettingsTTL = 30 * time.Second

func (r *BotSettingsRepository) refresh(ctx context.Context) error {
	rows, err := r.db.Pool.Query(ctx, `SELECT key, value FROM bot_settings`)
	if err != nil {
		return err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return err
		}
		out[k] = v
	}
	r.mu.Lock()
	r.cache = out
	r.exp = time.Now().Add(botSettingsTTL)
	r.mu.Unlock()
	return nil
}

func (r *BotSettingsRepository) Get(ctx context.Context, key, fallback string) string {
	r.mu.RLock()
	if time.Now().Before(r.exp) {
		v, ok := r.cache[key]
		r.mu.RUnlock()
		if ok {
			return v
		}
		return fallback
	}
	r.mu.RUnlock()
	if err := r.refresh(ctx); err != nil {
		return fallback
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if v, ok := r.cache[key]; ok {
		return v
	}
	return fallback
}

func (r *BotSettingsRepository) GetAll(ctx context.Context) (map[string]string, error) {
	if err := r.refresh(ctx); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]string, len(r.cache))
	for k, v := range r.cache {
		out[k] = v
	}
	return out, nil
}

func (r *BotSettingsRepository) Set(ctx context.Context, key, value string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("key is empty")
	}
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO bot_settings (key, value, updated_at) VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
	`, key, value)
	if err != nil {
		return err
	}
	r.mu.Lock()
	r.exp = time.Time{} // инвалидируем кэш
	r.mu.Unlock()
	return nil
}
