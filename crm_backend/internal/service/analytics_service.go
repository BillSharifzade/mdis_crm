package service

import (
	"context"

	"crm_backend/internal/model"
)

type AnalyticsService struct {
	repo AnalyticsRepository
}

func NewAnalyticsService(repo AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

func (s *AnalyticsService) GetDashboard(ctx context.Context) (*model.DashboardSummary, error) {
	return s.repo.GetDashboardSummary(ctx)
}

func (s *AnalyticsService) GetKPIs(ctx context.Context) ([]model.ManagerKPI, error) {
	return s.repo.GetManagerKPIs(ctx)
}

func (s *AnalyticsService) GetConversion(ctx context.Context) (*model.ConversionReport, error) {
	return s.repo.GetConversion(ctx)
}

func (s *AnalyticsService) GetSourceBreakdown(ctx context.Context) ([]model.SourceBreakdown, error) {
	return s.repo.GetSourceBreakdown(ctx)
}

func (s *AnalyticsService) GetFunnel(ctx context.Context) ([]model.StageFunnelPoint, error) {
	return s.repo.GetFunnel(ctx)
}

func (s *AnalyticsService) GetTimeSeries(ctx context.Context, days int) ([]model.LeadsTimeSeriesPoint, error) {
	return s.repo.GetTimeSeries(ctx, days)
}
