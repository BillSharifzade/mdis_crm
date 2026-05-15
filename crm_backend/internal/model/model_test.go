package model_test

import (
	"crm_backend/internal/model"
	"testing"
)

func TestDashboardSummary(t *testing.T) {
	summary := model.DashboardSummary{
		TotalNewLeads:         10,
		TotalDeals:            5,
		AverageProcessingSecs: 120.5,
	}

	if summary.TotalNewLeads != 10 {
		t.Errorf("Expected 10 new leads, got %d", summary.TotalNewLeads)
	}
	if summary.AverageProcessingSecs != 120.5 {
		t.Errorf("Expected 120.5 speed, got %v", summary.AverageProcessingSecs)
	}
}

func TestInteractionTask(t *testing.T) {
	i := model.Interaction{
		IsTask: true,
		IsDone: false,
	}
	if !i.IsTask {
		t.Error("Expected IsTask to be true")
	}
	if i.IsDone {
		t.Error("Expected IsDone to be false")
	}
}
