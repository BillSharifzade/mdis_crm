package service

import (
	"context"

	"crm_backend/internal/model"
	"crm_backend/pkg/events"
)

type InteractionService struct {
	repo InteractionRepository
	bus  *events.Bus
}

func NewInteractionService(repo InteractionRepository, bus *events.Bus) *InteractionService {
	return &InteractionService{repo: repo, bus: bus}
}

func (s *InteractionService) AddInteraction(ctx context.Context, req *model.CreateInteractionRequest) (*model.Interaction, error) {
	it, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	if s.bus != nil && it != nil {
		s.bus.Publish(events.Event{Type: "interaction.created", Payload: it})
	}
	return it, nil
}

func (s *InteractionService) GetLeadHistory(ctx context.Context, leadID int) ([]model.Interaction, error) {
	return s.repo.ListByLead(ctx, leadID)
}
