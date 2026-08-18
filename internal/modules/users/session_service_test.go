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
	tokenHash := []byte("hashed-session-token")
	wantUser := User{ID: 42, FirstName: "Ada", Login: "ada", Email: "ada@example.test", RoleID: 9}
	sessionRepo := &recordingSessionRepository{loaded: Session{ID: 7, UserID: wantUser.ID, TokenHash: tokenHash}}
	userRepo := &sessionUserRepository{user: wantUser}
	service := NewSessionService(userRepo, sessionRepo)

	params := CreateSessionParams{
		UserID:    wantUser.ID,
		TokenHash: tokenHash,
		CreatedAt: now,
		ExpiresAt: now.Add(7 * 24 * time.Hour),
	}
	if err := service.CreateSession(ctx, params); err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if sessionRepo.created.UserID != params.UserID || !bytes.Equal(sessionRepo.created.TokenHash, tokenHash) {
		t.Fatalf("CreateSession() params = %+v, want user and token hash", sessionRepo.created)
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
		t.Fatalf("DeleteSession() token hash = %x, want %x", sessionRepo.deletedHash, tokenHash)
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
			service := NewSessionService(&sessionUserRepository{}, sessionRepo)

			got, err := service.LoadCurrentUser(context.Background(), []byte("hashed-session-token"), time.Now())
			if got != (User{}) {
				t.Fatalf("LoadCurrentUser() user = %+v, want zero User", got)
			}
			if !errors.Is(err, ErrUnauthenticated) {
				t.Fatalf("LoadCurrentUser() error = %v, want ErrUnauthenticated", err)
			}
		})
	}
}
