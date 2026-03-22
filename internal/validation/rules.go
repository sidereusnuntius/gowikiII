package validation

import (
	"errors"
	"net/mail"
)

var (
	ErrInvalidEmail = errors.New("invalid email")
)

var (
	Username = Rule{Field: "username", Min: 3, Max: 16, Pattern: "(\\d|\\w|_)+"}
	Password = Rule{Field: "password", Min: 8, Max: 72}
	Email    = Rule{Field: "email", Min: 3, Max: 254, CustomFunc: validateEmail}
)

func validateEmail(in string) error {
	m, err := mail.ParseAddress(in)
	if err != nil || in != m.Address {
		return ErrInvalidEmail
	}

	return nil
}
