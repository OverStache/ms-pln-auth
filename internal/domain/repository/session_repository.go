package repository

import (
	"context"

	"ms-pln-auth/internal/domain/entity"
)

// SessionRepository defines the interface for session data operations
type SessionRepository interface {
	// Create creates a new session
	Create(ctx context.Context, session *entity.Session) error

	// GetByToken retrieves a session by token
	GetByToken(ctx context.Context, token string) (*entity.Session, error)

	// GetByUserID retrieves all sessions for a user
	GetByUserID(ctx context.Context, userID int64) ([]*entity.Session, error)

	// RevokeByToken revokes a session by token
	RevokeByToken(ctx context.Context, token string) error

	// RevokeAllByUserID revokes all sessions for a user
	RevokeAllByUserID(ctx context.Context, userID int64) error

	// DeleteExpired deletes all expired sessions
	DeleteExpired(ctx context.Context) error
}
