package users

import (
	"errors"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const (
	PasswordMinLength = 12
	PasswordMaxLength = 128
	passwordHashCost  = bcrypt.DefaultCost
)

var (
	ErrPasswordTooShort = errors.New("password is too short")
	ErrPasswordTooLong  = errors.New("password is too long")
)

type PasswordPolicy struct {
	cost int
}

func NewPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{cost: passwordHashCost}
}

func (p PasswordPolicy) Hash(password string) (string, error) {
	if err := validatePasswordLength(password); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), p.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (p PasswordPolicy) Verify(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func validatePasswordLength(password string) error {
	length := utf8.RuneCountInString(password)
	if length < PasswordMinLength {
		return ErrPasswordTooShort
	}
	if length > PasswordMaxLength {
		return ErrPasswordTooLong
	}
	return nil
}
