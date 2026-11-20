package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"

	"carhistory/internal/models"
	"carhistory/internal/repository"
	"carhistory/internal/services"
)

type ReminderWorker struct {
	reminderRepo     *repository.ReminderRepository
	notificationRepo *repository.NotificationRepository
	vehicleRepo      *repository.VehicleRepository
	userRepo         *repository.UserRepository
	emailSvc         *services.EmailService
	smsSvc           *services.SMSService
	cron             *cron.Cron
	schedule         string
}

func NewReminderWorker(
	reminderRepo *repository.ReminderRepository,
	notificationRepo *repository.NotificationRepository,
	vehicleRepo *repository.VehicleRepository,
	userRepo *repository.UserRepository,
	emailSvc *services.EmailService,
	smsSvc *services.SMSService,
	schedule string,
) *ReminderWorker {
	return &ReminderWorker{
		reminderRepo:     reminderRepo,
		notificationRepo: notificationRepo,
		vehicleRepo:      vehicleRepo,
		userRepo:         userRepo,
		emailSvc:         emailSvc,
		smsSvc:           smsSvc,
		cron:             cron.New(),
		schedule:         schedule,
	}
}

func (w *ReminderWorker) Start() {
	_, err := w.cron.AddFunc(w.schedule, func() {
		log.Println("[ReminderWorker] Running reminder check...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		w.processReminders(ctx)
	})
	if err != nil {
		log.Printf("[ReminderWorker] Failed to schedule job: %v", err)
		return
	}
	w.cron.Start()
	log.Printf("[ReminderWorker] Started with schedule: %s", w.schedule)
}

func (w *ReminderWorker) Stop() {
	w.cron.Stop()
	log.Println("[ReminderWorker] Stopped")
}

func (w *ReminderWorker) processReminders(ctx context.Context) {
	reminders, err := w.reminderRepo.FindActiveReminders(ctx)
	if err != nil {
		log.Printf("[ReminderWorker] Error fetching reminders: %v", err)
		return
	}

	log.Printf("[ReminderWorker] Processing %d active reminders", len(reminders))

	for _, rem := range reminders {
		w.processReminder(ctx, &rem)
	}

	_ = w.vehicleRepo.ExpireOldInvitations(ctx)
}

func (w *ReminderWorker) processReminder(ctx context.Context, rem *models.Reminder) {
	vehicle, err := w.vehicleRepo.FindByID(ctx, rem.VehicleID)
	if err != nil || vehicle == nil {
		return
	}

	triggered, message := w.shouldTrigger(rem, vehicle.CurrentOdometer)
	if !triggered {
		return
	}

	members, err := w.vehicleRepo.GetMembers(ctx, rem.VehicleID)
	if err != nil {
		return
	}

	vehicleName := rem.VehicleID.String()
	if vehicle.Nickname != nil {
		vehicleName = *vehicle.Nickname
	} else if vehicle.Brand != nil && vehicle.Model != nil {
		vehicleName = *vehicle.Brand + " " + *vehicle.Model
	}

	for _, member := range members {
		user, err := w.userRepo.FindByID(ctx, member.UserID)
		if err != nil || user == nil {
			continue
		}

		if user.EmailNotificationsEnabled && user.Email != nil {
			existing, _ := w.notificationRepo.FindPendingForReminder(ctx, rem.ID, "email")
			if existing == nil {
				n := w.createNotification(rem.ID, user.ID, vehicle.ID, "email", rem.Title, message)
				if err := w.notificationRepo.Create(ctx, n); err != nil {
					log.Printf("[ReminderWorker] Failed to create email notification: %v", err)
				} else {
					w.sendEmailNotification(ctx, n.ID, *user.Email, user.Name, vehicleName, rem.Title, message)
				}
			}
		}

		if user.SMSNotificationsEnabled && user.MobileNumber != nil {
			existing, _ := w.notificationRepo.FindPendingForReminder(ctx, rem.ID, "sms")
			if existing == nil {
				n := w.createNotification(rem.ID, user.ID, vehicle.ID, "sms", rem.Title, message)
				if err := w.notificationRepo.Create(ctx, n); err != nil {
					log.Printf("[ReminderWorker] Failed to create SMS notification: %v", err)
				} else {
					w.sendSMSNotification(ctx, n.ID, *user.MobileNumber, vehicleName, rem.Title, message)
				}
			}
		}
	}
}

func (w *ReminderWorker) shouldTrigger(rem *models.Reminder, currentKM float64) (bool, string) {
	now := time.Now()

	switch rem.ReminderType {
	case "date":
		return w.checkDateTrigger(rem, now)
	case "odometer":
		return w.checkOdometerTrigger(rem, currentKM)
	case "date_or_odometer":
		dateTriggered, dateMsg := w.checkDateTrigger(rem, now)
		kmTriggered, kmMsg := w.checkOdometerTrigger(rem, currentKM)
		if dateTriggered {
			return true, dateMsg
		}
		if kmTriggered {
			return true, kmMsg
		}
	}
	return false, ""
}

func (w *ReminderWorker) checkDateTrigger(rem *models.Reminder, now time.Time) (bool, string) {
	if rem.DueDate == nil {
		return false, ""
	}

	advanceDays := 30
	if rem.AdvanceDays != nil {
		advanceDays = *rem.AdvanceDays
	}

	daysLeft := int(rem.DueDate.Sub(now).Hours() / 24)

	if daysLeft < 0 {
		return true, rem.Title + " is overdue"
	}
	if daysLeft <= advanceDays {
		return true, rem.Title + " is due in " + daysString(daysLeft)
	}
	return false, ""
}

func (w *ReminderWorker) checkOdometerTrigger(rem *models.Reminder, currentKM float64) (bool, string) {
	if rem.DueOdometer == nil {
		return false, ""
	}

	advanceKM := 1000.0
	if rem.AdvanceKM != nil {
		advanceKM = *rem.AdvanceKM
	}

	kmLeft := *rem.DueOdometer - currentKM

	if kmLeft <= 0 {
		return true, rem.Title + " is overdue by odometer"
	}
	if kmLeft <= advanceKM {
		return true, rem.Title + " is due in " + kmString(kmLeft)
	}
	return false, ""
}

func (w *ReminderWorker) createNotification(reminderID, userID, vehicleID uuid.UUID, channel, subject, message string) *models.Notification {
	now := time.Now()
	return &models.Notification{
		ID:         uuid.New(),
		UserID:     userID,
		VehicleID:  &vehicleID,
		ReminderID: &reminderID,
		Channel:    channel,
		Status:     "pending",
		Subject:    &subject,
		Message:    message,
		CreatedAt:  now,
	}
}

func (w *ReminderWorker) sendEmailNotification(ctx context.Context, notifID uuid.UUID, email, userName, vehicleName, title, message string) {
	err := w.emailSvc.SendReminderNotification(email, userName, vehicleName, title, message)

	now := time.Now()
	if err != nil {
		errMsg := err.Error()
		_ = w.notificationRepo.UpdateStatus(ctx, notifID, "failed", nil, &errMsg)
		log.Printf("[ReminderWorker] Email failed to %s: %v", email, err)
	} else {
		_ = w.notificationRepo.UpdateStatus(ctx, notifID, "sent", &now, nil)
		log.Printf("[ReminderWorker] Email sent to %s", email)
	}
}

func (w *ReminderWorker) sendSMSNotification(ctx context.Context, notifID uuid.UUID, phone, vehicleName, title, message string) {
	err := w.smsSvc.SendReminderNotification(phone, vehicleName, title, message)

	now := time.Now()
	if err != nil {
		errMsg := err.Error()
		_ = w.notificationRepo.UpdateStatus(ctx, notifID, "failed", nil, &errMsg)
		log.Printf("[ReminderWorker] SMS failed to %s: %v", phone, err)
	} else {
		_ = w.notificationRepo.UpdateStatus(ctx, notifID, "sent", &now, nil)
		log.Printf("[ReminderWorker] SMS sent to %s", phone)
	}
}

func daysString(days int) string {
	if days == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", days)
}

func kmString(km float64) string {
	return fmt.Sprintf("%.0f KM", km)
}
