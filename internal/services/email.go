package services

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"strings"

	"carhistory/internal/config"
)

type EmailService struct {
	cfg config.SMTPConfig
}

func NewEmailService(cfg config.SMTPConfig) *EmailService {
	return &EmailService{cfg: cfg}
}

func (s *EmailService) send(to, subject, body string) error {
	if s.cfg.Username == "" {
		log.Printf("[EMAIL SKIP] To: %s | Subject: %s", to, subject)
		return nil
	}

	from := fmt.Sprintf("%s <%s>", s.cfg.FromName, s.cfg.FromEmail)
	msg := strings.Join([]string{
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=\"UTF-8\"",
		"From: " + from,
		"To: " + to,
		"Subject: " + subject,
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)

	tlsCfg := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         s.cfg.Host,
	}

	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		host, _, _ := net.SplitHostPort(addr)
		rawConn, dialErr := net.Dial("tcp", addr)
		if dialErr != nil {
			log.Printf("[EMAIL ERROR] Dial failed to %s: %v", addr, dialErr)
			return dialErr
		}
		client, clientErr := smtp.NewClient(rawConn, host)
		if clientErr != nil {
			return clientErr
		}
		if startErr := client.StartTLS(tlsCfg); startErr != nil {
			return startErr
		}
		if authErr := client.Auth(auth); authErr != nil {
			return authErr
		}
		if mailErr := client.Mail(s.cfg.FromEmail); mailErr != nil {
			return mailErr
		}
		if rcptErr := client.Rcpt(to); rcptErr != nil {
			return rcptErr
		}
		w, dataErr := client.Data()
		if dataErr != nil {
			return dataErr
		}
		if _, writeErr := fmt.Fprint(w, msg); writeErr != nil {
			return writeErr
		}
		if closeErr := w.Close(); closeErr != nil {
			return closeErr
		}
		client.Quit()
		log.Printf("[EMAIL SENT] To: %s | Subject: %s", to, subject)
		return nil
	}

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return err
	}
	defer client.Quit()

	if err := client.Auth(auth); err != nil {
		return err
	}
	if err := client.Mail(s.cfg.FromEmail); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := fmt.Fprint(w, msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	log.Printf("[EMAIL SENT] To: %s | Subject: %s", to, subject)
	return nil
}

func (s *EmailService) SendPasswordReset(to, name, resetURL string) {
	subject := "Reset Your Car History Password"
	body := fmt.Sprintf(`
		<h2>Hello %s,</h2>
		<p>You requested a password reset. Click the link below to reset your password:</p>
		<p><a href="%s">Reset Password</a></p>
		<p>This link expires in 30 minutes.</p>
		<p>If you did not request this, please ignore this email.</p>
	`, name, resetURL)
	if err := s.send(to, subject, body); err != nil {
		log.Printf("[EMAIL ERROR] SendPasswordReset: %v", err)
	}
}

func (s *EmailService) SendReminderNotification(to, name, vehicleName, reminderTitle, message string) error {
	subject := fmt.Sprintf("Reminder: %s — %s", vehicleName, reminderTitle)
	body := fmt.Sprintf(`
		<h2>Hello %s,</h2>
		<p>This is a reminder for your vehicle <strong>%s</strong>:</p>
		<p>%s</p>
		<p>Please log in to Car History to take action.</p>
	`, name, vehicleName, message)
	return s.send(to, subject, body)
}

func (s *EmailService) SendVehicleInvitation(to, inviterName, vehicleName, role, acceptURL string) {
	subject := fmt.Sprintf("%s has invited you to a vehicle on Car History", inviterName)
	body := fmt.Sprintf(`
		<h2>Vehicle Invitation</h2>
		<p><strong>%s</strong> has invited you to access <strong>%s</strong> as a <strong>%s</strong>.</p>
		<p>Click below to accept the invitation:</p>
		<p><a href="%s">Accept Invitation</a></p>
		<p>This invitation expires in 72 hours.</p>
	`, inviterName, vehicleName, role, acceptURL)
	if err := s.send(to, subject, body); err != nil {
		log.Printf("[EMAIL ERROR] SendVehicleInvitation: %v", err)
	}
}
