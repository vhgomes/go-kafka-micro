package port

import "notification-service/internal/domain"

type NotificationSender interface {
	Send(notification domain.Notification) error
}
