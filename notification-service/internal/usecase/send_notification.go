package usecase

import (
	"notification-service/internal/domain"
	"notification-service/internal/port"
)

type SendNotification struct {
	consumer           port.EventConsumer
	notificationSender port.NotificationSender
}

func NewSendNotification(
	consumer port.EventConsumer,
	notificationSender port.NotificationSender,
) *SendNotification {
	return &SendNotification{
		consumer:           consumer,
		notificationSender: notificationSender,
	}
}

func (sn *SendNotification) Execute(notification domain.Notification) error {
	return sn.notificationSender.Send(notification)
}
