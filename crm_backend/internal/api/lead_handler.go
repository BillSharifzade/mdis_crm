package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"crm_backend/internal/model"
	"crm_backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type LeadHandler struct {
	svc service.ILeadService
}

func NewLeadHandler(svc service.ILeadService) *LeadHandler {
	return &LeadHandler{svc: svc}
}

func (h *LeadHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.createLead)
	r.Get("/", h.listLeads)
	r.Get("/{id}", h.getLead)
	r.Put("/{id}", h.updateLead)
	r.Delete("/{id}", h.deleteLead)
	r.Patch("/{id}/status", h.updateLeadStatus)
	r.Post("/{id}/reminder/complete", h.completeReminder)
	r.Post("/merge", h.mergeLeads)
	r.Post("/import", h.importLeads)
	return r
}

// createLead godoc
// @Summary Create a new lead
// @Description Creates a new lead from a form submission, automatically linking a contact and opening a deal.
// @Tags leads
// @Accept  json
// @Produce  json
// @Param lead body model.CreateLeadRequest true "Create Lead Request"
// @Success 201 {object} model.Lead
// @Failure 400 {object} map[string]string
// @Router /leads [post]
func (h *LeadHandler) createLead(w http.ResponseWriter, r *http.Request) {
	var req model.CreateLeadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		return
	}

	if req.FirstName == "" || (req.Phone == "" && req.Email == "") {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "First name, and either phone or email are required"})
		return
	}

	lead, err := h.svc.CreateLeadFromForm(r.Context(), &req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create lead"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(lead)
}

type paginatedLeadsResponse struct {
	Data   []model.Lead `json:"data"`
	Total  int          `json:"total"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
}

// listLeads godoc
// @Summary List leads (paginated)
// @Description Get a paginated list of leads. Supports ?limit= and ?offset= query params.
// @Tags leads
// @Produce  json
// @Param limit query int false "Limit (default 50, max 500)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} paginatedLeadsResponse
// @Security Bearer
// @Router /leads [get]
func (h *LeadHandler) listLeads(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 50
	}

	leads, total, err := h.svc.ListLeads(r.Context(), limit, offset)
	if err != nil {
		log.Printf("listLeads error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to list leads"})
		return
	}

	if leads == nil {
		leads = []model.Lead{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(paginatedLeadsResponse{
		Data:   leads,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// getLead godoc
// @Summary Get a lead by ID
// @Description Get detailed information about a single lead.
// @Tags leads
// @Produce  json
// @Param id path int true "Lead ID"
// @Success 200 {object} model.Lead
// @Failure 404 {object} map[string]string
// @Security Bearer
// @Router /leads/{id} [get]
func (h *LeadHandler) getLead(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid lead ID"})
		return
	}

	lead, err := h.svc.GetLead(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Lead not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lead)
}

type updateStatusReq struct {
	StatusID      int        `json:"status_id" example:"2"`
	RefusalReason string     `json:"refusal_reason,omitempty" example:"Too expensive"`
	ExamDate      *time.Time `json:"exam_date,omitempty" example:"2026-04-10T10:00:00Z"`
}

// updateLeadStatus godoc
// @Summary Update lead status
// @Description Change the pipeline stage ID of an existing lead.
// @Tags leads
// @Accept json
// @Param id path int true "Lead ID"
// @Param status body updateStatusReq true "Status Update Request"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Security Bearer
// @Router /leads/{id}/status [patch]
func (h *LeadHandler) updateLeadStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	leadID, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid lead ID"})
		return
	}

	var req updateStatusReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid payload"})
		return
	}

	if req.StatusID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Valid status_id is required"})
		return
	}

	if req.StatusID == 7 && req.RefusalReason == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "refusal_reason is required when closing a deal as lost"})
		return
	}

	err = h.svc.UpdateLeadStatus(r.Context(), leadID, req.StatusID, req.RefusalReason, req.ExamDate)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update lead status"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// completeReminder godoc
// @Summary Mark a lead's follow-up reminder as done
// @Description Closes the calendar reminder so it stops appearing in notifications.
// @Tags leads
// @Param id path int true "Lead ID"
// @Success 204 "No Content"
// @Security Bearer
// @Router /leads/{id}/reminder/complete [post]
func (h *LeadHandler) completeReminder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid lead ID"})
		return
	}
	if err := h.svc.CompleteReminder(r.Context(), id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to complete reminder"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type mergeLeadsReq struct {
	TargetLeadID int `json:"target_lead_id"`
	SourceLeadID int `json:"source_lead_id"`
}

// mergeLeads godoc
// @Summary Merge two leads
// @Description Merges interactions from source lead into target lead and deletes source lead.
// @Tags leads
// @Accept json
// @Param request body mergeLeadsReq true "Merge Request"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Security Bearer
// @Router /leads/merge [post]
func (h *LeadHandler) mergeLeads(w http.ResponseWriter, r *http.Request) {
	var req mergeLeadsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		return
	}

	if req.TargetLeadID == 0 || req.SourceLeadID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "target_lead_id and source_lead_id are required"})
		return
	}

	err := h.svc.MergeLeads(r.Context(), req.TargetLeadID, req.SourceLeadID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to merge leads"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// updateLead godoc
// @Summary Update a lead
// @Description Update editable fields of a lead (name, contact info, program, assignee).
// @Tags leads
// @Accept json
// @Produce json
// @Param id path int true "Lead ID"
// @Param lead body model.UpdateLeadRequest true "Update Lead Request"
// @Success 200 {object} model.Lead
// @Security Bearer
// @Router /leads/{id} [put]
func (h *LeadHandler) updateLead(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid lead ID"})
		return
	}
	var req model.UpdateLeadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid payload"})
		return
	}
	if req.FirstName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "first_name is required"})
		return
	}
	lead, err := h.svc.UpdateLead(r.Context(), id, &req)
	if err != nil {
		log.Printf("updateLead error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update lead"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lead)
}

// deleteLead godoc
// @Summary Delete a lead
// @Description Permanently remove a lead and its interaction/status history.
// @Tags leads
// @Param id path int true "Lead ID"
// @Success 204 "No Content"
// @Security Bearer
// @Router /leads/{id} [delete]
func (h *LeadHandler) deleteLead(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid lead ID"})
		return
	}
	if err := h.svc.DeleteLead(r.Context(), id); err != nil {
		log.Printf("deleteLead error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete lead"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// importLeads godoc
// @Summary Import leads from Excel
// @Description Import multiple leads at once by uploading an Excel file.
// @Tags leads
// @Accept mpfd
// @Param file formData file true "Excel File"
// @Success 204 "No Content"
// @Security Bearer
// @Router /leads/import [post]
func (h *LeadHandler) importLeads(w http.ResponseWriter, r *http.Request) {
	// 30 MB
	if err := r.ParseMultipartForm(30 << 20); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "parse form: " + err.Error()})
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid file upload"})
		return
	}
	defer file.Close()

	// Передаём lookup-функции через сервисный слой — handler не должен знать о репозиториях,
	// поэтому здесь они nil и program/source резолвится позже (через UTM-medium=sheet).
	res, err := h.svc.ImportLeadsRich(r.Context(), file, nil, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Import failed: %v", err)})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
