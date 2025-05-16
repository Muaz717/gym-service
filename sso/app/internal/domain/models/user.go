package models

import (
	"github.com/go-playground/validator/v10"
)

type User struct {
	ID       int64
	Email    string `validate:"required,email"`
	PassHash []byte `validate:"required"`
}

func (u *User) Validate() map[string]string {
	validate := validator.New()

	err := validate.Struct(u)
	if err == nil {
		return nil
	}

	errs := make(map[string]string)

	for _, err := range err.(validator.ValidationErrors) {
		var msg string

		switch err.Field() {
		case "Email":
			if err.Tag() == "required" {
				msg = "Email is required"
			} else if err.Tag() == "email" {
				msg = "Invalid email format"
			}
		case "PassHash":
			if err.Tag() == "required" {
				msg = "Password hash is required"
			}
		default:
			msg = "Invalid value for field: " + err.Field()
		}

		errs[err.Field()] = msg
	}

	return errs
}
