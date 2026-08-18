package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/LeeDark/book-social/internal/modules/users"
	"github.com/lib/pq"
)

type sqlQueryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type UserRepository struct {
	db sqlQueryer
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) WithinTransaction(ctx context.Context, fn func(users.UserRepository) error) error {
	db, ok := r.db.(*sql.DB)
	if !ok {
		return users.ErrInternal
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return users.ErrInternal
	}
	if err := fn(&UserRepository{db: tx}); err != nil {
		_ = tx.Rollback()
		return mapRepositoryError(err)
	}
	if err := tx.Commit(); err != nil {
		return users.ErrInternal
	}
	return nil
}

func (r *UserRepository) CreateUser(ctx context.Context, params users.CreateUserParams) (users.User, error) {
	var user users.User
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users(first_name, login, password_hash, email, user_role_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, first_name, login, email, user_role_id
	`, params.FirstName, params.Login, params.PasswordHash, params.Email, params.RoleID).Scan(
		&user.ID, &user.FirstName, &user.Login, &user.Email, &user.RoleID,
	)
	if err != nil {
		return users.User{}, mapCreateUserError(err)
	}
	return user, nil
}

func (r *UserRepository) FindRoleByName(ctx context.Context, name string) (users.Role, error) {
	var role users.Role
	if err := r.db.QueryRowContext(ctx, `
		SELECT id, role_name, is_admin
		FROM roles
		WHERE role_name = $1
		LIMIT 1
	`, name).Scan(&role.ID, &role.Name, &role.IsAdmin); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return users.Role{}, users.ErrRoleNotFound
		}
		return users.Role{}, users.ErrInternal
	}
	return role, nil
}

func (r *UserRepository) FindCredentials(ctx context.Context, identifier string) (users.Credentials, error) {
	var credentials users.Credentials
	if err := r.db.QueryRowContext(ctx, `
		SELECT id, first_name, login, email, user_role_id, password_hash
		FROM users
		WHERE login = $1 OR email = $1
		LIMIT 1
	`, identifier).Scan(
		&credentials.User.ID,
		&credentials.User.FirstName,
		&credentials.User.Login,
		&credentials.User.Email,
		&credentials.User.RoleID,
		&credentials.PasswordHash,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return users.Credentials{}, users.ErrUserNotFound
		}
		return users.Credentials{}, users.ErrInternal
	}
	return credentials, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int) (users.User, error) {
	var user users.User
	if err := r.db.QueryRowContext(ctx, `
		SELECT id, first_name, login, email, user_role_id
		FROM users
		WHERE id = $1
		LIMIT 1
	`, id).Scan(&user.ID, &user.FirstName, &user.Login, &user.Email, &user.RoleID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return users.User{}, users.ErrUserNotFound
		}
		return users.User{}, users.ErrInternal
	}
	return user, nil
}

type SessionRepository struct {
	db sqlQueryer
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) CreateSession(ctx context.Context, params users.CreateSessionParams) (users.Session, error) {
	var session users.Session
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO sessions(user_id, token_hash, created_at, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, token_hash, created_at, expires_at
	`, params.UserID, params.TokenHash, params.CreatedAt.UTC(), params.ExpiresAt.UTC()).Scan(
		&session.ID, &session.UserID, &session.TokenHash, &session.CreatedAt, &session.ExpiresAt,
	)
	if err != nil {
		return users.Session{}, users.ErrInternal
	}
	return session, nil
}

func (r *SessionRepository) LoadSession(ctx context.Context, tokenHash []byte, now time.Time) (users.Session, error) {
	var session users.Session
	if err := r.db.QueryRowContext(ctx, `
		SELECT s.id, s.user_id, s.token_hash, s.created_at, s.expires_at
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.expires_at > $2
		LIMIT 1
	`, tokenHash, now.UTC()).Scan(
		&session.ID, &session.UserID, &session.TokenHash, &session.CreatedAt, &session.ExpiresAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return users.Session{}, users.ErrUnauthenticated
		}
		return users.Session{}, users.ErrInternal
	}
	return session, nil
}

func (r *SessionRepository) DeleteSession(ctx context.Context, tokenHash []byte) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash = $1`, tokenHash); err != nil {
		return users.ErrInternal
	}
	return nil
}

func mapCreateUserError(err error) error {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) || pqErr.Code != "23505" {
		return users.ErrInternal
	}
	switch pqErr.Constraint {
	case "uq_users_login":
		return users.ErrLoginTaken
	case "uq_users_email":
		return users.ErrEmailTaken
	default:
		return users.ErrInternal
	}
}

func mapRepositoryError(err error) error {
	switch {
	case errors.Is(err, users.ErrLoginTaken):
		return users.ErrLoginTaken
	case errors.Is(err, users.ErrEmailTaken):
		return users.ErrEmailTaken
	case errors.Is(err, users.ErrRoleNotFound), errors.Is(err, users.ErrUserNotFound):
		return err
	case errors.Is(err, users.ErrInternal):
		return users.ErrInternal
	default:
		return fmt.Errorf("repository operation: %w", err)
	}
}
