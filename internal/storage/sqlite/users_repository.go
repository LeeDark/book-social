package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/LeeDark/book-social/internal/modules/users"
	moderncsqlite "modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
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

	txRepo := &UserRepository{db: tx}
	if err := fn(txRepo); err != nil {
		_ = tx.Rollback()
		return mapRepositoryError(err)
	}
	if err := tx.Commit(); err != nil {
		return users.ErrInternal
	}
	return nil
}

func (r *UserRepository) CreateUser(ctx context.Context, params users.CreateUserParams) (users.User, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO users(first_name, login, password_hash, email, user_role_id)
		VALUES (?, ?, ?, ?, ?)
	`, params.FirstName, params.Login, params.PasswordHash, params.Email, params.RoleID)
	if err != nil {
		return users.User{}, r.mapCreateUserError(ctx, err, params)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return users.User{}, users.ErrInternal
	}
	return users.User{
		ID:        int(id),
		FirstName: params.FirstName,
		Login:     params.Login,
		Email:     params.Email,
		RoleID:    params.RoleID,
	}, nil
}

func (r *UserRepository) FindRoleByName(ctx context.Context, name string) (users.Role, error) {
	var role users.Role
	if err := r.db.QueryRowContext(ctx, `
		SELECT id, role_name, is_admin
		FROM roles
		WHERE role_name = ?
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
		WHERE login = ? OR email = ?
		LIMIT 1
	`, identifier, identifier).Scan(
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
		WHERE id = ?
		LIMIT 1
	`, id).Scan(&user.ID, &user.FirstName, &user.Login, &user.Email, &user.RoleID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return users.User{}, users.ErrUserNotFound
		}
		return users.User{}, users.ErrInternal
	}
	return user, nil
}

func (r *UserRepository) mapCreateUserError(ctx context.Context, err error, params users.CreateUserParams) error {
	var sqliteErr *moderncsqlite.Error
	if !errors.As(err, &sqliteErr) || sqliteErr.Code() != sqlite3.SQLITE_CONSTRAINT_UNIQUE {
		return users.ErrInternal
	}

	var exists int
	if queryErr := r.db.QueryRowContext(ctx, `SELECT 1 FROM users WHERE login = ? LIMIT 1`, params.Login).Scan(&exists); queryErr == nil {
		return users.ErrLoginTaken
	}
	if queryErr := r.db.QueryRowContext(ctx, `SELECT 1 FROM users WHERE email = ? LIMIT 1`, params.Email).Scan(&exists); queryErr == nil {
		return users.ErrEmailTaken
	}
	return users.ErrInternal
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

type SessionRepository struct {
	db sqlQueryer
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) CreateSession(ctx context.Context, params users.CreateSessionParams) (users.Session, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions(user_id, token_hash, created_at, expires_at)
		VALUES (?, ?, ?, ?)
	`, params.UserID, params.TokenHash, formatSQLiteTime(params.CreatedAt), formatSQLiteTime(params.ExpiresAt))
	if err != nil {
		return users.Session{}, users.ErrInternal
	}

	id, err := result.LastInsertId()
	if err != nil {
		return users.Session{}, users.ErrInternal
	}
	return users.Session{
		ID:        int(id),
		UserID:    params.UserID,
		TokenHash: append([]byte(nil), params.TokenHash...),
		CreatedAt: params.CreatedAt.UTC(),
		ExpiresAt: params.ExpiresAt.UTC(),
	}, nil
}

func (r *SessionRepository) LoadSession(ctx context.Context, tokenHash []byte, now time.Time) (users.Session, error) {
	var session users.Session
	var createdAt, expiresAt string
	if err := r.db.QueryRowContext(ctx, `
		SELECT s.id, s.user_id, s.token_hash, s.created_at, s.expires_at
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = ? AND s.expires_at > ?
		LIMIT 1
	`, tokenHash, formatSQLiteTime(now)).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&createdAt,
		&expiresAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return users.Session{}, users.ErrUnauthenticated
		}
		return users.Session{}, users.ErrInternal
	}

	var err error
	if session.CreatedAt, err = parseSQLiteTime(createdAt); err != nil {
		return users.Session{}, users.ErrInternal
	}
	if session.ExpiresAt, err = parseSQLiteTime(expiresAt); err != nil {
		return users.Session{}, users.ErrInternal
	}
	return session, nil
}

func (r *SessionRepository) DeleteSession(ctx context.Context, tokenHash []byte) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash = ?`, tokenHash); err != nil {
		return users.ErrInternal
	}
	return nil
}

func formatSQLiteTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func parseSQLiteTime(value string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, value)
}
