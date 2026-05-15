package service_test

import (
	"context"
	"testing"

	"crm_backend/internal/model"
	"crm_backend/internal/service"
)

type mockAnalyticsRepo struct {
	summary    *model.DashboardSummary
	kpis       []model.ManagerKPI
	conversion *model.ConversionReport
}

func (m *mockAnalyticsRepo) GetDashboardSummary(ctx context.Context) (*model.DashboardSummary, error) {
	return m.summary, nil
}
func (m *mockAnalyticsRepo) GetManagerKPIs(ctx context.Context) ([]model.ManagerKPI, error) {
	return m.kpis, nil
}
func (m *mockAnalyticsRepo) GetConversion(ctx context.Context) (*model.ConversionReport, error) {
	return m.conversion, nil
}
func (m *mockAnalyticsRepo) GetSourceBreakdown(ctx context.Context) ([]model.SourceBreakdown, error) {
	return nil, nil
}
func (m *mockAnalyticsRepo) GetFunnel(ctx context.Context) ([]model.StageFunnelPoint, error) {
	return nil, nil
}
func (m *mockAnalyticsRepo) GetTimeSeries(ctx context.Context, days int) ([]model.LeadsTimeSeriesPoint, error) {
	return nil, nil
}

func TestAnalyticsService(t *testing.T) {
	repo := &mockAnalyticsRepo{
		summary: &model.DashboardSummary{TotalNewLeads: 10, TotalDeals: 5, TotalInSchool: 2},
		kpis: []model.ManagerKPI{
			{ManagerID: 1, ManagerName: "Alice", CallsCount: 15, ClosedDeals: 3},
		},
		conversion: &model.ConversionReport{TotalRequests: 100, TotalContracts: 10, ConversionRate: 10.0},
	}

	svc := service.NewAnalyticsService(repo)

	dash, err := svc.GetDashboard(context.Background())
	if err != nil {
		t.Fatalf("expected no err, got %v", err)
	}
	if dash.TotalNewLeads != 10 {
		t.Errorf("expected 10 leads, got %d", dash.TotalNewLeads)
	}

	kpis, err := svc.GetKPIs(context.Background())
	if err != nil {
		t.Fatalf("expected no err, got %v", err)
	}
	if len(kpis) != 1 || kpis[0].ManagerName != "Alice" {
		t.Errorf("expected Alice kpi, got %v", kpis)
	}

	conv, err := svc.GetConversion(context.Background())
	if err != nil {
		t.Fatalf("expected no err, got %v", err)
	}
	if conv.ConversionRate != 10.0 {
		t.Errorf("expected 10.0 conversion, got %v", conv.ConversionRate)
	}
}
