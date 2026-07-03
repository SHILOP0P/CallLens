package models

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypeInvitation       NotificationType = "invitation"
	NotificationTypeReportReady      NotificationType = "report_ready"
	NotificationTypeSubscription     NotificationType = "subscription"
	NotificationTypeProcessingFailed NotificationType = "processing_failed"
)

type Notification struct {
	ID         uuid.UUID
	UserUUID   uuid.UUID
	Type       NotificationType
	Title      string
	Body       string
	EntityType *string
	EntityUUID uuid.NullUUID
	ReadAt     *time.Time
	CreatedAt  time.Time
}

type CreateNotificationInput struct {
	UserUUID   uuid.UUID
	Type       NotificationType
	Title      string
	Body       string
	EntityType *string
	EntityUUID uuid.NullUUID
	CreatedAt  time.Time
}

type ListNotificationsInput struct {
	UserUUID   uuid.UUID
	UnreadOnly bool
	Limit      int
	Offset     int
}

type ListNotificationsResult struct {
	Notifications []Notification
	UnreadCount   int
	Limit         int
	Offset        int
}
