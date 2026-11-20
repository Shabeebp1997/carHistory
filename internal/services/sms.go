package services

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"carhistory/internal/config"
)

type SMSService struct {
	cfg config.SMSConfig
}

func NewSMSService(cfg config.SMSConfig) *SMSService {
	return &SMSService{cfg: cfg}
}

func (s *SMSService) Send(to, message string) error {
	if s.cfg.AccountSID == "" || s.cfg.AuthToken == "" {
		log.Printf("[SMS SKIP] To: %s | Message: %s", to, message)
		return nil
	}

	apiURL := fmt.Sprintf(
		"https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json",
		s.cfg.AccountSID,
	)

	data := url.Values{}
	data.Set("To", to)
	data.Set("From", s.cfg.FromNumber)
	data.Set("Body", message)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("sms: failed to build request: %w", err)
	}

	req.SetBasicAuth(s.cfg.AccountSID, s.cfg.AuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[SMS ERROR] To: %s | Error: %v", to, err)
		return fmt.Errorf("sms: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf("[SMS ERROR] To: %s | HTTP Status: %d", to, resp.StatusCode)
		return fmt.Errorf("sms: twilio returned status %d", resp.StatusCode)
	}

	log.Printf("[SMS SENT] To: %s", to)
	return nil
}

func (s *SMSService) SendReminderNotification(to, vehicleName, title, message string) error {
	body := fmt.Sprintf("[Car History] %s — %s: %s", vehicleName, title, message)
	return s.Send(to, body)
}

func (s *SMSService) SendPasswordResetOTP(to, otp string) error {
	body := fmt.Sprintf("[Car History] Your password reset OTP is: %s. Valid for 30 minutes.", otp)
	return s.Send(to, body)
}

func (s *SMSService) SendVehicleInvitation(to, inviterName, vehicleName string) error {
	body := fmt.Sprintf("[Car History] %s has invited you to access vehicle: %s. Open the app to accept.", inviterName, vehicleName)
	return s.Send(to, body)
}
