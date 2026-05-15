package model

import "time"

// KPITarget — KPI-цель для одного пользователя (admissions).
// metric: "processed" — количество лидов, по которым менеджер сделал хоть одно
//                       действие (звонок/сообщение/заметка) за окно period_days.
//         "created"   — количество лидов, созданных самим менеджером.
type KPITarget struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Metric      string    `json:"metric"`
	TargetCount int       `json:"target_count"`
	PeriodDays  int       `json:"period_days"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// KPIStats — текущее состояние по пользователю.
type KPIStats struct {
	UserID         int  `json:"user_id"`
	Processed      int  `json:"processed"`        // лиды, по которым были действия за период
	Created        int  `json:"created"`          // лиды, созданные менеджером
	CallsCount     int  `json:"calls_count"`
	NotesCount     int  `json:"notes_count"`
	MessagesCount  int  `json:"messages_count"`
	PeriodDays     int  `json:"period_days"`
	ProcessedGoal  int  `json:"processed_goal,omitempty"`
	CreatedGoal    int  `json:"created_goal,omitempty"`
}

// CallLogRequest — структура запроса при сохранении звонка из UI.
type CallLogRequest struct {
	LeadID          int    `json:"lead_id"`
	Outcome         string `json:"outcome"`           // "answered" | "no_answer"
	Comment         string `json:"comment"`
	DurationMinutes int    `json:"duration_minutes"`  // только если answered
}
