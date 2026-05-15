package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"crm_backend/internal/model"
	"crm_backend/internal/repository"
	"crm_backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/xuri/excelize/v2"
)

// ImportHandler принимает Excel-файл "База лидов" и создаёт лидов в CRM.
// Структура файла: по листу на программу (Free Sessions, EET, FYC, MBA, ...).
// Колонки в одной книге могут различаться — мы детектируем заголовок по словам:
//   Day/Date, Name, number/phone/email, source/Lead source, Programme/Faculty,
//   Enrollment progress, Follow-up action, type (холодный/теплый/горячий),
//   status (лид/студент/след.год.), Test
// Распознанная пара (имя, телефон) — обязательный минимум. Иначе строка скип.
type ImportHandler struct {
	leadSvc  service.ILeadService
	settings *repository.SettingsRepository
}

func NewImportHandler(leadSvc service.ILeadService, settings *repository.SettingsRepository) *ImportHandler {
	return &ImportHandler{leadSvc: leadSvc, settings: settings}
}

func (h *ImportHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/preview", h.preview) // только парсит, ничего не пишет
	r.Post("/commit", h.commit)   // парсит и создаёт лидов
	return r
}

const maxImportBytes = 30 << 20 // 30 MB

type importRow struct {
	Sheet       string  `json:"sheet"`
	Row         int     `json:"row"`
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	Phone       string  `json:"phone"`
	Email       string  `json:"email"`
	Source      string  `json:"source"`
	Program     string  `json:"program"`
	Notes       string  `json:"notes"`
	StatusHint  string  `json:"status_hint"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
}

type importResult struct {
	Total       int         `json:"total"`
	Imported    int         `json:"imported"`
	Skipped     int         `json:"skipped"`
	Errors      []string    `json:"errors,omitempty"`
	Sample      []importRow `json:"sample,omitempty"`
}

func (h *ImportHandler) preview(w http.ResponseWriter, r *http.Request) {
	rows, err := parseExcelUpload(r)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	res := importResult{Total: len(rows), Sample: rows}
	if len(res.Sample) > 50 {
		res.Sample = res.Sample[:50]
	}
	writeJSON(w, 200, res)
}

func (h *ImportHandler) commit(w http.ResponseWriter, r *http.Request) {
	rows, err := parseExcelUpload(r)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	ctx := r.Context()

	// Кэш id справочников
	progIDs := map[string]int{}
	srcIDs := map[string]int{}
	if list, err := h.settings.ListPrograms(ctx, false); err == nil {
		for _, p := range list {
			progIDs[strings.ToLower(p.Name)] = p.ID
		}
	}
	if list, err := h.settings.ListSources(ctx, false); err == nil {
		for _, s := range list {
			srcIDs[strings.ToLower(s.Name)] = s.ID
		}
	}

	res := importResult{Total: len(rows)}
	for _, row := range rows {
		req := &model.CreateLeadRequest{
			FirstName: row.FirstName,
			LastName:  row.LastName,
			Phone:     row.Phone,
			Email:     row.Email,
		}
		if row.CreatedAt != nil {
			t := *row.CreatedAt
			req.CreatedAt = &t
		}
		if row.Program != "" {
			if id, ok := progIDs[strings.ToLower(row.Program)]; ok {
				req.ProgramID = &id
			} else if id, err := h.ensureProgram(ctx, row.Program, progIDs); err == nil {
				req.ProgramID = &id
			}
		}
		if row.Source != "" {
			if id, ok := srcIDs[strings.ToLower(row.Source)]; ok {
				req.SourceID = &id
			} else if id, err := h.ensureSource(ctx, row.Source, srcIDs); err == nil {
				req.SourceID = &id
			}
		}
		req.UTMSource = "excel_import"
		req.UTMMedium = row.Sheet

		if _, err := h.leadSvc.CreateLeadFromForm(ctx, req); err != nil {
			res.Skipped++
			if len(res.Errors) < 20 {
				res.Errors = append(res.Errors,
					fmt.Sprintf("%s r%d: %v", row.Sheet, row.Row, err))
			}
			continue
		}
		res.Imported++
	}
	writeJSON(w, 200, res)
}

func (h *ImportHandler) ensureProgram(ctx context.Context, name string, cache map[string]int) (int, error) {
	p, err := h.settings.CreateProgram(ctx, name, "")
	if err != nil {
		return 0, err
	}
	cache[strings.ToLower(name)] = p.ID
	return p.ID, nil
}

func (h *ImportHandler) ensureSource(ctx context.Context, name string, cache map[string]int) (int, error) {
	s, err := h.settings.CreateSource(ctx, name)
	if err != nil {
		return 0, err
	}
	cache[strings.ToLower(name)] = s.ID
	return s.ID, nil
}

func parseExcelUpload(r *http.Request) ([]importRow, error) {
	r.Body = http.MaxBytesReader(nil, r.Body, maxImportBytes)
	if err := r.ParseMultipartForm(maxImportBytes); err != nil {
		return nil, fmt.Errorf("parse form: %w", err)
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("file is required (form field 'file')")
	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	f, err := excelize.OpenReader(strings.NewReader(string(buf)))
	if err != nil {
		return nil, fmt.Errorf("not a valid xlsx: %w", err)
	}
	defer f.Close()

	out := []importRow{}
	for _, sheet := range f.GetSheetList() {
		rows, err := f.GetRows(sheet)
		if err != nil || len(rows) < 2 {
			continue
		}
		headerIdx := findHeaderRow(rows)
		if headerIdx < 0 {
			continue
		}
		header := rows[headerIdx]
		cols := mapColumns(header)
		// Если нет ни имени, ни телефона — пропускаем лист
		if cols["name"] == -1 && cols["phone"] == -1 {
			continue
		}
		for i := headerIdx + 1; i < len(rows); i++ {
			row := rows[i]
			ir, ok := buildRow(sheet, i+1, row, cols)
			if !ok {
				continue
			}
			out = append(out, ir)
		}
	}
	return out, nil
}

var phoneCleanRe = regexp.MustCompile(`[^\d+]`)
var emailRe = regexp.MustCompile(`[\w.+-]+@[\w-]+\.[\w.-]+`)

// columnAlias — какие заголовки маппим в какие логические колонки.
var columnAlias = map[string]string{
	"name":               "name",
	"имя":                "name",
	"фио":                "name",
	"phone":              "phone",
	"number":             "phone",
	"phone number":       "phone",
	"phone number, email": "phone",
	"телефон":            "phone",
	"номер":              "phone",
	"email":              "email",
	"e-mail":             "email",
	"почта":              "email",
	"maill":              "email",
	"mail":               "email",
	"source":             "source",
	"lead source":        "source",
	"источник":           "source",
	"канал":              "source",
	"program":            "program",
	"programme":          "program",
	"программа":          "program",
	"факультет":          "program",
	"faculty":            "program",
	"day":                "day",
	"date":               "day",
	"дата":               "day",
	"notes":              "notes",
	"заметка":            "notes",
	"заметки":            "notes",
	"follow-up action":   "notes",
	"enrollment progress": "notes",
	"comment":            "notes",
	"status":             "status",
	"статус":             "status",
	"type":               "status",
	"тип":                "status",
}

func mapColumns(header []string) map[string]int {
	out := map[string]int{
		"name": -1, "phone": -1, "email": -1, "source": -1,
		"program": -1, "day": -1, "notes": -1, "status": -1,
	}
	for i, h := range header {
		key := strings.ToLower(strings.TrimSpace(h))
		if alias, ok := columnAlias[key]; ok {
			if out[alias] == -1 {
				out[alias] = i
			}
		}
	}
	return out
}

// findHeaderRow ищет первую строку, где хотя бы одна ячейка из списка алиасов.
func findHeaderRow(rows [][]string) int {
	for i := 0; i < len(rows) && i < 6; i++ {
		for _, cell := range rows[i] {
			key := strings.ToLower(strings.TrimSpace(cell))
			if _, ok := columnAlias[key]; ok {
				return i
			}
		}
	}
	return -1
}

func cell(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

func splitName(full string) (first, last string) {
	parts := strings.Fields(full)
	if len(parts) == 0 {
		return "", ""
	}
	first = parts[0]
	if len(parts) > 1 {
		last = strings.Join(parts[1:], " ")
	}
	return
}

// buildRow собирает importRow или возвращает false если нет минимума (имя+что-то).
func buildRow(sheet string, rowIdx int, row []string, cols map[string]int) (importRow, bool) {
	rawName := cell(row, cols["name"])
	rawPhone := cell(row, cols["phone"])
	rawEmail := cell(row, cols["email"])

	// Иногда email сидит в "Phone number, email" — пробуем извлечь.
	if rawEmail == "" {
		if m := emailRe.FindString(rawPhone); m != "" {
			rawEmail = m
			rawPhone = strings.TrimSpace(strings.Replace(rawPhone, m, "", 1))
		}
	}

	phone := phoneCleanRe.ReplaceAllString(rawPhone, "")
	// Колонка phone может содержать буквы (например "telegram") — отбрасываем.
	if len(phone) < 4 {
		phone = ""
	}

	if rawName == "" && phone == "" && rawEmail == "" {
		return importRow{}, false
	}
	if rawName == "" {
		return importRow{}, false // имя обязательно
	}

	fn, ln := splitName(rawName)

	ir := importRow{
		Sheet:      sheet,
		Row:        rowIdx,
		FirstName:  fn,
		LastName:   ln,
		Phone:      phone,
		Email:      rawEmail,
		Source:     cell(row, cols["source"]),
		Program:    cell(row, cols["program"]),
		Notes:      strings.TrimSpace(cell(row, cols["notes"])),
		StatusHint: cell(row, cols["status"]),
	}

	// Эвристика: если в имени листа есть программа, а в строке нет — используем имя листа.
	if ir.Program == "" {
		s := strings.ToUpper(strings.TrimSpace(sheet))
		switch {
		case strings.HasPrefix(s, "MBA"):
			ir.Program = "Masters of Business Administration (MBA)"
		case strings.HasPrefix(s, "PCIE"):
			ir.Program = "Professional Certificate in English (PCIE)"
		case strings.HasPrefix(s, "FYC"):
			ir.Program = "BA (Hons) Business and Financial Management"
		case strings.HasPrefix(s, "EET") || strings.HasPrefix(s, "ЕЕТ"):
			ir.Program = "" // EET — это тест, не программа
		}
	}

	if dayStr := cell(row, cols["day"]); dayStr != "" {
		if t, err := parseDate(dayStr); err == nil {
			ir.CreatedAt = &t
		}
	}

	return ir, true
}

var dateLayouts = []string{
	"2006-01-02 15:04:05",
	"2006-01-02",
	"02.01.2006",
	"02/01/2006",
	"01/02/2006",
}

func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, l := range dateLayouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognised date %q", s)
}

// Ensure json import is used (silence unused if compiled without preview reachable)
var _ = json.Marshal
