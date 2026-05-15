package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"crm_backend/internal/api"
	"crm_backend/internal/model"
	"crm_backend/internal/service"

	"github.com/go-chi/chi/v5"
)

type mockLeadService struct {
	lead *model.Lead
	err  error
}

func (m *mockLeadService) CreateLeadFromForm(ctx context.Context, req *model.CreateLeadRequest) (*model.Lead, error) {
	return m.lead, m.err
}
func (m *mockLeadService) GetLead(ctx context.Context, id int) (*model.Lead, error) {
	return m.lead, m.err
}
func (m *mockLeadService) ListLeads(ctx context.Context, limit, offset int) ([]model.Lead, int, error) {
	return []model.Lead{*m.lead}, 1, m.err
}
func (m *mockLeadService) UpdateLeadStatus(ctx context.Context, leadID int, statusID int, refusalReason string, examDate *time.Time) error {
	return m.err
}
func (m *mockLeadService) MergeLeads(ctx context.Context, targetLeadID, sourceLeadID int) error {
	return m.err
}
func (m *mockLeadService) ImportLeads(ctx context.Context, r io.Reader) error {
	return m.err
}
func (m *mockLeadService) ImportLeadsRich(ctx context.Context, r io.Reader,
	programLookup func(ctx context.Context, name string) (int, bool),
	sourceLookup func(ctx context.Context, name string) (int, bool),
) (*service.ImportResult, error) {
	return &service.ImportResult{}, m.err
}
func (m *mockLeadService) UpdateLead(ctx context.Context, leadID int, req *model.UpdateLeadRequest) (*model.Lead, error) {
	return m.lead, m.err
}
func (m *mockLeadService) DeleteLead(ctx context.Context, leadID int) error {
	return m.err
}

type mockAuthService struct {
	res *model.LoginResponse
	err error
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (*model.LoginResponse, error) {
	return m.res, m.err
}

type mockInteractionService struct {
	interaction *model.Interaction
	err         error
}

func (m *mockInteractionService) AddInteraction(ctx context.Context, req *model.CreateInteractionRequest) (*model.Interaction, error) {
	return m.interaction, m.err
}
func (m *mockInteractionService) GetLeadHistory(ctx context.Context, leadID int) ([]model.Interaction, error) {
	return []model.Interaction{*m.interaction}, m.err
}

type mockAnalyticsService struct {
	summary    *model.DashboardSummary
	kpis       []model.ManagerKPI
	conversion *model.ConversionReport
	err        error
}

func (m *mockAnalyticsService) GetDashboard(ctx context.Context) (*model.DashboardSummary, error) {
	return m.summary, m.err
}
func (m *mockAnalyticsService) GetKPIs(ctx context.Context) ([]model.ManagerKPI, error) {
	return m.kpis, m.err
}
func (m *mockAnalyticsService) GetConversion(ctx context.Context) (*model.ConversionReport, error) {
	return m.conversion, m.err
}
func (m *mockAnalyticsService) GetSourceBreakdown(ctx context.Context) ([]model.SourceBreakdown, error) {
	return nil, m.err
}
func (m *mockAnalyticsService) GetFunnel(ctx context.Context) ([]model.StageFunnelPoint, error) {
	return nil, m.err
}
func (m *mockAnalyticsService) GetTimeSeries(ctx context.Context, days int) ([]model.LeadsTimeSeriesPoint, error) {
	return nil, m.err
}

type mockExportService struct {
	err error
}

func (m *mockExportService) GenerateLeadsExcel(ctx context.Context, w io.Writer) error {
	return m.err
}
func (m *mockExportService) GenerateLeadsPDF(ctx context.Context) ([]byte, error) {
	return []byte("pdf"), m.err
}

type mockIntegrationService struct {
	err error
}

func (m *mockIntegrationService) ProcessTelegramWebhook(ctx context.Context, payload *model.TelegramWebhookRequest) error {
	return m.err
}
func (m *mockIntegrationService) ProcessTelephonyWebhook(ctx context.Context, payload *model.TelephonyWebhookRequest) error {
	return m.err
}

type mockUserService struct {
	users []model.User
	user  *model.User
	err   error
}

func (m *mockUserService) ListUsers(ctx context.Context) ([]model.User, error) {
	return m.users, m.err
}
func (m *mockUserService) CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.User, error) {
	return m.user, m.err
}

type mockStatusHistoryService struct {
	history []model.StatusHistory
	err     error
}

func (m *mockStatusHistoryService) GetLeadStatusHistory(ctx context.Context, leadID int) ([]model.StatusHistory, error) {
	return m.history, m.err
}
func (m *mockStatusHistoryService) RecordStatusChange(ctx context.Context, leadID int, oldStatusID *int, newStatusID int, changedBy *int, refusalReason string) error {
	return m.err
}

func TestAuthHandler_Login(t *testing.T) {
	svc := &mockAuthService{
		res: &model.LoginResponse{Token: "test-token"},
	}
	handler := api.NewAuthHandler(svc)

	reqBody, _ := json.Marshal(model.LoginRequest{Email: "test@example.com", Password: "password"})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var res model.LoginResponse
	json.NewDecoder(rr.Body).Decode(&res)
	if res.Token != "test-token" {
		t.Errorf("expected token 'test-token', got %s", res.Token)
	}
}

func TestLeadHandler_CRUD(t *testing.T) {
	lead := &model.Lead{ID: 1, FirstName: "John"}
	svc := &mockLeadService{lead: lead}
	handler := api.NewLeadHandler(svc)

	t.Run("Create", func(t *testing.T) {
		body, _ := json.Marshal(model.CreateLeadRequest{FirstName: "John", Email: "john@test.com"})
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		handler.Routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rr.Code)
		}
	})

	t.Run("List", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		handler.Routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("Get", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/1", nil)
		rr := httptest.NewRecorder()
		handler.Routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("UpdateStatus", func(t *testing.T) {
		body, _ := json.Marshal(map[string]int{"status_id": 2})
		req := httptest.NewRequest(http.MethodPatch, "/1/status", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		handler.Routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusNoContent {
			t.Errorf("expected 204, got %d", rr.Code)
		}
	})
}

func TestInteractionHandler(t *testing.T) {
	interaction := &model.Interaction{ID: 1}
	svc := &mockInteractionService{interaction: interaction}
	handler := api.NewInteractionHandler(svc)

	t.Run("Add", func(t *testing.T) {
		leadID := 1
		body, _ := json.Marshal(model.CreateInteractionRequest{
			LeadID:  &leadID,
			Type:    "call",
			Content: "Follow up",
		})
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		handler.Routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rr.Code)
		}
	})

	t.Run("GetByLead", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/lead/1", nil)
		rr := httptest.NewRecorder()
		handler.Routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})
}

func TestAnalyticsHandler(t *testing.T) {
	svc := &mockAnalyticsService{
		summary:    &model.DashboardSummary{TotalDeals: 10},
		conversion: &model.ConversionReport{TotalRequests: 100},
		kpis:       []model.ManagerKPI{{ManagerName: "Alice"}},
	}
	handler := api.NewAnalyticsHandler(svc, &mockExportService{})

	t.Run("Dashboard", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		rr := httptest.NewRecorder()
		handler.Routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("Conversion", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/conversion", nil)
		rr := httptest.NewRecorder()
		handler.Routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("KPI", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/kpi", nil)
		rr := httptest.NewRecorder()
		handler.Routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})
}

func TestIntegrationHandler(t *testing.T) {
	svc := &mockIntegrationService{}
	handler := api.NewIntegrationHandler(svc)

	t.Run("Telegram", func(t *testing.T) {
		payload := model.TelegramWebhookRequest{}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/telegram/webhook", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		handler.Routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("Telephony", func(t *testing.T) {
		payload := model.TelephonyWebhookRequest{}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/telephony/webhook", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		handler.Routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})
}

func TestUserHandler(t *testing.T) {
	user := &model.User{ID: 1, Name: "Test Admin", Email: "admin@test.com", Role: "admin"}
	svc := &mockUserService{
		users: []model.User{*user},
		user:  user,
	}
	handler := api.NewUserHandler(svc)

	t.Run("ListUsers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		handler.Routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("CreateUser", func(t *testing.T) {
		body, _ := json.Marshal(model.CreateUserRequest{
			Name:     "New User",
			Email:    "new@test.com",
			Password: "password123",
			Role:     "admissions",
		})
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		handler.Routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rr.Code)
		}
	})
}

func TestStatusHistoryHandler(t *testing.T) {
	svc := &mockStatusHistoryService{
		history: []model.StatusHistory{
			{ID: 1, LeadID: 1, NewStatusID: 2},
		},
	}
	handler := api.NewStatusHistoryHandler(svc)

	t.Run("GetByLead", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/lead/1", nil)
		rr := httptest.NewRecorder()
		handler.Routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})
}

func TestHealthCheck(t *testing.T) {
	mux := chi.NewRouter()
	mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
