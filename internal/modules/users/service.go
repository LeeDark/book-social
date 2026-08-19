package users

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
)

type RegistrationInput struct {
	FirstName            string
	Login                string
	Email                string
	Password             string
	PasswordConfirmation string
}

type Service struct {
	repo   RegistrationRepository
	policy passwordBoundary
}

type passwordBoundary interface {
	Hash(password string) (string, error)
	Verify(hash, password string) error
}

var nonexistentUserHash = func() string {
	hash, err := NewPasswordPolicy().Hash("nonexistent-account-password")
	if err != nil {
		panic("users: generate nonexistent-account password hash")
	}
	return hash
}()

func NewService(repo RegistrationRepository, policy passwordBoundary) *Service {
	return &Service{repo: repo, policy: policy}
}

func (s *Service) Register(ctx context.Context, input RegistrationInput) (User, error) {
	input, err := normalizeRegistrationInput(input)
	if err != nil {
		return User{}, err
	}

	passwordHash, err := s.policy.Hash(input.Password)
	if err != nil {
		if errors.Is(err, ErrPasswordTooShort) || errors.Is(err, ErrPasswordTooLong) {
			return User{}, ValidationError{Field: "password", Message: err.Error()}
		}
		return User{}, ErrInternal
	}

	var created User
	err = s.repo.WithinTransaction(ctx, func(tx UserRepository) error {
		role, err := tx.FindRoleByName(ctx, "user")
		if err != nil {
			return err
		}
		if role.ID == 0 || role.Name != "user" {
			return ErrInternal
		}

		created, err = tx.CreateUser(ctx, CreateUserParams{
			FirstName:    input.FirstName,
			Login:        input.Login,
			Email:        input.Email,
			PasswordHash: passwordHash,
			RoleID:       role.ID,
		})
		return err
	})
	if err != nil {
		return User{}, mapRegistrationError(err)
	}

	return created, nil
}

func (s *Service) Authenticate(ctx context.Context, identifier, password string) (User, error) {
	identifier = strings.ToLower(strings.TrimSpace(identifier))
	if identifier == "" || password == "" {
		return User{}, ErrInvalidCredentials
	}

	credentials, err := s.repo.FindCredentials(ctx, identifier)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) || errors.Is(err, ErrInvalidCredentials) {
			// Keep the expensive verification path for an unknown account so the
			// lookup result is not exposed through an obvious timing difference.
			_ = s.policy.Verify(nonexistentUserHash, password)
			return User{}, ErrInvalidCredentials
		}
		return User{}, ErrInternal
	}
	if err := s.policy.Verify(credentials.PasswordHash, password); err != nil {
		if errors.Is(err, ErrInternal) {
			return User{}, ErrInternal
		}
		return User{}, ErrInvalidCredentials
	}
	return credentials.User, nil
}

func normalizeRegistrationInput(input RegistrationInput) (RegistrationInput, error) {
	input.FirstName = strings.TrimSpace(input.FirstName)
	input.Login = strings.ToLower(strings.TrimSpace(input.Login))
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	validations := []struct {
		field   string
		value   string
		maximum int
	}{
		{field: "first_name", value: input.FirstName, maximum: 100},
		{field: "login", value: input.Login, maximum: 64},
		{field: "email", value: input.Email, maximum: 254},
	}
	for _, validation := range validations {
		if validation.value == "" {
			return RegistrationInput{}, ValidationError{Field: validation.field, Message: "is required"}
		}
		if utf8.RuneCountInString(validation.value) > validation.maximum {
			return RegistrationInput{}, ValidationError{Field: validation.field, Message: "is too long"}
		}
	}
	if !strings.Contains(input.Email, "@") {
		return RegistrationInput{}, ValidationError{Field: "email", Message: "is invalid"}
	}
	if input.Password != input.PasswordConfirmation {
		return RegistrationInput{}, ValidationError{Field: "password_confirmation", Message: "does not match password"}
	}

	return input, nil
}

func mapRegistrationError(err error) error {
	switch {
	case errors.Is(err, ErrLoginTaken):
		return ErrLoginTaken
	case errors.Is(err, ErrEmailTaken):
		return ErrEmailTaken
	case errors.Is(err, ErrInternal):
		return ErrInternal
	default:
		return ErrInternal
	}
}
