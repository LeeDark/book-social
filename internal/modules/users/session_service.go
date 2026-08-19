package users

import (
	"context"
	"errors"
	"time"
)

type SessionService struct {
	users    UserRepository
	sessions SessionRepository
	lifetime time.Duration
	now      func() time.Time
}

func NewSessionService(users UserRepository, sessions SessionRepository, lifetime time.Duration) *SessionService {
	return &SessionService{
		users:    users,
		sessions: sessions,
		lifetime: lifetime,
		now:      time.Now,
	}
}

func (s *SessionService) CreateSession(ctx context.Context, userID int, tokenHash []byte) error {
	if s == nil || s.sessions == nil || s.lifetime <= 0 || userID <= 0 || len(tokenHash) != SessionTokenHashSize {
		return ErrInternal
	}

	createdAt := s.now().UTC()
	params := CreateSessionParams{
		UserID:    userID,
		TokenHash: append([]byte(nil), tokenHash...),
		CreatedAt: createdAt,
		ExpiresAt: createdAt.Add(s.lifetime),
	}
	if _, err := s.sessions.CreateSession(ctx, params); err != nil {
		return mapSessionServiceError(err)
	}
	return nil
}

func (s *SessionService) LoadCurrentUser(ctx context.Context, tokenHash []byte, now time.Time) (User, error) {
	session, err := s.sessions.LoadSession(ctx, tokenHash, now)
	if err != nil {
		return User{}, mapSessionLoadError(err)
	}

	user, err := s.users.FindByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) || errors.Is(err, ErrUnauthenticated) {
			return User{}, ErrUnauthenticated
		}
		return User{}, ErrInternal
	}
	return user, nil
}

func (s *SessionService) DeleteSession(ctx context.Context, tokenHash []byte) error {
	if err := s.sessions.DeleteSession(ctx, tokenHash); err != nil {
		return mapSessionServiceError(err)
	}
	return nil
}

func mapSessionLoadError(err error) error {
	if errors.Is(err, ErrUnauthenticated) {
		return ErrUnauthenticated
	}
	return ErrInternal
}

func mapSessionServiceError(err error) error {
	if errors.Is(err, ErrUnauthenticated) {
		return ErrUnauthenticated
	}
	return ErrInternal
}
