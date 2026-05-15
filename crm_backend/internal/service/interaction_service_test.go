package service_test

import (
	"context"
	"testing"

	"crm_backend/internal/model"
	"crm_backend/internal/service"
)

type mockInteractionRepo struct {
	createdInt *model.Interaction
	history    []model.Interaction
}

func (m *mockInteractionRepo) Create(ctx context.Context, req *model.CreateInteractionRequest) (*model.Interaction, error) {
	return m.createdInt, nil
}
func (m *mockInteractionRepo) ListByLead(ctx context.Context, leadID int) ([]model.Interaction, error) {
	return m.history, nil
}

func TestInteractionService(t *testing.T) {
	expectedInt := &model.Interaction{ID: 1, LeadID: func() *int { i := 1; return &i }(), Type: "call"}
	repo := &mockInteractionRepo{
		createdInt: expectedInt,
		history:    []model.Interaction{*expectedInt},
	}
	svc := service.NewInteractionService(repo, nil)

	req := &model.CreateInteractionRequest{Type: "call", Content: "Hello"}
	res, err := svc.AddInteraction(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no err, got %v", err)
	}

	if res.ID != 1 {
		t.Errorf("expected int ID 1, got %d", res.ID)
	}

	history, err := svc.GetLeadHistory(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no err, got %v", err)
	}

	if len(history) != 1 {
		t.Errorf("expected 1 item, got %d", len(history))
	}
}
