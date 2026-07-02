package service_test

import (
	"context"
	"crm_backend/internal/model"
	"crm_backend/internal/service"
	"testing"
)

type mockInteractionService struct {
	added bool
}

func (m *mockInteractionService) AddInteraction(ctx context.Context, req *model.CreateInteractionRequest) (*model.Interaction, error) {
	m.added = true
	return &model.Interaction{ID: 1}, nil
}
func (m *mockInteractionService) GetLeadHistory(ctx context.Context, leadID int) ([]model.Interaction, error) {
	return nil, nil
}

type mockTGChatRepo struct {
	chat *model.TelegramChat
}

func (m *mockTGChatRepo) GetByChatID(ctx context.Context, chatID int64) (*model.TelegramChat, error) {
	return m.chat, nil
}
func (m *mockTGChatRepo) GetByLeadID(ctx context.Context, leadID int) (*model.TelegramChat, error) {
	return m.chat, nil
}
func (m *mockTGChatRepo) Create(ctx context.Context, chatID int64, username, firstName string) (*model.TelegramChat, error) {
	c := &model.TelegramChat{ID: 1, ChatID: chatID, BotState: model.BotStateGreet, BotActive: true}
	m.chat = c
	return c, nil
}
func (m *mockTGChatRepo) UpdateState(ctx context.Context, id int, state string) error {
	if m.chat != nil {
		m.chat.BotState = state
	}
	return nil
}
func (m *mockTGChatRepo) SetBotActive(ctx context.Context, id int, active bool) error {
	if m.chat != nil {
		m.chat.BotActive = active
	}
	return nil
}
func (m *mockTGChatRepo) SetLeadID(ctx context.Context, id, leadID int) error {
	if m.chat != nil {
		m.chat.LeadID = &leadID
	}
	return nil
}
func (m *mockTGChatRepo) SetCollected(ctx context.Context, id int, field, value string) error {
	if m.chat == nil {
		return nil
	}
	switch field {
	case "collected_name":
		m.chat.CollectedName = value
	case "collected_program":
		m.chat.CollectedProgram = value
	case "collected_phone":
		m.chat.CollectedPhone = value
	}
	return nil
}

func (m *mockTGChatRepo) GetPollOffset(ctx context.Context) (int, error) { return 0, nil }
func (m *mockTGChatRepo) SetPollOffset(ctx context.Context, offset int) error { return nil }

func TestIntegrationService_ProcessTelegramWebhook(t *testing.T) {
	contactRepo := &mockContactRepo{contact: &model.Contact{ID: 1}}
	leadRepo := &mockLeadRepo{createdLead: &model.Lead{ID: 1}}
	leadSvc := service.NewLeadService(leadRepo, contactRepo, &mockDealRepo{deal: &model.Deal{ID: 10}}, &mockUserRepo{managers: []model.User{{ID: 1}}}, &mockNotificationService{}, &mockStatusHistoryRepo{}, nil)
	intSvc := &mockInteractionService{}
	tgRepo := &mockTGChatRepo{}

	// token "" → bot scenario работает без отправки в Telegram (только state machine)
	bot := service.NewTelegramBotService("", tgRepo, leadSvc, intSvc, contactRepo, leadRepo, nil, nil, nil, nil)
	svc := service.NewIntegrationService(leadSvc, intSvc, contactRepo, leadRepo, bot)

	payload := &model.TelegramWebhookRequest{}
	payload.Message.Text = "/start"
	payload.Message.From.ID = 123
	payload.Message.From.FirstName = "TestUser"
	payload.Message.From.Username = "test_tg"
	payload.Message.Chat.ID = 555

	if err := svc.ProcessTelegramWebhook(context.Background(), payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tgRepo.chat == nil {
		t.Fatal("telegram chat should be created")
	}
	if tgRepo.chat.BotState != model.BotStateAskName {
		t.Errorf("after /start state should be ask_name, got %q", tgRepo.chat.BotState)
	}
}
