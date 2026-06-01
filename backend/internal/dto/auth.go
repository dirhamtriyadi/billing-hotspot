package dto

import (
	"time"

	"github.com/dirhamt/billing-hotspot/backend/internal/models"
)

// LoginRequest authenticates an admin/operator.
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"admin123"`
}

// UserResponse is the safe public projection of a user.
type UserResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}

// LoginResponse carries the issued JWT and the authenticated profile.
type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      UserResponse `json:"user"`
}

// ChangePasswordRequest updates the current user's password.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=72"`
}

// ToUserResponse maps a model to its public projection.
func ToUserResponse(u models.User) UserResponse {
	return UserResponse{
		ID:       u.ID,
		Name:     u.Name,
		Username: u.Username,
		Email:    u.Email,
		Role:     u.Role,
		IsActive: u.IsActive,
	}
}
