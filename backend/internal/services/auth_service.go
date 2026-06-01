package services

import (
	"context"
	"errors"

	"github.com/dirhamt/billing-hotspot/backend/internal/apperror"
	"github.com/dirhamt/billing-hotspot/backend/internal/config"
	"github.com/dirhamt/billing-hotspot/backend/internal/dto"
	"github.com/dirhamt/billing-hotspot/backend/internal/models"
	"github.com/dirhamt/billing-hotspot/backend/internal/token"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService handles admin authentication.
type AuthService struct {
	db  *gorm.DB
	jwt config.JWTConfig
}

// NewAuthService builds an AuthService.
func NewAuthService(db *gorm.DB, jwt config.JWTConfig) *AuthService {
	return &AuthService{db: db, jwt: jwt}
}

// Login verifies credentials and issues a JWT.
func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	var user models.User
	err := s.db.WithContext(ctx).Where("username = ?", req.Username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.Unauthorized("Invalid username or password")
		}
		return nil, mapDBError(err)
	}

	if !user.IsActive {
		return nil, apperror.Forbidden("This account has been disabled")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		return nil, apperror.Unauthorized("Invalid username or password")
	}

	signed, expiresAt, err := token.Generate(s.jwt.Secret, s.jwt.Issuer, s.jwt.AccessTTL, user.ID, user.Username, user.Role)
	if err != nil {
		return nil, apperror.Internal("Failed to issue token").WithCause(err)
	}

	return &dto.LoginResponse{
		Token:     signed,
		ExpiresAt: expiresAt,
		User:      dto.ToUserResponse(user),
	}, nil
}

// Me returns the authenticated user's profile.
func (s *AuthService) Me(ctx context.Context, userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, mapDBError(err)
	}
	return &user, nil
}

// ChangePassword updates the current user's password after verifying the old one.
func (s *AuthService) ChangePassword(ctx context.Context, userID uint, req dto.ChangePasswordRequest) error {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return mapDBError(err)
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)) != nil {
		return apperror.BadRequest("Current password is incorrect")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return apperror.Internal("Failed to hash password").WithCause(err)
	}
	if err := s.db.WithContext(ctx).Model(&user).Update("password", string(hash)).Error; err != nil {
		return mapDBError(err)
	}
	return nil
}
