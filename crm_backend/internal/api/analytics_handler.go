package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"crm_backend/internal/model"
	"crm_backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type AnalyticsHandler struct {
	svc    service.IAnalyticsService
	export service.IExportService
}

func NewAnalyticsHandler(svc service.IAnalyticsService, export service.IExportService) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc, export: export}
}

func (h *AnalyticsHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/dashboard", h.getDashboard)
	r.Get("/conversion", h.getConversion)
	r.Get("/kpi", h.getKPIs)
	r.Get("/sources", h.getSources)
	r.Get("/funnel", h.getFunnel)
	r.Get("/timeseries", h.getTimeSeries)
	r.Get("/export", h.exportLeads)
	return r
}

func (h *AnalyticsHandler) getSources(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.GetSourceBreakdown(r.Context())
	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	if out == nil {
		out = []model.SourceBreakdown{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (h *AnalyticsHandler) getFunnel(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.GetFunnel(r.Context())
	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	if out == nil {
		out = []model.StageFunnelPoint{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (h *AnalyticsHandler) getTimeSeries(w http.ResponseWriter, r *http.Request) {
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	out, err := h.svc.GetTimeSeries(r.Context(), days)
	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	if out == nil {
		out = []model.LeadsTimeSeriesPoint{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

// getDashboard godoc
// @Summary Full CRM dashboard summary
// @Description Real-time stats for leads, total deals, and student intake.
// @Tags analytics
// @Produce json
// @Success 200 {object} model.DashboardSummary
// @Security Bearer
// @Router /analytics/dashboard [get]
func (h *AnalyticsHandler) getDashboard(w http.ResponseWriter, r *http.Request) {
	var summary *model.DashboardSummary
	var err error
	summary, err = h.svc.GetDashboard(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to load dashboard"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// getConversion godoc
// @Summary Pipeline conversion report
// @Description Calculates conversion ratios from lead request to signed contract.
// @Tags analytics
// @Produce json
// @Success 200 {object} model.ConversionReport
// @Security Bearer
// @Router /analytics/conversion [get]
func (h *AnalyticsHandler) getConversion(w http.ResponseWriter, r *http.Request) {
	var report *model.ConversionReport
	var err error
	report, err = h.svc.GetConversion(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to load conversion report"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// getKPIs godoc
// @Summary Manager performance KPI
// @Description Review activity (calls) and outcome (success deals) per admissions manager.
// @Tags analytics
// @Produce json
// @Success 200 {array} model.ManagerKPI
// @Security Bearer
// @Router /analytics/kpi [get]
func (h *AnalyticsHandler) getKPIs(w http.ResponseWriter, r *http.Request) {
	var kpis []model.ManagerKPI
	var err error
	kpis, err = h.svc.GetKPIs(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to load manager KPIs"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(kpis)
}

// exportLeads godoc
// @Summary Export leads to Excel or PDF
// @Description Downloads a formatted file of all leads.
// @Tags analytics
// @Param format query string true "Format (xlsx, pdf)"
// @Success 200 {file} file
// @Security Bearer
// @Router /analytics/export [get]
func (h *AnalyticsHandler) exportLeads(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")

	switch format {
	case "xlsx":
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename=leads_export.xlsx")
		if err := h.export.GenerateLeadsExcel(r.Context(), w); err != nil {
			http.Error(w, "Failed to generate Excel", http.StatusInternalServerError)
			return
		}

	case "pdf":
		doc, err := h.export.GenerateLeadsPDF(r.Context())
		if err != nil {
			http.Error(w, "Failed to generate PDF", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=leads_report.pdf")
		w.Write(doc)

	default:
		http.Error(w, "Invalid format. Supported: xlsx, pdf", http.StatusBadRequest)
	}
}

