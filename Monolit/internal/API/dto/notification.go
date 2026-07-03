package dto

type NotificationsResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	UnreadCount   int                    `json:"unread_count"`
}

type NotificationResponse struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	Title      string  `json:"title"`
	Body       string  `json:"body"`
	EntityType *string `json:"entity_type"`
	EntityUUID *string `json:"entity_uuid"`
	ReadAt     *string `json:"read_at"`
	CreatedAt  string  `json:"created_at"`
}
