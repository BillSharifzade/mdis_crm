package service

import (
	"fmt"
	"net/smtp"
	"os"
)

type NotificationService struct {
	smtpHost string
	smtpPort string
	smtpUser string
	smtpPass string
}

func NewNotificationService() *NotificationService {
	return &NotificationService{
		smtpHost: os.Getenv("SMTP_HOST"),
		smtpPort: os.Getenv("SMTP_PORT"),
		smtpUser: os.Getenv("SMTP_USER"),
		smtpPass: os.Getenv("SMTP_PASS"),
	}
}

func (s *NotificationService) SendEmail(to string, subject string, body string) error {
	if s.smtpHost == "" {
		fmt.Printf("[Email Dry-Run] To: %s, Subject: %s, Body: %s\n", to, subject, body)
		return nil
	}

	auth := smtp.PlainAuth("", s.smtpUser, s.smtpPass, s.smtpHost)
	msg := []byte("Subject: " + subject + "\r\n" +
		"To: " + to + "\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"\r\n" +
		body + "\r\n")

	addr := s.smtpHost + ":" + s.smtpPort
	return smtp.SendMail(addr, auth, s.smtpUser, []string{to}, msg)
}

func (s *NotificationService) SendSMSAlias(phone string, message string) error {
	fmt.Printf("[SMS Logic] Sending to %s: %s\n", phone, message)
	return nil
}
