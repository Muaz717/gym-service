package models

import (
	"github.com/go-playground/validator/v10"
)

type Person struct {
	Id    int    `json:"id,omitempty"`
	Name  string `json:"name,omitempty" db:"full_name" validate:"required,min=2,max=50"`
	Phone string `json:"phone,omitempty" validate:"required,len=11,number"`
	//Memberships []Subscription `json:"memberships,omitempty" required:"false"`
}

func (p *Person) Validate() map[string]string {
	validate := validator.New()

	err := validate.Struct(p)
	if err == nil {
		return nil
	}

	errs := make(map[string]string)

	for _, err := range err.(validator.ValidationErrors) {
		var msg string

		switch err.Field() {
		case "Name":
			if err.Tag() == "required" {
				msg = "ФИО обязательно для заполнения"
			} else if err.Tag() == "min" {
				msg = "ФИО должно содержать не менее 2 символов"
			} else if err.Tag() == "max" {
				msg = "ФИО должно содержать не более 50 символов"
			}
		case "Phone":
			if err.Tag() == "required" {
				msg = "Телефон обязателен для заполнения"
			} else if err.Tag() == "len" {
				msg = "Телефон должен содержать 11 цифр"
			} else if err.Tag() == "number" {
				msg = "Телефон должен содержать только цифры"
			}
		default:
			msg = "Некорректное значение поля" + err.Field()
		}

		errs[err.Field()] = msg
	}

	return errs
}
