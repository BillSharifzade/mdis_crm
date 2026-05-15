package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"crm_backend/internal/model"
	"crm_backend/pkg/database"
)

type CronService struct {
	db             *database.DB
	intService     IInteractionService
	notifService   INotificationService
	analyticsSvc   IAnalyticsService
	statusHistory  StatusHistoryRepository
}

func NewCronService(db *database.DB, intService IInteractionService, notifService INotificationService, analyticsSvc IAnalyticsService, statusHistory StatusHistoryRepository) *CronService {
	return &CronService{
		db:            db,
		intService:    intService,
		notifService:  notifService,
		analyticsSvc:  analyticsSvc,
		statusHistory: statusHistory,
	}
}

func (s *CronService) Start(ctx context.Context) {
	// Run every day
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		// Run once on launch for demo purposes
		s.RunDailyTasks(ctx)

		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				s.RunDailyTasks(ctx)
			}
		}
	}()
}

func (s *CronService) RunDailyTasks(ctx context.Context) {
	log.Println("[CRON] Running daily tasks...")
	s.checkUpcomingExams(ctx)
	s.checkStaleLeads(ctx)
	
	// Monthly report on the 1st day of the month
	if time.Now().Day() == 1 {
		s.generateMonthlyReport(ctx)
	}
}

func (s *CronService) generateMonthlyReport(ctx context.Context) {
	log.Println("[CRON] Generating monthly report...")
	summary, err := s.analyticsSvc.GetDashboard(ctx)
	if err != nil {
		log.Printf("[CRON] Error getting dashboard for report: %v", err)
		return
	}

	report := fmt.Sprintf("Системный отчет за месяц:\nВсего лидов: %d\nКонверсия: %.2f%%\nСр. скорость: %.1f сек.", 
		summary.TotalLeads, summary.OverallConversion, summary.AverageProcessingSecs)
	
	// Send to admin email (hardcoded for now or from env)
	s.notifService.SendEmail("admin@mdis_crm.edu.tj", "MDIS CRM: Консолидированный отчет", report)
}

func (s *CronService) checkUpcomingExams(ctx context.Context) {
	// Find deals with exam tomorrow
	query := `
		SELECT d.id, c.first_name, c.email, c.phone, d.exam_date
		FROM deals d
		JOIN contacts c ON d.contact_id = c.id
		WHERE d.exam_date >= CURRENT_DATE + INTERVAL '1 day'
		  AND d.exam_date < CURRENT_DATE + INTERVAL '2 days'
	`
	rows, err := s.db.Pool.Query(ctx, query)
	if err != nil {
		log.Printf("[CRON] Error querying exams: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var dealID int
		var firstName, email, phone string
		var examDate time.Time
		if err := rows.Scan(&dealID, &firstName, &email, &phone, &examDate); err != nil {
			continue
		}

		msg := fmt.Sprintf("Напоминание: у вас экзамен завтра (%s)", examDate.Format("2006-01-02"))
		if email != "" {
			s.notifService.SendEmail(email, "Напоминание об экзамене MDIS", msg)
		}
		if phone != "" {
			s.notifService.SendSMSAlias(phone, msg)
		}
	}
}

func (s *CronService) checkStaleLeads(ctx context.Context) {
	// leads that haven't been updated in 3 days and are not closed (6=Зачисление, 7=Отказ)
	query := `
		SELECT id, assignee_id
		FROM leads
		WHERE updated_at <= CURRENT_DATE - INTERVAL '3 days'
		  AND status_id NOT IN (6, 7)
	`
	rows, err := s.db.Pool.Query(ctx, query)
	if err != nil {
		log.Printf("[CRON] Error querying stale leads: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var leadID int
		var assigneeID *int
		if err := rows.Scan(&leadID, &assigneeID); err != nil {
			continue
		}

		// Idempotency check: skip if a task was already created today for this lead
		hasTask, err := s.statusHistory.HasRecentTask(ctx, leadID)
		if err != nil {
			log.Printf("[CRON] Error checking recent task for lead %d: %v", leadID, err)
			continue
		}
		if hasTask {
			continue
		}

		req := &model.CreateInteractionRequest{
			LeadID:  &leadID,
			Type:    "manual_note",
			Content: "Задача: Перезвонить! Статус заявки не менялся более 3 дней.",
			IsTask:  true,
		}
		if assigneeID != nil {
			req.CreatedBy = assigneeID
		}
		
		_, err = s.intService.AddInteraction(ctx, req)
		if err != nil {
			log.Printf("[CRON] Error adding task for lead %d: %v", leadID, err)
		}
	}
}
