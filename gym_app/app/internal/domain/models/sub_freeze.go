package models

import "time"

type SubscriptionFreeze struct {
	ID                 int       `json:"id"`
	SubscriptionNumber string    `json:"subscription_number" validate:"required"` // Связь с PersonSubscription.Number
	FreezeStart        time.Time `json:"freeze_start" validate:"required"`
	FreezeEnd          time.Time `json:"freeze_end,omitempty"` // Может быть nil, если еще не разморожен
	DaysUsed           int       `json:"days_used,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}
