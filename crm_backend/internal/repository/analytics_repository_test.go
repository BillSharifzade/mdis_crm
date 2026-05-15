package repository_test

import (
	"context"
	"crm_backend/internal/repository"
	"crm_backend/pkg/database"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
)

func TestAnalyticsRepository_GetDashboardSummary(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	db := &database.DB{Pool: mock}
	repo := repository.NewAnalyticsRepository(db)

	rows := pgxmock.NewRows([]string{
		"total_new_leads", "total_deals", "total_in_school", "total_lost",
		"total_pending_tasks", "total_leads",
		"overall_conversion", "average_processing_secs", "avg_lead_to_enrolled_days",
	}).AddRow(10, 5, 2, 1, 3, 20, 15.5, 150.5, 4.2)

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	summary, err := repo.GetDashboardSummary(context.Background())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if summary.TotalNewLeads != 10 {
		t.Errorf("Expected 10 new leads, got %d", summary.TotalNewLeads)
	}
	if summary.TotalPendingTasks != 3 {
		t.Errorf("Expected 3 tasks, got %d", summary.TotalPendingTasks)
	}
	if summary.AverageProcessingSecs != 150.5 {
		t.Errorf("Expected 150.5 speed, got %v", summary.AverageProcessingSecs)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectations were not met: %s", err)
	}
}
