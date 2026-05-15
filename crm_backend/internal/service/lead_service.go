package service

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"crm_backend/internal/model"
	"crm_backend/pkg/events"
	"github.com/xuri/excelize/v2"
)

type LeadService struct {
	leadRepo      LeadRepository
	contactRepo   ContactRepository
	dealRepo      DealRepository
	userRepo      UserRepository
	notifications INotificationService
	statusHistory StatusHistoryRepository
	bus           *events.Bus
}

func NewLeadService(
	leadRepo LeadRepository,
	contactRepo ContactRepository,
	dealRepo DealRepository,
	userRepo UserRepository,
	notifications INotificationService,
	statusHistory StatusHistoryRepository,
	bus *events.Bus,
) *LeadService {
	return &LeadService{
		leadRepo:      leadRepo,
		contactRepo:   contactRepo,
		dealRepo:      dealRepo,
		userRepo:      userRepo,
		notifications: notifications,
		statusHistory: statusHistory,
		bus:           bus,
	}
}

func (s *LeadService) publish(evtType string, payload interface{}) {
	if s.bus != nil {
		s.bus.Publish(events.Event{Type: evtType, Payload: payload})
	}
}

func (s *LeadService) CreateLeadFromForm(ctx context.Context, req *model.CreateLeadRequest) (*model.Lead, error) {
	statusID := 1
	sourceID := 1
	if req.StatusID != nil {
		statusID = *req.StatusID
	}
	if req.SourceID != nil {
		sourceID = *req.SourceID
	}

	// Round-Robin Assignment (если ассайн не задан явно).
	if req.AssigneeID == nil {
		assigneeID, err := s.getNextManagerRoundRobin(ctx)
		if err == nil {
			req.AssigneeID = &assigneeID
		}
	}

	contact, err := s.contactRepo.FindOrCreate(ctx, req.FirstName, req.LastName, req.Email, req.Phone, req.TelegramID, req.WhatsAppID, req.VKID)
	if err != nil {
		return nil, fmt.Errorf("failed to sync contact: %w", err)
	}

	deal, err := s.dealRepo.Create(ctx, contact.ID, statusID, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to auto-create deal: %w", err)
	}

	// ProgramID is passed through from the request (no longer silently dropped)
	lead, err := s.leadRepo.Create(ctx, req, sourceID, statusID)
	if err != nil {
		return nil, fmt.Errorf("failed to create lead: %w", err)
	}

	err = s.leadRepo.LinkToContact(ctx, lead.ID, contact.ID)
	if err != nil {
		fmt.Printf("Warning: failed to link contact id to lead id: %v\n", err)
	}
	lead.ContactID = &contact.ID

	fmt.Printf("Successfully mapped Lead #%d to Contact #%d and Deal #%d\n", lead.ID, contact.ID, deal.ID)

	// Record initial status in audit trail
	_ = s.statusHistory.Create(ctx, lead.ID, nil, statusID, nil, "")

	welcomeMsg := "Ваша заявка принята, менеджер свяжется с вами в течение 15 минут"
	_ = s.notifications.SendEmail(lead.Email, "Welcome to MDIS", welcomeMsg)
	_ = s.notifications.SendSMSAlias(lead.Phone, welcomeMsg)

	// Re-fetch so we emit a fully hydrated lead (with program_name etc.)
	if full, ferr := s.leadRepo.GetByID(ctx, lead.ID); ferr == nil {
		s.publish("lead.created", full)
	} else {
		s.publish("lead.created", lead)
	}

	return lead, nil
}

func (s *LeadService) GetLead(ctx context.Context, id int) (*model.Lead, error) {
	return s.leadRepo.GetByID(ctx, id)
}

func (s *LeadService) ListLeads(ctx context.Context, limit, offset int) ([]model.Lead, int, error) {
	return s.leadRepo.List(ctx, limit, offset)
}

func (s *LeadService) UpdateLeadStatus(ctx context.Context, leadID int, statusID int, refusalReason string, examDate *time.Time) error {
	lead, err := s.leadRepo.GetByID(ctx, leadID)
	if err != nil {
		return fmt.Errorf("lead not found: %w", err)
	}

	oldStatusID := lead.StatusID

	err = s.leadRepo.UpdateStatus(ctx, leadID, statusID)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Record status change in audit trail
	_ = s.statusHistory.Create(ctx, leadID, oldStatusID, statusID, nil, refusalReason)

	if lead.ContactID != nil {
		_ = s.dealRepo.UpdateStatusByContactID(ctx, *lead.ContactID, statusID, refusalReason, examDate)
	}

	s.publish("lead.status_changed", map[string]interface{}{
		"lead_id":        leadID,
		"old_status_id":  oldStatusID,
		"new_status_id":  statusID,
		"refusal_reason": refusalReason,
	})

	return nil
}

func (s *LeadService) MergeLeads(ctx context.Context, targetLeadID, sourceLeadID int) error {
	return s.leadRepo.MergeLeads(ctx, targetLeadID, sourceLeadID)
}

func (s *LeadService) UpdateLead(ctx context.Context, leadID int, req *model.UpdateLeadRequest) (*model.Lead, error) {
	existing, err := s.leadRepo.GetByID(ctx, leadID)
	if err != nil {
		return nil, fmt.Errorf("lead not found: %w", err)
	}
	req.ContactID = existing.ContactID
	if err := s.leadRepo.Update(ctx, leadID, req); err != nil {
		return nil, err
	}
	updated, err := s.leadRepo.GetByID(ctx, leadID)
	if err == nil {
		s.publish("lead.updated", updated)
	}
	return updated, err
}

func (s *LeadService) DeleteLead(ctx context.Context, leadID int) error {
	if err := s.leadRepo.Delete(ctx, leadID); err != nil {
		return err
	}
	s.publish("lead.deleted", map[string]interface{}{"id": leadID})
	return nil
}

func (s *LeadService) getNextManagerRoundRobin(ctx context.Context) (int, error) {
	managers, err := s.userRepo.ListManagers(ctx)
	if err != nil || len(managers) == 0 {
		return 0, fmt.Errorf("no managers available")
	}

	lastID, _ := s.leadRepo.GetLastAssignedManagerID(ctx)
	
	nextIndex := 0
	for i, m := range managers {
		if m.ID == lastID {
			nextIndex = (i + 1) % len(managers)
			break
		}
	}

	nextID := managers[nextIndex].ID
	_ = s.leadRepo.UpdateLastAssignedManagerID(ctx, nextID)
	return nextID, nil
}

// ImportResult — сводка по итогам импорта (используется handler-ом для ответа JSON).
type ImportResult struct {
	Total    int      `json:"total"`
	Imported int      `json:"imported"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors,omitempty"`
}

// ImportLeads совместимая обёртка над ImportLeadsRich для интерфейса сервиса.
func (s *LeadService) ImportLeads(ctx context.Context, r io.Reader) error {
	_, err := s.ImportLeadsRich(ctx, r, nil, nil)
	return err
}

// ImportLeadsRich парсит файл «База лидов_03.04.2026.xlsx» (и совместимые шаблоны)
// и создаёт лиды. Поддерживает многолистовые книги, Excel-serial даты,
// колонки type (temperature) и status (лид/студент/след.год.).
//
// Аргументы programLookup/sourceLookup — кэш id-имён (см. ImportHandler).
// Можно передать nil — тогда импортируем без привязки к program/source.
func (s *LeadService) ImportLeadsRich(
	ctx context.Context,
	r io.Reader,
	programLookup func(ctx context.Context, name string) (int, bool),
	sourceLookup func(ctx context.Context, name string) (int, bool),
) (*ImportResult, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("not a valid xlsx: %w", err)
	}
	defer f.Close()

	result := &ImportResult{}
	for _, sheet := range f.GetSheetList() {
		rows, err := f.GetRows(sheet)
		if err != nil || len(rows) < 2 {
			continue
		}
		headerIdx := findHeaderRowInRows(rows)
		if headerIdx < 0 {
			continue
		}
		cols := mapImportColumns(rows[headerIdx])
		// Без имени/телефона лист не несёт лидов — пропускаем (он может быть справочником).
		if cols["name"] == -1 && cols["phone"] == -1 {
			continue
		}
		for i := headerIdx + 1; i < len(rows); i++ {
			req, ok := buildImportRequest(sheet, rows[i], cols)
			if !ok {
				continue
			}
			result.Total++
			// Привязка program/source если справочники переданы
			if programLookup != nil {
				prog := getColCell(rows[i], cols["program"])
				if prog == "" {
					prog = sheetToProgram(sheet)
				}
				if prog != "" {
					if id, ok := programLookup(ctx, prog); ok {
						req.ProgramID = &id
					}
				}
			}
			if sourceLookup != nil {
				src := getColCell(rows[i], cols["source"])
				if src != "" {
					if id, ok := sourceLookup(ctx, src); ok {
						req.SourceID = &id
					}
				}
			}
			req.UTMSource = "excel_import"
			req.UTMMedium = sheet
			if _, err := s.CreateLeadFromForm(ctx, req); err != nil {
				result.Skipped++
				if len(result.Errors) < 20 {
					result.Errors = append(result.Errors, fmt.Sprintf("%s r%d: %v", sheet, i+1, err))
				}
				continue
			}
			result.Imported++
		}
	}
	return result, nil
}

// ── helpers ──────────────────────────────────────────────────────────────

var importColAlias = map[string]string{
	"name": "name", "имя": "name", "фио": "name", "fio": "name",
	"phone": "phone", "number": "phone", "phone number": "phone",
	"phone number, email": "phone", "телефон": "phone", "номер": "phone",
	"email": "email", "e-mail": "email", "почта": "email", "maill": "email", "mail": "email",
	"source": "source", "lead source": "source", "источник": "source", "канал": "source",
	"program": "program", "programme": "program", "программа": "program",
	"факультет": "program", "faculty": "program",
	"day": "day", "date": "day", "дата": "day",
	"notes": "notes", "заметка": "notes", "заметки": "notes",
	"follow-up action": "notes", "enrollment progress": "notes", "comment": "notes",
	"type": "temperature", "тип": "temperature",
	"status": "status_label", "статус": "status_label",
	"social": "social", "соцсеть": "social", "соц.сеть": "social",
	"instagram": "social", "vk": "social", "facebook": "social",
	"test": "test", "тест": "test",
	"application form": "form_status",
}

func mapImportColumns(header []string) map[string]int {
	out := map[string]int{
		"name": -1, "phone": -1, "email": -1, "source": -1,
		"program": -1, "day": -1, "notes": -1,
		"temperature": -1, "status_label": -1, "social": -1, "test": -1, "form_status": -1,
	}
	for i, h := range header {
		key := strings.ToLower(strings.TrimSpace(h))
		if alias, ok := importColAlias[key]; ok {
			if out[alias] == -1 {
				out[alias] = i
			}
		}
	}
	return out
}

func findHeaderRowInRows(rows [][]string) int {
	for i := 0; i < len(rows) && i < 6; i++ {
		for _, cell := range rows[i] {
			key := strings.ToLower(strings.TrimSpace(cell))
			if _, ok := importColAlias[key]; ok {
				return i
			}
		}
	}
	return -1
}

func getColCell(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

var phoneCleanRx = regexp.MustCompile(`[^\d+]`)
var emailRx = regexp.MustCompile(`[\w.+-]+@[\w-]+\.[\w.-]+`)

func buildImportRequest(sheet string, row []string, cols map[string]int) (*model.CreateLeadRequest, bool) {
	rawName := getColCell(row, cols["name"])
	rawPhone := getColCell(row, cols["phone"])
	rawEmail := getColCell(row, cols["email"])

	if rawEmail == "" {
		if m := emailRx.FindString(rawPhone); m != "" {
			rawEmail = m
			rawPhone = strings.TrimSpace(strings.Replace(rawPhone, m, "", 1))
		}
	}
	phone := phoneCleanRx.ReplaceAllString(rawPhone, "")
	if len(phone) < 4 {
		phone = ""
	}
	if rawName == "" {
		return nil, false
	}
	parts := strings.Fields(rawName)
	fn := parts[0]
	ln := ""
	if len(parts) > 1 {
		ln = strings.Join(parts[1:], " ")
	}
	req := &model.CreateLeadRequest{
		FirstName: fn,
		LastName:  ln,
		Phone:     phone,
		Email:     rawEmail,
	}
	if dayStr := getColCell(row, cols["day"]); dayStr != "" {
		if t, err := parseExcelDate(dayStr); err == nil {
			req.CreatedAt = &t
		}
	}
	if social := getColCell(row, cols["social"]); social != "" && strings.Contains(social, "://") {
		req.SocialURL = social
	}
	return req, true
}

func sheetToProgram(sheet string) string {
	s := strings.ToUpper(strings.TrimSpace(sheet))
	switch {
	case strings.HasPrefix(s, "MBA"):
		return "Masters of Business Administration (MBA)"
	case strings.HasPrefix(s, "PCIE"):
		return "Professional Certificate in English (PCIE)"
	case strings.HasPrefix(s, "FYC"):
		return "BA (Hons) Business and Financial Management"
	}
	return ""
}

var importDateLayouts = []string{
	"2006-01-02 15:04:05",
	"2006-01-02",
	"02.01.2006",
	"02/01/2006",
	"01/02/2006",
}

func parseExcelDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	// Excel serial date (число дней с 1899-12-30)
	if serial, err := strconv.Atoi(s); err == nil && serial > 30000 && serial < 70000 {
		return time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC).Add(time.Duration(serial) * 24 * time.Hour), nil
	}
	for _, l := range importDateLayouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognised date %q", s)
}
