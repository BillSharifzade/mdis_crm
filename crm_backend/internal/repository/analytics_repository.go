package repository

import (
	"context"
	"fmt"

	"crm_backend/internal/model"
	"crm_backend/pkg/database"
)

type AnalyticsRepository struct {
	db *database.DB
}

func NewAnalyticsRepository(db *database.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func (r *AnalyticsRepository) GetDashboardSummary(ctx context.Context) (*model.DashboardSummary, error) {
	// За «зачислен» считаем lead со status_id из этапа is_final=true и name содержит «Зачисление»
	// fallback на статический 6 (зачисление).
	query := `
		WITH
		enroll_id AS (
			SELECT COALESCE(
				(SELECT id FROM pipeline_stages WHERE LOWER(name) LIKE '%зачисл%' LIMIT 1),
				6
			) AS sid
		),
		lost_id AS (
			SELECT COALESCE(
				(SELECT id FROM pipeline_stages WHERE LOWER(name) LIKE '%отказ%' OR LOWER(name) LIKE '%проигр%' LIMIT 1),
				7
			) AS sid
		),
		summary AS (
			SELECT
				(SELECT count(*) FROM leads WHERE status_id = 1) AS total_new_leads,
				(SELECT count(*) FROM deals) AS total_deals,
				(SELECT count(*) FROM leads WHERE status_id = (SELECT sid FROM enroll_id)) AS total_in_school,
				(SELECT count(*) FROM leads WHERE status_id = (SELECT sid FROM lost_id)) AS total_lost,
				(SELECT count(*) FROM interactions WHERE is_task = TRUE AND is_done = FALSE) AS total_pending_tasks,
				(SELECT count(*) FROM leads) AS total_leads,
				COALESCE(
					(SELECT count(*)::float8 FROM leads WHERE status_id = (SELECT sid FROM enroll_id))
					/ NULLIF((SELECT count(*) FROM leads), 0) * 100,
					0
				) AS overall_conversion,
				COALESCE((SELECT AVG(EXTRACT(EPOCH FROM (i.created_at - l.created_at)))
					FROM leads l
					JOIN (SELECT lead_id, MIN(created_at) AS created_at FROM interactions WHERE lead_id IS NOT NULL GROUP BY lead_id) i
					  ON l.id = i.lead_id
				), 0)::float8 AS average_processing_secs,
				COALESCE((SELECT AVG(EXTRACT(EPOCH FROM (sh.created_at - l.created_at)))/86400.0
					FROM leads l
					JOIN status_history sh ON sh.lead_id = l.id
					WHERE sh.new_status_id = (SELECT sid FROM enroll_id)
				), 0)::float8 AS avg_lead_to_enrolled_days
		)
		SELECT * FROM summary
	`

	var summary model.DashboardSummary
	if err := r.db.Pool.QueryRow(ctx, query).Scan(
		&summary.TotalNewLeads, &summary.TotalDeals, &summary.TotalInSchool, &summary.TotalLost,
		&summary.TotalPendingTasks, &summary.TotalLeads,
		&summary.OverallConversion, &summary.AverageProcessingSecs, &summary.AverageLeadToEnrolledDays,
	); err != nil {
		return nil, fmt.Errorf("failed to get dashboard summary: %w", err)
	}

	return &summary, nil
}

// GetManagerKPIs — реальные показатели по каждому admin/admissions.
// leads_count = лиды, на которых пользователь стоит в assignee.
// calls_count = interactions type='call' этим пользователем.
// closed_deals = leads.status_id = 'зачислен' для assignee.
func (r *AnalyticsRepository) GetManagerKPIs(ctx context.Context) ([]model.ManagerKPI, error) {
	query := `
		WITH enroll_id AS (
			SELECT COALESCE(
				(SELECT id FROM pipeline_stages WHERE LOWER(name) LIKE '%зачисл%' LIMIT 1),
				6
			) AS sid
		)
		SELECT
			u.id,
			u.name,
			(SELECT count(*) FROM leads l WHERE l.assignee_id = u.id) AS leads_count,
			(SELECT count(*) FROM interactions i WHERE i.created_by = u.id AND i.type = 'call') AS calls_count,
			(SELECT count(*) FROM leads l WHERE l.assignee_id = u.id AND l.status_id = (SELECT sid FROM enroll_id)) AS closed_deals
		FROM users u
		WHERE u.role IN ('admissions','admin')
		ORDER BY closed_deals DESC, leads_count DESC
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get manager kpis: %w", err)
	}
	defer rows.Close()

	var kpis []model.ManagerKPI
	for rows.Next() {
		var kpi model.ManagerKPI
		if err := rows.Scan(&kpi.ManagerID, &kpi.ManagerName, &kpi.LeadsCount, &kpi.CallsCount, &kpi.ClosedDeals); err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}
		kpis = append(kpis, kpi)
	}

	return kpis, nil
}

func (r *AnalyticsRepository) GetConversion(ctx context.Context) (*model.ConversionReport, error) {

	query := `
		WITH enroll_id AS (
			SELECT COALESCE(
				(SELECT id FROM pipeline_stages WHERE LOWER(name) LIKE '%зачисл%' LIMIT 1),
				6
			) AS sid
		)
		SELECT
			(SELECT count(*) FROM leads) AS total_requests,
			(SELECT count(*) FROM leads WHERE status_id = (SELECT sid FROM enroll_id)) AS total_contracts
	`
	var report model.ConversionReport
	if err := r.db.Pool.QueryRow(ctx, query).Scan(
		&report.TotalRequests, &report.TotalContracts,
	); err != nil {
		return nil, fmt.Errorf("failed to get conversion report: %w", err)
	}

	if report.TotalRequests > 0 {
		report.ConversionRate = float64(report.TotalContracts) / float64(report.TotalRequests) * 100
	} else {
		report.ConversionRate = 0
	}

	return &report, nil
}

// GetSourceBreakdown — распределение лидов по utm_source (или source.name если utm пуст).
func (r *AnalyticsRepository) GetSourceBreakdown(ctx context.Context) ([]model.SourceBreakdown, error) {
	q := `
		WITH src AS (
			SELECT COALESCE(NULLIF(l.utm_source, ''), s.name, 'unknown') AS source
			FROM leads l LEFT JOIN sources s ON s.id = l.source_id
		),
		total AS (SELECT COUNT(*)::float8 AS n FROM src)
		SELECT src.source, COUNT(*) AS cnt,
		       CASE WHEN total.n = 0 THEN 0 ELSE ROUND(COUNT(*) * 100.0 / total.n, 1) END AS pct
		FROM src, total
		GROUP BY src.source, total.n
		ORDER BY cnt DESC
		LIMIT 12
	`
	rows, err := r.db.Pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.SourceBreakdown{}
	for rows.Next() {
		var s model.SourceBreakdown
		if err := rows.Scan(&s.Source, &s.Count, &s.Pct); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

// GetFunnel — текущее число лидов на каждом этапе (без отказа).
func (r *AnalyticsRepository) GetFunnel(ctx context.Context) ([]model.StageFunnelPoint, error) {
	q := `
		SELECT s.id, s.name, COUNT(l.id) AS cnt
		FROM pipeline_stages s
		LEFT JOIN leads l ON l.status_id = s.id
		WHERE s.is_active = TRUE
		GROUP BY s.id, s.name, s."order"
		ORDER BY s."order"
	`
	rows, err := r.db.Pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.StageFunnelPoint{}
	for rows.Next() {
		var p model.StageFunnelPoint
		if err := rows.Scan(&p.StageID, &p.Name, &p.Count); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

// GetTimeSeries — динамика заявок/зачислений по неделям за N последних дней.
func (r *AnalyticsRepository) GetTimeSeries(ctx context.Context, days int) ([]model.LeadsTimeSeriesPoint, error) {
	if days <= 0 {
		days = 56 // 8 недель
	}
	q := `
		WITH enroll_id AS (
			SELECT COALESCE(
				(SELECT id FROM pipeline_stages WHERE LOWER(name) LIKE '%зачисл%' LIMIT 1),
				6
			) AS sid
		),
		buckets AS (
			SELECT generate_series(
				date_trunc('week', NOW() - ($1::int * INTERVAL '1 day')),
				date_trunc('week', NOW()),
				'1 week'
			) AS wk
		),
		lead_counts AS (
			SELECT date_trunc('week', created_at) AS wk, COUNT(*) AS cnt
			FROM leads
			WHERE created_at >= NOW() - ($1::int * INTERVAL '1 day')
			GROUP BY date_trunc('week', created_at)
		),
		enroll_counts AS (
			SELECT date_trunc('week', sh.created_at) AS wk, COUNT(DISTINCT sh.lead_id) AS cnt
			FROM status_history sh, enroll_id
			WHERE sh.new_status_id = enroll_id.sid
			  AND sh.created_at >= NOW() - ($1::int * INTERVAL '1 day')
			GROUP BY date_trunc('week', sh.created_at)
		)
		SELECT to_char(b.wk, 'YYYY-MM-DD') AS bucket,
		       COALESCE(l.cnt, 0) AS leads_cnt,
		       COALESCE(e.cnt, 0) AS enrolled_cnt
		FROM buckets b
		LEFT JOIN lead_counts l ON l.wk = b.wk
		LEFT JOIN enroll_counts e ON e.wk = b.wk
		ORDER BY b.wk
	`
	rows, err := r.db.Pool.Query(ctx, q, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.LeadsTimeSeriesPoint{}
	for rows.Next() {
		var p model.LeadsTimeSeriesPoint
		if err := rows.Scan(&p.Bucket, &p.Leads, &p.Enrolled); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}
