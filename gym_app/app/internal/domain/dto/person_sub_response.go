package dto

import "time"

type PersonSubResponse struct {
	Number            string    `json:"number"`
	PersonID          int       `json:"person_id"`
	PersonName        string    `json:"person_name"`
	SubscriptionID    int       `json:"subscription_id"`
	SubscriptionTitle string    `json:"subscription_title,omitempty"`
	SubscriptionPrice float64   `json:"subscription_price,omitempty"` // <--- Добавить!
	StartDate         time.Time `json:"start_date,omitempty"`
	EndDate           time.Time `json:"end_date,omitempty"`
	Status            string    `json:"status,omitempty"`
	Discount          float64   `json:"discount,omitempty"` // Скидка в рублях
	FinalPrice        float64   `json:"final_price,omitempty"`
}
