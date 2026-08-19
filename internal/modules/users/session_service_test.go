package users

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"
)

type recordingSessionRepository struct {
	created     CreateSessionParams
	createdErr  error
	loaded      Session
	loadErr     error
	deletedHash []byte
	deleteErr   error
}

func (r *recordingSessionRepository) CreateSession(ctx context.Context, params CreateSessionParams) (Session, error) {
	r.created = params
	if r.createdErr != nil {
		return Session{}, r.createdErr
	}
	return r.loaded, nil
}

func (r *recordingSessionRepository) LoadSession(ctx context.Context, tokenHash []byte, now time.Time) (Session, error) {
	if r.loadErr != nil {
		return Session{}, r.loadErr
	}
	return r.loaded, nil
}

func (r *recordingSessionRepository) DeleteSession(ctx context.Context, tokenHash []byte) error {
	r.deletedHash = append([]byte(nil), tokenHash...)
	return r.deleteErr
}

type sessionUserRepository struct {
	user    User
	userErr error
}

func (r *sessionUserRepository) CreateUser(ctx context.Context, params CreateUserParams) (User, error) {
	return User{}, nil
}

func (r *sessionUserRepository) FindRoleByName(ctx context.Context, name string) (Role, error) {
	return Role{}, nil
}

func (r *sessionUserRepository) FindCredentials(ctx context.Context, identifier string) (Credentials, error) {
	return Credentials{}, nil
}

func (r *sessionUserRepository) FindByID(ctx context.Context, id int) (User, error) {
	if r.userErr != nil {
		return User{}, r.userErr
	}
	return r.user, nil
}

func TestSessionServiceLifecycleReturnsOnlyCurrentUser(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	lifetime := 7 * 24 * time.Hour
	tokenHash := bytes.Repeat([]byte{0x42}, SessionTokenHashSize)
	wantUser := User{ID: 42, FirstName: "Ada", Login: "ada", Email: "ada@example.test", RoleID: 9}
	sessionRepo := &recordingSessionRepository{loaded: Session{ID: 7, UserID: wantUser.ID, TokenHash: tokenHash}}
	userRepo := &sessionUserRepository{user: wantUser}
	service := NewSessionService(userRepo, sessionRepo, lifetime)
	service.now = func() time.Time { return now }

	if err := service.CreateSession(ctx, wantUser.ID, tokenHash); err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if sessionRepo.created.UserID != wantUser.ID {
		t.Fatalf("CreateSession() user ID = %d, want %d", sessionRepo.created.UserID, wantUser.ID)
	}
	if !bytes.Equal(sessionRepo.created.TokenHash, tokenHash) {
		t.Fatal("CreateSession() persisted an unexpected token hash")
	}
	if !sessionRepo.created.CreatedAt.Equal(now) {
		t.Fatalf("CreateSession() created at = %s, want %s", sessionRepo.created.CreatedAt, now)
	}
	if !sessionRepo.created.ExpiresAt.Equal(now.Add(lifetime)) {
		t.Fatalf("CreateSession() expires at = %s, want seven-day lifetime", sessionRepo.created.ExpiresAt)
	}

	got, err := service.LoadCurrentUser(ctx, tokenHash, now)
	if err != nil {
		t.Fatalf("LoadCurrentUser() error = %v", err)
	}
	if got != wantUser {
		t.Fatalf("LoadCurrentUser() user = %+v, want %+v", got, wantUser)
	}

	if err := service.DeleteSession(ctx, tokenHash); err != nil {
		t.Fatalf("DeleteSession() error = %v", err)
	}
	if !bytes.Equal(sessionRepo.deletedHash, tokenHash) {
		t.Fatal("DeleteSession() received an unexpected token hash")
	}
}

func TestSessionServiceRejectsInvalidCreationPolicy(t *testing.T) {
	validHash := bytes.Repeat([]byte{0x42}, SessionTokenHashSize)
	tests := []struct {
		name      string
		lifetime  time.Duration
		userID    int
		tokenHash []byte
	}{
		{name: "missing lifetime", userID: 42, tokenHash: validHash},
		{name: "missing user", lifetime: 7 * 24 * time.Hour, tokenHash: validHash},
		{name: "invalid token hash", lifetime: 7 * 24 * time.Hour, userID: 42, tokenHash: []byte("not-a-sha256-hash")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &recordingSessionRepository{}
			service := NewSessionService(&sessionUserRepository{}, repo, tt.lifetime)

			err := service.CreateSession(context.Background(), tt.userID, tt.tokenHash)
			if !errors.Is(err, ErrInternal) {
				t.Fatalf("CreateSession() error = %v, want ErrInternal", err)
			}
			if repo.created.UserID != 0 {
				t.Fatal("CreateSession() called repository with an invalid creation policy")
			}
		})
	}
}

func TestSessionServiceRejectsExpiredOrMissingSession(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "expired", err: ErrUnauthenticated},
		{name: "missing", err: ErrUnauthenticated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionRepo := &recordingSessionRepository{loadErr: tt.err}
			service := NewSessionService(&sessionUserRepository{}, sessionRepo, 7*24*time.Hour)

			got, err := service.LoadCurrentUser(
				context.Background(),
				bytes.Repeat([]byte{0x42}, SessionTokenHashSize),
				time.Now(),
			)
			if got != (User{}) {
				t.Fatalf("LoadCurrentUser() user = %+v, want zero User", got)
			}
			if !errors.Is(err, ErrUnauthenticated) {
				t.Fatalf("LoadCurrentUser() error = %v, want ErrUnauthenticated", err)
			}
		})
	}
}
