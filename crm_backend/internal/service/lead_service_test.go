package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"crm_backend/internal/model"
	"crm_backend/internal/service"
)

type mockLeadRepo struct {
	createdLead *model.Lead
	leads       []model.Lead
	shouldError bool
	getStatus   int
}

func (m *mockLeadRepo) Create(ctx context.Context, lead *model.CreateLeadRequest, sourceID int, statusID int) (*model.Lead, error) {
	if m.shouldError {
		return nil, fmt.Errorf("error")
	}
	m.getStatus = statusID
	return m.createdLead, nil
}
func (m *mockLeadRepo) GetByID(ctx context.Context, id int) (*model.Lead, error) { 
	for _, l := range m.leads {
		if l.ID == id {
			return &l, nil
		}
	}
	return &model.Lead{ID: id}, nil 
}
func (m *mockLeadRepo) List(ctx context.Context, limit, offset int) ([]model.Lead, int, error) {
	return m.leads, len(m.leads), nil
}
func (m *mockLeadRepo) UpdateStatus(ctx context.Context, leadID int, statusID int) error {
	return nil
}
func (m *mockLeadRepo) LinkToContact(ctx context.Context, leadID int, contactID int) error {
	return nil
}
func (m *mockLeadRepo) MergeLeads(ctx context.Context, targetLeadID, sourceLeadID int) error {
	return nil
}
func (m *mockLeadRepo) GetLastAssignedManagerID(ctx context.Context) (int, error) { return 0, nil }
func (m *mockLeadRepo) UpdateLastAssignedManagerID(ctx context.Context, id int) error { return nil }
func (m *mockLeadRepo) FindByContactID(ctx context.Context, contactID int) (*model.Lead, error) {
	if m.createdLead != nil {
		return m.createdLead, nil
	}
	return nil, fmt.Errorf("not found")
}

type mockUserRepo struct {
	managers []model.User
	users    []model.User
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) { return nil, nil }
func (m *mockUserRepo) ListManagers(ctx context.Context) ([]model.User, error) { return m.managers, nil }
func (m *mockUserRepo) ListAll(ctx context.Context) ([]model.User, error) { return m.users, nil }
func (m *mockUserRepo) CreateUser(ctx context.Context, name, email, passwordHash, role string) (*model.User, error) {
	return &model.User{ID: 99, Name: name, Email: email, Role: role}, nil
}

type mockContactRepo struct {
	contact *model.Contact
	err     error
}

func (m *mockContactRepo) FindOrCreate(ctx context.Context, firstName, lastName, email, phone, telegramID, whatsappID, vkID string) (*model.Contact, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.contact, nil
}

type mockDealRepo struct {
	deal *model.Deal
}

func (m *mockDealRepo) Create(ctx context.Context, contactID, stageID, sourceID int) (*model.Deal, error) {
	return m.deal, nil
}
func (m *mockDealRepo) UpdateStatusByContactID(ctx context.Context, contactID int, stageID int, refusalReason string, examDate *time.Time) error {
	return nil
}

type mockNotificationService struct{}

func (m *mockNotificationService) SendEmail(to, subject, body string) error { return nil }
func (m *mockNotificationService) SendSMSAlias(phone, message string) error { return nil }

type mockStatusHistoryRepo struct{}

func (m *mockStatusHistoryRepo) Create(ctx context.Context, leadID int, oldStatusID *int, newStatusID int, changedBy *int, refusalReason string) error {
	return nil
}
func (m *mockStatusHistoryRepo) GetByLeadID(ctx context.Context, leadID int) ([]model.StatusHistory, error) {
	return nil, nil
}
func (m *mockStatusHistoryRepo) HasRecentTask(ctx context.Context, leadID int) (bool, error) {
	return false, nil
}

func TestLeadService_CreateLeadFromForm(t *testing.T) {
	contactRepo := &mockContactRepo{
		contact: &model.Contact{ID: 1, FirstName: "John"},
	}
	dealRepo := &mockDealRepo{
		deal: &model.Deal{ID: 10},
	}
	expectedLead := &model.Lead{
		ID:        100,
		FirstName: "John",
		CreatedAt: time.Now(),
	}
	leadRepo := &mockLeadRepo{
		createdLead: expectedLead,
	}

	userRepo := &mockUserRepo{
		managers: []model.User{{ID: 1, Name: "Manager 1", Role: "admissions"}},
	}

	svc := service.NewLeadService(leadRepo, contactRepo, dealRepo, userRepo, &mockNotificationService{}, &mockStatusHistoryRepo{}, nil)

	req := &model.CreateLeadRequest{
		FirstName: "John",
		Email:     "john@example.com",
	}

	lead, err := svc.CreateLeadFromForm(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if lead.ID != expectedLead.ID {
		t.Errorf("expected lead ID %d, got %d", expectedLead.ID, lead.ID)
	}

	if leadRepo.getStatus != 1 {
		t.Errorf("expected status 'Новая заявка' (1), got %d", leadRepo.getStatus)
	}
}
