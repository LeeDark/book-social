package users

import (
	"context"
	"testing"
)

type recordingUserRepository struct {
	role        Role
	created     CreateUserParams
	createdUser User
	roleLookup  string
	txCalls     int
}

func (r *recordingUserRepository) CreateUser(ctx context.Context, params CreateUserParams) (User, error) {
	r.created = params
	return r.createdUser, nil
}

func (r *recordingUserRepository) FindRoleByName(ctx context.Context, name string) (Role, error) {
	r.roleLookup = name
	return r.role, nil
}

func (r *recordingUserRepository) FindCredentials(ctx context.Context, identifier string) (Credentials, error) {
	return Credentials{}, nil
}

func (r *recordingUserRepository) FindByID(ctx context.Context, id int) (User, error) {
	return User{}, nil
}

func (r *recordingUserRepository) WithinTransaction(ctx context.Context, fn func(UserRepository) error) error {
	r.txCalls++
	return fn(r)
}

func TestRegistrationCreatesUserWithDefaultRoleAndPasswordHash(t *testing.T) {
	ctx := context.Background()
	repo := &recordingUserRepository{
		role:        Role{ID: 9, Name: "user"},
		createdUser: User{ID: 42, FirstName: "Ada", Login: "ada", Email: "ada@example.test", RoleID: 9},
	}
	service := NewService(repo, NewPasswordPolicy())
	input := RegistrationInput{
		FirstName:            "Ada",
		Login:                "ada",
		Email:                "ada@example.test",
		Password:             "correct horse battery staple",
		PasswordConfirmation: "correct horse battery staple",
	}

	got, err := service.Register(ctx, input)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if repo.txCalls != 1 {
		t.Fatalf("transaction calls = %d, want 1", repo.txCalls)
	}
	if got != repo.createdUser {
		t.Fatalf("Register() user = %+v, want %+v", got, repo.createdUser)
	}
	if repo.roleLookup != "user" {
		t.Fatalf("role lookup = %q, want %q", repo.roleLookup, "user")
	}
	if repo.created.RoleID != repo.role.ID {
		t.Fatalf("created role ID = %d, want %d", repo.created.RoleID, repo.role.ID)
	}
	if repo.created.PasswordHash == "" {
		t.Fatal("created password hash is empty")
	}
	if repo.created.PasswordHash == input.Password {
		t.Fatal("created password hash contains the raw password")
	}
	if err := NewPasswordPolicy().Verify(repo.created.PasswordHash, input.Password); err != nil {
		t.Fatalf("created password hash does not verify: %v", err)
	}
}
