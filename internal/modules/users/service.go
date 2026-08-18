package users

import "context"

type RegistrationInput struct {
	FirstName            string
	Login                string
	Email                string
	Password             string
	PasswordConfirmation string
}

type Service struct {
	repo   RegistrationRepository
	policy PasswordPolicy
}

func NewService(repo RegistrationRepository, policy PasswordPolicy) *Service {
	return &Service{repo: repo, policy: policy}
}

func (s *Service) Register(ctx context.Context, input RegistrationInput) (User, error) {
	passwordHash, err := s.policy.Hash(input.Password)
	if err != nil {
		return User{}, err
	}

	var created User
	err = s.repo.WithinTransaction(ctx, func(tx UserRepository) error {
		role, err := tx.FindRoleByName(ctx, "user")
		if err != nil {
			return err
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
		return User{}, err
	}

	return created, nil
}
