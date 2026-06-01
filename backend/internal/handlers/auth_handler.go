package handlers

import (
	"github.com/dirhamt/billing-hotspot/backend/internal/dto"
	"github.com/dirhamt/billing-hotspot/backend/internal/middleware"
	"github.com/dirhamt/billing-hotspot/backend/internal/response"
	"github.com/dirhamt/billing-hotspot/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// AuthHandler exposes authentication endpoints.
type AuthHandler struct {
	svc *services.AuthService
}

// NewAuthHandler builds an AuthHandler.
func NewAuthHandler(svc *services.AuthService) *AuthHandler { return &AuthHandler{svc: svc} }

// Login godoc
// @Summary  Admin login
// @Tags     Auth
// @Accept   json
// @Produce  json
// @Param    payload body dto.LoginRequest true "Credentials"
// @Success  200 {object} response.Envelope{data=dto.LoginResponse}
// @Failure  401 {object} response.Envelope
// @Failure  422 {object} response.Envelope
// @Router   /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !bindJSON(c, &req) {
		return
	}
	res, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Login successful", res)
}

// Me godoc
// @Summary  Current authenticated user
// @Tags     Auth
// @Produce  json
// @Security BearerAuth
// @Success  200 {object} response.Envelope{data=dto.UserResponse}
// @Router   /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	user, err := h.svc.Me(c.Request.Context(), middleware.CurrentUserID(c))
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "OK", dto.ToUserResponse(*user))
}

// ChangePassword godoc
// @Summary  Change current user's password
// @Tags     Auth
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    payload body dto.ChangePasswordRequest true "Passwords"
// @Success  200 {object} response.Envelope
// @Router   /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.ChangePassword(c.Request.Context(), middleware.CurrentUserID(c), req); err != nil {
		fail(c, err)
		return
	}
	response.NoContent(c, "Password updated")
}
