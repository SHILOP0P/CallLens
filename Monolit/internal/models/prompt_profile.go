package models

import "github.com/google/uuid"

type PromptIndustry struct {
	Key         string `json:"key"`
	Perspective string `json:"perspective"`
	Title       string `json:"title"`
	SortOrder   int    `json:"sort_order"`
}
type PromptTopic struct {
	Key          string `json:"key"`
	IndustryKey  string `json:"industry_key"`
	Title        string `json:"title"`
	PromptModule string `json:"prompt_module"`
	SortOrder    int    `json:"sort_order"`
	Source       string `json:"source,omitempty"`
}
type PromptProfile struct {
	ID          uuid.UUID     `json:"id"`
	OwnerUserID uuid.UUID     `json:"owner_user_id"`
	Title       string        `json:"title"`
	Perspective string        `json:"perspective"`
	IndustryKey string        `json:"industry_key"`
	Description string        `json:"description"`
	IsDefault   bool          `json:"is_default"`
	Topics      []PromptTopic `json:"topics"`
}
type CallPromptContext struct {
	CallID      uuid.UUID `json:"call_uuid"`
	ProfileID   uuid.UUID `json:"profile_uuid"`
	OwnerUserID uuid.UUID `json:"owner_user_id"`
	TopicKeys   []string  `json:"topic_keys"`
}
type PromptUserSettings struct {
	UserID      uuid.UUID        `json:"user_id"`
	Description string           `json:"description"`
	Industries  []PromptIndustry `json:"industries"`
	Topics      []PromptTopic    `json:"topics"`
}
