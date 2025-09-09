package sender

import (
	"log"
	"notification-service/internal/domain"
	"notification-service/internal/port"
)

type LogNotificationSender struct{}

func NewLogNotificationSender() port.NotificationSender {
	return &LogNotificationSender{}
}

func (s *LogNotificationSender) Send(notification domain.Notification) error {
	log.Printf("📩 Notificação enviada: User=%s, Msg=%s\n", notification.UserID, notification.Message)
	return nil
}
