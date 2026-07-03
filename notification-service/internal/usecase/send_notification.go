package usecase

import (
	"notification-service/internal/domain"
	"notification-service/internal/port"
)

type SendNotification struct {
	notificationSender port.NotificationSender
}

func NewSendNotification(
	notificationSender port.NotificationSender,
) *SendNotification {
	return &SendNotification{
		notificationSender: notificationSender,
	}
}

func (sn *SendNotification) Execute(notification domain.Notification) error {
	return sn.notificationSender.Send(notification)
}
