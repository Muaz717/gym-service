package models

import (
	"github.com/go-playground/validator/v10"
	"time"
)

// PersonSubscription представляет подписку клиента на абонемент
type PersonSubscription struct {
	Number            string    `json:"number" validate:"required"`          // Номер абонемента
	PersonID          int       `json:"person_id" validate:"required"`       // ID клиента
	SubscriptionID    int       `json:"subscription_id" validate:"required"` // ID абонемента
	SubscriptionPrice float64   `json:"subscription_price,omitempty"`
	StartDate         time.Time `json:"start_date,omitempty"`
	EndDate           time.Time `json:"end_date,omitempty"`
	Status            string    `json:"status,omitempty"`
	Discount          float64   `json:"discount,omitempty"` // Скидка в рублях
	FinalPrice        float64   `json:"final_price,omitempty"`
}

func (p *PersonSubscription) Validate() map[string]string {
	validate := validator.New()
	err := validate.Struct(p)

	if err == nil {
		return nil
	}

	errs := make(map[string]string)

	for _, err := range err.(validator.ValidationErrors) {
		var msg string

		switch err.Field() {
		case "Number":
			if err.Tag() == "required" {
				msg = "Номер абонемента обязателен для заполнения"
			}
		case "PersonID":
			if err.Tag() == "required" {
				msg = "ID клиента обязателен для заполнения"
			}
		case "SubscriptionID":
			if err.Tag() == "required" {
				msg = "ID абонемента обязателен для заполнения"
			}
		default:
			msg = "Некорректное значение поля" + err.Field()
		}

		errs[err.Field()] = msg
	}

	return errs
}
