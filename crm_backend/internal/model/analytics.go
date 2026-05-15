package model

type DashboardSummary struct {
	TotalNewLeads             int     `json:"total_new_leads"`
	TotalDeals                int     `json:"total_deals"`
	TotalInSchool             int     `json:"total_in_school"`
	TotalPendingTasks         int     `json:"total_pending_tasks"`
	TotalLeads                int     `json:"total_leads"`
	TotalLost                 int     `json:"total_lost"`
	OverallConversion         float64 `json:"overall_conversion"`
	AverageProcessingSecs     float64 `json:"average_processing_secs"`
	AverageLeadToEnrolledDays float64 `json:"average_lead_to_enrolled_days"`
}

type ConversionReport struct {
	TotalRequests  int     `json:"total_requests"`
	TotalContracts int     `json:"total_contracts"`
	ConversionRate float64 `json:"conversion_rate"`
}

type ManagerKPI struct {
	ManagerID    int     `json:"manager_id"`
	ManagerName  string  `json:"manager_name"`
	LeadsCount   int     `json:"leads_count"`
	CallsCount   int     `json:"calls_count"`
	ClosedDeals  int     `json:"closed_deals"`
	RevenueAdded float64 `json:"revenue_added"`
}

// SourceBreakdown — статистика по utm_source / sources.name.
type SourceBreakdown struct {
	Source string  `json:"source"`
	Count  int     `json:"count"`
	Pct    float64 `json:"pct"`
}

// StageFunnelPoint — точка воронки на дашборде.
type StageFunnelPoint struct {
	StageID int    `json:"stage_id"`
	Name    string `json:"name"`
	Count   int    `json:"count"`
}

// LeadsTimeSeriesPoint — динамика заявок/зачислений по неделям.
type LeadsTimeSeriesPoint struct {
	Bucket   string `json:"bucket"` // "2026-W18" или ISO-date
	Leads    int    `json:"leads"`
	Enrolled int    `json:"enrolled"`
}
