package user

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrEmailExists           = errors.New("email already exists")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrInvalidRole           = errors.New("invalid role")
	ErrInvalidResetToken     = errors.New("invalid or expired reset token")
	ErrPasswordMismatch      = errors.New("current password is incorrect")
	ErrSamePassword          = errors.New("new password must be different from current password")
)

// Service defines user service interface
type Service interface {
	RegisterUser(ctx context.Context, req RegisterRequest) (*User, error)
	AuthenticateUser(ctx context.Context, req LoginRequest) (*User, error)
	GetUserByID(ctx context.Context, id uint) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, id uint, req UpdateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, id uint) error
	ListUsers(ctx context.Context, filters UserFilterParams, page, perPage int) ([]User, int64, error)
	PromoteToAdmin(ctx context.Context, userID uint) error
	ChangePassword(ctx context.Context, userID uint, req ChangePasswordRequest) error
	RequestPasswordReset(ctx context.Context, email string) (*PasswordResetResponse, error)
	ResetPassword(ctx context.Context, req ResetPasswordRequest) error
	AssignUserRole(ctx context.Context, userID uint, roleName string) error
	RemoveUserRole(ctx context.Context, userID uint, roleName string) error
	GetUserRoles(ctx context.Context, userID uint) ([]Role, error)
}

type service struct {
	repo Repository
}

// NewService creates a new user service
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// RegisterUser registers a new user
func (s *service) RegisterUser(ctx context.Context, req RegisterRequest) (*User, error) {
	existingUser, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing email: %w", err)
	}
	if existingUser != nil {
		return nil, ErrEmailExists
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	// Use transaction to ensure atomic user creation and role assignment
	err = s.repo.Transaction(ctx, func(txCtx context.Context) error {
		if err := s.repo.Create(txCtx, user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		if err := s.repo.AssignRole(txCtx, user.ID, RoleUser); err != nil {
			return fmt.Errorf("failed to assign default role: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload user with roles after successful transaction
	user, err = s.repo.FindByID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("failed to reload user: user not found after creation")
	}

	return user, nil
}

// AuthenticateUser authenticates a user with email and password
func (s *service) AuthenticateUser(ctx context.Context, req LoginRequest) (*User, error) {
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := verifyPassword(user.PasswordHash, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *service) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *service) GetUserByID(ctx context.Context, id uint) (*User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateUser updates a user's information
func (s *service) UpdateUser(ctx context.Context, id uint, req UpdateUserRequest) (*User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		existingUser, err := s.repo.FindByEmail(ctx, req.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing email: %w", err)
		}
		if existingUser != nil && existingUser.ID != user.ID {
			return nil, ErrEmailExists
		}
		user.Email = req.Email
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeleteUser deletes a user
func (s *service) DeleteUser(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// ListUsers retrieves paginated list of users with filtering
func (s *service) ListUsers(ctx context.Context, filters UserFilterParams, page, perPage int) ([]User, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		return nil, 0, fmt.Errorf("page must be >= 1")
	}
	if perPage < 1 {
		return nil, 0, fmt.Errorf("perPage must be >= 1")
	}
	if perPage > 100 {
		return nil, 0, fmt.Errorf("perPage must be <= 100")
	}

	if filters.Role != "" && filters.Role != RoleUser && filters.Role != RoleAdmin {
		return nil, 0, ErrInvalidRole
	}

	users, total, err := s.repo.ListAllUsers(ctx, filters, page, perPage)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// PromoteToAdmin promotes a user to admin role
func (s *service) PromoteToAdmin(ctx context.Context, userID uint) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	if user.HasRole(RoleAdmin) {
		return nil
	}

	if err := s.repo.AssignRole(ctx, userID, RoleAdmin); err != nil {
		return fmt.Errorf("failed to assign admin role: %w", err)
	}

	return nil
}

func (s *service) ChangePassword(ctx context.Context, userID uint, req ChangePasswordRequest) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	if err := verifyPassword(user.PasswordHash, req.CurrentPassword); err != nil {
		return ErrPasswordMismatch
	}

	if req.CurrentPassword == req.NewPassword {
		return ErrSamePassword
	}

	hashedPassword, err := hashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = hashedPassword
	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (s *service) RequestPasswordReset(ctx context.Context, email string) (*PasswordResetResponse, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, nil
	}

	rawToken, err := generateRandomToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate reset token: %w", err)
	}

	tokenHash := HashToken(rawToken)

	dbToken := &PasswordResetToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := s.repo.CreatePasswordResetToken(ctx, dbToken); err != nil {
		return nil, fmt.Errorf("failed to store reset token: %w", err)
	}

	return &PasswordResetResponse{
		Token:     rawToken,
		ExpiresIn: 3600,
	}, nil
}

func (s *service) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	tokenHash := HashToken(req.Token)

	storedToken, err := s.repo.FindPasswordResetTokenByHash(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to find reset token: %w", err)
	}
	if storedToken == nil {
		return ErrInvalidResetToken
	}

	if storedToken.UsedAt != nil {
		return ErrInvalidResetToken
	}

	if time.Now().After(storedToken.ExpiresAt) {
		return ErrInvalidResetToken
	}

	hashedPassword, err := hashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	err = s.repo.Transaction(ctx, func(txCtx context.Context) error {
		user, err := s.repo.FindByID(txCtx, storedToken.UserID)
		if err != nil {
			return fmt.Errorf("failed to find user: %w", err)
		}
		if user == nil {
			return ErrUserNotFound
		}

		user.PasswordHash = hashedPassword
		if err := s.repo.Update(txCtx, user); err != nil {
			return fmt.Errorf("failed to update password: %w", err)
		}

		if err := s.repo.MarkPasswordResetTokenUsed(txCtx, storedToken.ID); err != nil {
			return fmt.Errorf("failed to mark token as used: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *service) AssignUserRole(ctx context.Context, userID uint, roleName string) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	role, err := s.repo.FindRoleByName(ctx, roleName)
	if err != nil {
		return fmt.Errorf("failed to find role: %w", err)
	}
	if role == nil {
		return ErrInvalidRole
	}

	if err := s.repo.AssignRole(ctx, userID, roleName); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

func (s *service) RemoveUserRole(ctx context.Context, userID uint, roleName string) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	if err := s.repo.RemoveRole(ctx, userID, roleName); err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	return nil
}

func (s *service) GetUserRoles(ctx context.Context, userID uint) ([]Role, error) {
	roles, err := s.repo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	return roles, nil
}

// hashPassword hashes a plain text password using bcrypt
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// verifyPassword verifies a password against a hash
func verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// generateRandomToken generates a cryptographically secure random token
func generateRandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// HashToken creates a SHA256 hash of the token
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
