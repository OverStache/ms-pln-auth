package postgres

import (
	"context"
	"database/sql"
	"errors"

	"ms-pln-auth/internal/domain/entity"
	"ms-pln-auth/internal/domain/repository"
)

// SessionRepositoryImpl implements repository.SessionRepository using PostgreSQL
type SessionRepositoryImpl struct {
	db *sql.DB
}

// NewSessionRepository creates a new SessionRepositoryImpl
func NewSessionRepository(db *sql.DB) repository.SessionRepository {
	return &SessionRepositoryImpl{db: db}
}

// Create creates a new session
func (r *SessionRepositoryImpl) Create(ctx context.Context, session *entity.Session) error {
	query := `
		INSERT INTO sessions (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, is_revoked
	`

	err := r.db.QueryRowContext(ctx, query, session.UserID, session.Token, session.ExpiresAt).
		Scan(&session.ID, &session.CreatedAt, &session.IsRevoked)
	if err != nil {
		return err
	}

	return nil
}

// GetByToken retrieves a session by token
func (r *SessionRepositoryImpl) GetByToken(ctx context.Context, token string) (*entity.Session, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at, is_revoked
		FROM sessions
		WHERE token = $1
	`

	session := &entity.Session{}
	err := r.db.QueryRowContext(ctx, query, token).
		Scan(&session.ID, &session.UserID, &session.Token, &session.ExpiresAt, &session.CreatedAt, &session.IsRevoked)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return session, nil
}

// GetByUserID retrieves all sessions for a user
func (r *SessionRepositoryImpl) GetByUserID(ctx context.Context, userID int64) ([]*entity.Session, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at, is_revoked
		FROM sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*entity.Session
	for rows.Next() {
		session := &entity.Session{}
		err := rows.Scan(&session.ID, &session.UserID, &session.Token, &session.ExpiresAt, &session.CreatedAt, &session.IsRevoked)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

// RevokeByToken revokes a session by token
func (r *SessionRepositoryImpl) RevokeByToken(ctx context.Context, token string) error {
	query := `UPDATE sessions SET is_revoked = TRUE WHERE token = $1`

	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("session not found")
	}

	return nil
}

// RevokeAllByUserID revokes all sessions for a user
func (r *SessionRepositoryImpl) RevokeAllByUserID(ctx context.Context, userID int64) error {
	query := `UPDATE sessions SET is_revoked = TRUE WHERE user_id = $1 AND is_revoked = FALSE`

	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// DeleteExpired deletes all expired sessions
func (r *SessionRepositoryImpl) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < NOW() OR is_revoked = TRUE`

	_, err := r.db.ExecContext(ctx, query)
	return err
}
