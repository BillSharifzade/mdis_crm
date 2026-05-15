package service

import (
	"context"

	"crm_backend/internal/model"
)

type StatusHistoryService struct {
	repo StatusHistoryRepository
}

func NewStatusHistoryService(repo StatusHistoryRepository) *StatusHistoryService {
	return &StatusHistoryService{repo: repo}
}

func (s *StatusHistoryService) GetLeadStatusHistory(ctx context.Context, leadID int) ([]model.StatusHistory, error) {
	return s.repo.GetByLeadID(ctx, leadID)
}

func (s *StatusHistoryService) RecordStatusChange(ctx context.Context, leadID int, oldStatusID *int, newStatusID int, changedBy *int, refusalReason string) error {
	return s.repo.Create(ctx, leadID, oldStatusID, newStatusID, changedBy, refusalReason)
}
