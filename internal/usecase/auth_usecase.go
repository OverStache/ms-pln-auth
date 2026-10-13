package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"ms-pln-auth/internal/domain/entity"
	"ms-pln-auth/internal/domain/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

// AuthUsecase defines the interface for authentication operations
type AuthUsecase interface {
	// Login authenticates a user and returns a session token
	Login(ctx context.Context, username, password string) (*LoginResponse, error)

	// Logout invalidates a session token
	Logout(ctx context.Context, token string) error

	// ValidateToken validates a session token and returns the user
	ValidateToken(ctx context.Context, token string) (*entity.User, error)

	// Register creates a new user account
	Register(ctx context.Context, req *RegisterRequest) (*entity.User, error)
}

// LoginResponse represents the response from a successful login
type LoginResponse struct {
	User      *entity.User `json:"user"`
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
}

// RegisterRequest represents the request to register a new user
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// AuthUsecaseImpl implements AuthUsecase
type AuthUsecaseImpl struct {
	userRepo       repository.UserRepository
	sessionRepo    repository.SessionRepository
	tokenDuration  time.Duration
	bcryptCost     int
}

// AuthUsecaseConfig holds configuration for AuthUsecaseImpl
type AuthUsecaseConfig struct {
	TokenDuration time.Duration
	BcryptCost    int
}

// DefaultAuthUsecaseConfig returns default configuration
func DefaultAuthUsecaseConfig() AuthUsecaseConfig {
	return AuthUsecaseConfig{
		TokenDuration: 24 * time.Hour,
		BcryptCost:    bcrypt.DefaultCost,
	}
}

// NewAuthUsecase creates a new AuthUsecaseImpl
func NewAuthUsecase(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	config AuthUsecaseConfig,
) AuthUsecase {
	return &AuthUsecaseImpl{
		userRepo:      userRepo,
		sessionRepo:   sessionRepo,
		tokenDuration: config.TokenDuration,
		bcryptCost:    config.BcryptCost,
	}
}

// Login authenticates a user and returns a session token
func (u *AuthUsecaseImpl) Login(ctx context.Context, username, password string) (*LoginResponse, error) {
	// Get user by username
	user, err := u.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate session token
	token, err := generateToken(32)
	if err != nil {
		return nil, err
	}

	// Create session
	expiresAt := time.Now().Add(u.tokenDuration)
	session := &entity.Session{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	if err := u.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return &LoginResponse{
		User:      user,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// Logout invalidates a session token
func (u *AuthUsecaseImpl) Logout(ctx context.Context, token string) error {
	// Get session to verify it exists
	session, err := u.sessionRepo.GetByToken(ctx, token)
	if err != nil {
		return err
	}
	if session == nil {
		return ErrInvalidToken
	}

	// Revoke the session
	return u.sessionRepo.RevokeByToken(ctx, token)
}

// ValidateToken validates a session token and returns the user
func (u *AuthUsecaseImpl) ValidateToken(ctx context.Context, token string) (*entity.User, error) {
	// Get session
	session, err := u.sessionRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrInvalidToken
	}

	// Check if session is valid
	if !session.IsValid() {
		return nil, ErrInvalidToken
	}

	// Get user
	user, err := u.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// Register creates a new user account
func (u *AuthUsecaseImpl) Register(ctx context.Context, req *RegisterRequest) (*entity.User, error) {
	// Check if username already exists
	existingUser, err := u.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Check if email already exists
	existingUser, err = u.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), u.bcryptCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &entity.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// generateToken generates a random token of the specified length in bytes
func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
