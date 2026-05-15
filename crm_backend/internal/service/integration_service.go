package service

import (
	"context"
	"fmt"

	"crm_backend/internal/model"
)

type IntegrationService struct {
	leadSvc     ILeadService
	intSvc      IInteractionService
	contactRepo ContactRepository
	leadRepo    LeadRepository
	botSvc      *TelegramBotService
}

func NewIntegrationService(leadSvc ILeadService, intSvc IInteractionService, contactRepo ContactRepository, leadRepo LeadRepository, botSvc *TelegramBotService) *IntegrationService {
	return &IntegrationService{
		leadSvc:     leadSvc,
		intSvc:      intSvc,
		contactRepo: contactRepo,
		leadRepo:    leadRepo,
		botSvc:      botSvc,
	}
}

// ProcessTelegramWebhook делегирует обработку входящего сообщения сервису бота
// (state machine + outbound replies). Сам по себе вебхук теперь — тонкий враппер.
func (s *IntegrationService) ProcessTelegramWebhook(ctx context.Context, payload *model.TelegramWebhookRequest) error {
	if s.botSvc == nil {
		return fmt.Errorf("telegram bot service not configured")
	}
	return s.botSvc.HandleIncoming(ctx, payload)
}

func (s *IntegrationService) ProcessTelephonyWebhook(ctx context.Context, payload *model.TelephonyWebhookRequest) error {
	if payload.Caller == "" {
		return nil
	}

	// Try to find existing contact by phone
	contact, err := s.contactRepo.FindOrCreate(ctx, "Unknown Caller", "", "", payload.Caller, "", "", "")
	if err != nil {
		return fmt.Errorf("failed to sync telephony contact: %w", err)
	}

	// Check if lead already exists for this contact
	var leadID int
	existingLead, err := s.leadRepo.FindByContactID(ctx, contact.ID)
	if err == nil && existingLead != nil {
		leadID = existingLead.ID
	} else {
		// Create new lead
		req := &model.CreateLeadRequest{
			FirstName: "Unknown Caller",
			Phone:     payload.Caller,
		}
		lead, err := s.leadSvc.CreateLeadFromForm(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to sync telephony lead: %w", err)
		}
		leadID = lead.ID
	}

	isTask := false
	if payload.Status == "missed" || payload.Status == "no-answer" || payload.Status == "busy" {
		isTask = true
	}

	interactionReq := &model.CreateInteractionRequest{
		LeadID:         &leadID,
		Type:           "call",
		Content:        fmt.Sprintf("Call status: %s, Recording: %s", payload.Status, payload.Recording),
		DurationSecond: &payload.Duration,
		IsTask:         isTask,
	}

	_, err = s.intSvc.AddInteraction(ctx, interactionReq)
	if err != nil {
		return fmt.Errorf("failed to save call interaction history: %w", err)
	}

	return nil
}
