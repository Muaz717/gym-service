package dto

import (
	"github.com/go-playground/validator/v10"
	"time"
)

type PersonSubResponse struct {
	Number            string    `json:"number"`
	PersonID          int       `json:"person_id"`
	PersonName        string    `json:"person_name"`
	SubscriptionID    int       `json:"subscription_id"`
	SubscriptionTitle string    `json:"subscription_title,omitempty"`
	SubscriptionPrice float64   `json:"subscription_price,omitempty"`
	StartDate         time.Time `json:"start_date,omitempty"`
	EndDate           time.Time `json:"end_date,omitempty"`
	Status            string    `json:"status,omitempty"`
	Discount          float64   `json:"discount,omitempty"`
	FinalPrice        float64   `json:"final_price,omitempty"`
	FreezeDays        int       `json:"freeze_days"`      // <--- добавить!
	UsedFreezeDays    int       `json:"used_freeze_days"` // <--- добавить!
}

// Промежуточная структура со строками для дат
type PersonSubInput struct {
	Number            string  `json:"number" validate:"required"`
	PersonID          int     `json:"person_id" validate:"required"`
	SubscriptionID    int     `json:"subscription_id" validate:"required"`
	SubscriptionPrice float64 `json:"subscription_price,omitempty"`
	StartDate         string  `json:"start_date,omitempty"` // строка
	EndDate           string  `json:"end_date,omitempty"`   // строка
	Status            string  `json:"status,omitempty"`
	Discount          float64 `json:"discount,omitempty"`
	FinalPrice        float64 `json:"final_price,omitempty"`
}

func (p *PersonSubInput) Validate() map[string]string {
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
