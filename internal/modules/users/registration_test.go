package users

import (
	"context"
	"errors"
	"testing"
)

type recordingUserRepository struct {
	role           Role
	created        CreateUserParams
	createdUser    User
	roleLookup     string
	txCalls        int
	roleErr        error
	createErr      error
	credentials    Credentials
	credentialsErr error
}

type recordingPasswordBoundary struct {
	verifyCalls int
	verifyHash  string
	verifyErr   error
}

func (p *recordingPasswordBoundary) Hash(string) (string, error) {
	return "unused-test-hash", nil
}

func (p *recordingPasswordBoundary) Verify(hash, _ string) error {
	p.verifyCalls++
	p.verifyHash = hash
	return p.verifyErr
}

func (r *recordingUserRepository) CreateUser(ctx context.Context, params CreateUserParams) (User, error) {
	r.created = params
	if r.createErr != nil {
		return User{}, r.createErr
	}
	return r.createdUser, nil
}

func (r *recordingUserRepository) FindRoleByName(ctx context.Context, name string) (Role, error) {
	r.roleLookup = name
	if r.roleErr != nil {
		return Role{}, r.roleErr
	}
	return r.role, nil
}

func (r *recordingUserRepository) FindCredentials(ctx context.Context, identifier string) (Credentials, error) {
	if r.credentialsErr != nil {
		return Credentials{}, r.credentialsErr
	}
	return r.credentials, nil
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

func TestRegistrationRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input RegistrationInput
	}{
		{
			name: "missing first name",
			input: RegistrationInput{
				Login:                "ada",
				Email:                "ada@example.test",
				Password:             "correct horse battery staple",
				PasswordConfirmation: "correct horse battery staple",
			},
		},
		{
			name: "password confirmation mismatch",
			input: RegistrationInput{
				FirstName:            "Ada",
				Login:                "ada",
				Email:                "ada@example.test",
				Password:             "correct horse battery staple",
				PasswordConfirmation: "different password",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &recordingUserRepository{role: Role{ID: 9, Name: "user"}}
			_, err := NewService(repo, NewPasswordPolicy()).Register(context.Background(), tt.input)
			var validationErr ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("Register() error = %v, want ValidationError", err)
			}
			if repo.txCalls != 0 {
				t.Fatalf("transaction calls = %d, want 0", repo.txCalls)
			}
		})
	}
}

func TestRegistrationMapsDuplicateIdentityErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "duplicate login", err: ErrLoginTaken},
		{name: "duplicate email", err: ErrEmailTaken},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &recordingUserRepository{
				role:      Role{ID: 9, Name: "user"},
				createErr: tt.err,
			}
			_, err := NewService(repo, NewPasswordPolicy()).Register(context.Background(), validRegistrationInput())
			if !errors.Is(err, tt.err) {
				t.Fatalf("Register() error = %v, want %v", err, tt.err)
			}
		})
	}
}

func TestRegistrationMapsRepositoryFailureToInternalError(t *testing.T) {
	repo := &recordingUserRepository{
		role:      Role{ID: 9, Name: "user"},
		createErr: errors.New("database connection details must not escape"),
	}

	_, err := NewService(repo, NewPasswordPolicy()).Register(context.Background(), validRegistrationInput())
	if !errors.Is(err, ErrInternal) {
		t.Fatalf("Register() error = %v, want ErrInternal", err)
	}
	if err.Error() != ErrInternal.Error() {
		t.Fatalf("Register() error = %q, want generic %q", err, ErrInternal)
	}
}

func TestAuthenticateReturnsMinimalUserForValidCredentials(t *testing.T) {
	password := "correct horse battery staple"
	hash, err := NewPasswordPolicy().Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	wantUser := User{ID: 42, FirstName: "Ada", Login: "ada", Email: "ada@example.test", RoleID: 9}
	repo := &recordingUserRepository{credentials: Credentials{User: wantUser, PasswordHash: hash}}

	got, err := NewService(repo, NewPasswordPolicy()).Authenticate(context.Background(), "ada", password)
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if got != wantUser {
		t.Fatalf("Authenticate() user = %+v, want %+v", got, wantUser)
	}
}

func TestAuthenticateUsesNeutralRefusalForInvalidOrMissingCredentials(t *testing.T) {
	password := "correct horse battery staple"
	hash, err := NewPasswordPolicy().Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	tests := []struct {
		name           string
		credentials    Credentials
		credentialsErr error
		inputPassword  string
	}{
		{
			name:          "wrong password",
			credentials:   Credentials{User: User{ID: 42, Login: "ada"}, PasswordHash: hash},
			inputPassword: "wrong password",
		},
		{
			name:           "missing user",
			credentialsErr: ErrUserNotFound,
			inputPassword:  password,
		},
	}

	var firstErr error
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &recordingUserRepository{credentials: tt.credentials, credentialsErr: tt.credentialsErr}
			got, err := NewService(repo, NewPasswordPolicy()).Authenticate(context.Background(), "ada", tt.inputPassword)
			if got != (User{}) {
				t.Fatalf("Authenticate() user = %+v, want zero User", got)
			}
			if !errors.Is(err, ErrInvalidCredentials) {
				t.Fatalf("Authenticate() error = %v, want ErrInvalidCredentials", err)
			}
			if firstErr == nil {
				firstErr = err
			} else if err.Error() != firstErr.Error() {
				t.Fatalf("Authenticate() refusal = %q, want neutral %q", err, firstErr)
			}
		})
	}
}

func TestAuthenticateVerifiesDummyHashForMissingUser(t *testing.T) {
	repo := &recordingUserRepository{credentialsErr: ErrUserNotFound}
	policy := &recordingPasswordBoundary{verifyErr: ErrInvalidCredentials}

	got, err := NewService(repo, policy).Authenticate(
		context.Background(),
		"missing@example.test",
		"correct horse battery staple",
	)
	if got != (User{}) {
		t.Fatal("Authenticate() returned a user for a missing account")
	}
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Authenticate() error = %v, want ErrInvalidCredentials", err)
	}
	if policy.verifyCalls != 1 {
		t.Fatalf("password verification calls = %d, want 1", policy.verifyCalls)
	}
	if policy.verifyHash == "" {
		t.Fatal("missing account did not use a dummy password hash")
	}
}

func TestAuthenticateMapsInvalidStoredHashToInternalError(t *testing.T) {
	repo := &recordingUserRepository{
		credentials: Credentials{User: User{ID: 42}, PasswordHash: "invalid stored hash"},
	}

	got, err := NewService(repo, NewPasswordPolicy()).Authenticate(
		context.Background(),
		"ada",
		"correct horse battery staple",
	)
	if got != (User{}) {
		t.Fatal("Authenticate() returned a user for an invalid stored hash")
	}
	if !errors.Is(err, ErrInternal) {
		t.Fatalf("Authenticate() error = %v, want ErrInternal", err)
	}
}

func validRegistrationInput() RegistrationInput {
	return RegistrationInput{
		FirstName:            "Ada",
		Login:                "ada",
		Email:                "ada@example.test",
		Password:             "correct horse battery staple",
		PasswordConfirmation: "correct horse battery staple",
	}
}
