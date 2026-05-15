package service_test

import (
	"bytes"
	"context"
	"crm_backend/internal/model"
	"crm_backend/internal/service"
	"testing"
)

func TestExportService_GenerateLeadsExcel(t *testing.T) {
	leadRepo := &mockLeadRepo{
		leads: []model.Lead{
			{ID: 1, FirstName: "John"},
		},
	}
	svc := service.NewExportService(leadRepo)

	var buf bytes.Buffer
	err := svc.GenerateLeadsExcel(context.Background(), &buf)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Excel buffer is empty")
	}
}
