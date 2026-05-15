package service

import (
	"context"
	"io"
	"time"

	"crm_backend/internal/model"
)

// ---- Repository Interfaces ----

type LeadRepository interface {
	Create(ctx context.Context, lead *model.CreateLeadRequest, sourceID int, statusID int) (*model.Lead, error)
	GetByID(ctx context.Context, id int) (*model.Lead, error)
	List(ctx context.Context, limit, offset int) ([]model.Lead, int, error)
	UpdateStatus(ctx context.Context, leadID int, statusID int) error
	Update(ctx context.Context, leadID int, req *model.UpdateLeadRequest) error
	Delete(ctx context.Context, leadID int) error
	LinkToContact(ctx context.Context, leadID int, contactID int) error
	MergeLeads(ctx context.Context, targetLeadID, sourceLeadID int) error
	GetLastAssignedManagerID(ctx context.Context) (int, error)
	UpdateLastAssignedManagerID(ctx context.Context, id int) error
	FindByContactID(ctx context.Context, contactID int) (*model.Lead, error)
}

type ContactRepository interface {
	FindOrCreate(ctx context.Context, firstName, lastName, email, phone, telegramID, whatsappID, vkID string) (*model.Contact, error)
}

type DealRepository interface {
	Create(ctx context.Context, contactID, stageID, sourceID int) (*model.Deal, error)
	UpdateStatusByContactID(ctx context.Context, contactID int, stageID int, refusalReason string, examDate *time.Time) error
}

type InteractionRepository interface {
	Create(ctx context.Context, req *model.CreateInteractionRequest) (*model.Interaction, error)
	ListByLead(ctx context.Context, leadID int) ([]model.Interaction, error)
}

type AnalyticsRepository interface {
	GetDashboardSummary(ctx context.Context) (*model.DashboardSummary, error)
	GetManagerKPIs(ctx context.Context) ([]model.ManagerKPI, error)
	GetConversion(ctx context.Context) (*model.ConversionReport, error)
	GetSourceBreakdown(ctx context.Context) ([]model.SourceBreakdown, error)
	GetFunnel(ctx context.Context) ([]model.StageFunnelPoint, error)
	GetTimeSeries(ctx context.Context, days int) ([]model.LeadsTimeSeriesPoint, error)
}

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	ListManagers(ctx context.Context) ([]model.User, error)
	ListAll(ctx context.Context) ([]model.User, error)
	CreateUser(ctx context.Context, name, email, passwordHash, role string) (*model.User, error)
	UpdateUser(ctx context.Context, id int, name, email, role, passwordHash string) (*model.User, error)
	DeleteUser(ctx context.Context, id int) error
}

type StatusHistoryRepository interface {
	Create(ctx context.Context, leadID int, oldStatusID *int, newStatusID int, changedBy *int, refusalReason string) error
	GetByLeadID(ctx context.Context, leadID int) ([]model.StatusHistory, error)
	HasRecentTask(ctx context.Context, leadID int) (bool, error)
}

// ---- Service Interfaces ----

type ILeadService interface {
	CreateLeadFromForm(ctx context.Context, req *model.CreateLeadRequest) (*model.Lead, error)
	GetLead(ctx context.Context, id int) (*model.Lead, error)
	ListLeads(ctx context.Context, limit, offset int) ([]model.Lead, int, error)
	UpdateLeadStatus(ctx context.Context, leadID int, statusID int, refusalReason string, examDate *time.Time) error
	UpdateLead(ctx context.Context, leadID int, req *model.UpdateLeadRequest) (*model.Lead, error)
	DeleteLead(ctx context.Context, leadID int) error
	MergeLeads(ctx context.Context, targetLeadID, sourceLeadID int) error
	ImportLeads(ctx context.Context, r io.Reader) error
	ImportLeadsRich(ctx context.Context, r io.Reader,
		programLookup func(ctx context.Context, name string) (int, bool),
		sourceLookup func(ctx context.Context, name string) (int, bool)) (*ImportResult, error)
}

type IInteractionService interface {
	AddInteraction(ctx context.Context, req *model.CreateInteractionRequest) (*model.Interaction, error)
	GetLeadHistory(ctx context.Context, leadID int) ([]model.Interaction, error)
}

type IAnalyticsService interface {
	GetDashboard(ctx context.Context) (*model.DashboardSummary, error)
	GetKPIs(ctx context.Context) ([]model.ManagerKPI, error)
	GetConversion(ctx context.Context) (*model.ConversionReport, error)
	GetSourceBreakdown(ctx context.Context) ([]model.SourceBreakdown, error)
	GetFunnel(ctx context.Context) ([]model.StageFunnelPoint, error)
	GetTimeSeries(ctx context.Context, days int) ([]model.LeadsTimeSeriesPoint, error)
}

type IAuthService interface {
	Login(ctx context.Context, email, password string) (*model.LoginResponse, error)
}

type IIntegrationService interface {
	ProcessTelegramWebhook(ctx context.Context, payload *model.TelegramWebhookRequest) error
	ProcessTelephonyWebhook(ctx context.Context, payload *model.TelephonyWebhookRequest) error
}

type IExportService interface {
	GenerateLeadsExcel(ctx context.Context, w io.Writer) error
	GenerateLeadsPDF(ctx context.Context) ([]byte, error)
}

type INotificationService interface {
	SendEmail(to string, subject string, body string) error
	SendSMSAlias(phone string, message string) error
}

type IUserService interface {
	ListUsers(ctx context.Context) ([]model.User, error)
	CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.User, error)
	UpdateUser(ctx context.Context, id int, req *model.UpdateUserRequest) (*model.User, error)
	DeleteUser(ctx context.Context, id int) error
}

type IStatusHistoryService interface {
	GetLeadStatusHistory(ctx context.Context, leadID int) ([]model.StatusHistory, error)
	RecordStatusChange(ctx context.Context, leadID int, oldStatusID *int, newStatusID int, changedBy *int, refusalReason string) error
}
