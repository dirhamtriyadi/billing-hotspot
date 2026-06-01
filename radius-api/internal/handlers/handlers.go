// Package handlers contains the radius-api HTTP handlers.
package handlers

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/dirhamt/billing-hotspot/radius-api/internal/dto"
	"github.com/dirhamt/billing-hotspot/radius-api/internal/radiusreload"
	"github.com/dirhamt/billing-hotspot/radius-api/internal/radiussql"
	"github.com/dirhamt/billing-hotspot/radius-api/internal/response"
	"github.com/gin-gonic/gin"
)

// Handler bundles the radius service for all endpoints.
type Handler struct {
	svc      *radiussql.Service
	reloader *radiusreload.Reloader
}

// New builds a Handler.
func New(svc *radiussql.Service, reloader *radiusreload.Reloader) *Handler {
	return &Handler{svc: svc, reloader: reloader}
}

func bind(c *gin.Context, target interface{}) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		response.Error(c, 422, "VALIDATION_ERROR", err.Error())
		return false
	}
	return true
}

func fail(c *gin.Context, err error) {
	if errors.Is(err, radiussql.ErrNotFound) {
		response.Error(c, 404, "NOT_FOUND", "Resource not found")
		return
	}
	slog.Error("radius-api error", slog.Any("error", err))
	response.Error(c, 500, "INTERNAL_ERROR", err.Error())
}

// UpsertProfile godoc
// @Summary  Create or update a profile (FreeRADIUS group)
// @Tags     Profiles
// @Accept   json
// @Produce  json
// @Security ApiKeyAuth
// @Param    payload body dto.ProfileRequest true "Profile"
// @Success  200 {object} response.Envelope
// @Router   /profiles [post]
func (h *Handler) UpsertProfile(c *gin.Context) {
	var req dto.ProfileRequest
	if !bind(c, &req) {
		return
	}
	if err := h.svc.UpsertProfile(c.Request.Context(), req); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Profile saved", gin.H{"profile": req.Name})
}

// CreateUser godoc
// @Summary  Provision a single credential
// @Tags     Users
// @Accept   json
// @Produce  json
// @Security ApiKeyAuth
// @Param    payload body dto.UserRequest true "User"
// @Success  201 {object} response.Envelope
// @Router   /users [post]
func (h *Handler) CreateUser(c *gin.Context) {
	var req dto.UserRequest
	if !bind(c, &req) {
		return
	}
	if err := h.svc.CreateUser(c.Request.Context(), req); err != nil {
		fail(c, err)
		return
	}
	response.Created(c, "User provisioned", gin.H{"username": req.Username})
}

// CreateUsers godoc
// @Summary  Bulk-provision credentials
// @Tags     Users
// @Accept   json
// @Produce  json
// @Security ApiKeyAuth
// @Param    payload body dto.BulkUsersRequest true "Users"
// @Success  201 {object} response.Envelope
// @Router   /users/bulk [post]
func (h *Handler) CreateUsers(c *gin.Context) {
	var req dto.BulkUsersRequest
	if !bind(c, &req) {
		return
	}
	if err := h.svc.CreateUsers(c.Request.Context(), req.Users); err != nil {
		fail(c, err)
		return
	}
	response.Created(c, "Users provisioned", gin.H{"count": len(req.Users)})
}

// GetUser godoc
// @Summary  Get a credential's attributes
// @Tags     Users
// @Produce  json
// @Security ApiKeyAuth
// @Param    username path string true "Username"
// @Success  200 {object} response.Envelope{data=dto.UserDetail}
// @Router   /users/{username} [get]
func (h *Handler) GetUser(c *gin.Context) {
	detail, err := h.svc.GetUser(c.Request.Context(), c.Param("username"))
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "OK", detail)
}

// DeleteUser godoc
// @Summary  Remove a credential
// @Tags     Users
// @Produce  json
// @Security ApiKeyAuth
// @Param    username path string true "Username"
// @Success  200 {object} response.Envelope
// @Router   /users/{username} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	if err := h.svc.DeleteUser(c.Request.Context(), c.Param("username")); err != nil {
		fail(c, err)
		return
	}
	response.NoContent(c, "User removed")
}

// DisconnectUser godoc
// @Summary  Disconnect a user's active sessions (CoA/PoD)
// @Tags     Users
// @Produce  json
// @Security ApiKeyAuth
// @Param    username path string true "Username"
// @Success  200 {object} response.Envelope
// @Router   /users/{username}/disconnect [post]
func (h *Handler) DisconnectUser(c *gin.Context) {
	count, err := h.svc.DisconnectUser(c.Request.Context(), c.Param("username"))
	if err != nil {
		fail(c, err)
		return
	}
	if count == 0 {
		response.OK(c, "No active sessions to disconnect", gin.H{"disconnected": 0})
		return
	}
	response.OK(c, "User disconnected", gin.H{"disconnected": count})
}

// ListSessions godoc
// @Summary  List active sessions
// @Tags     Sessions
// @Produce  json
// @Security ApiKeyAuth
// @Param    username query string false "Filter by username"
// @Success  200 {object} response.Envelope{data=[]models.RadAcct}
// @Router   /sessions [get]
func (h *Handler) ListSessions(c *gin.Context) {
	sessions, err := h.svc.ListSessions(c.Request.Context(), c.Query("username"))
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "OK", sessions)
}

// ListNAS godoc
// @Summary  List NAS clients
// @Tags     NAS
// @Produce  json
// @Security ApiKeyAuth
// @Success  200 {object} response.Envelope{data=[]models.Nas}
// @Router   /nas [get]
func (h *Handler) ListNAS(c *gin.Context) {
	list, err := h.svc.ListNAS(c.Request.Context())
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "OK", list)
}

// UpsertNAS godoc
// @Summary  Register or update a NAS client
// @Tags     NAS
// @Accept   json
// @Produce  json
// @Security ApiKeyAuth
// @Param    payload body dto.NASRequest true "NAS"
// @Success  200 {object} response.Envelope{data=models.Nas}
// @Router   /nas [post]
func (h *Handler) UpsertNAS(c *gin.Context) {
	var req dto.NASRequest
	if !bind(c, &req) {
		return
	}
	nas, err := h.svc.UpsertNAS(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	h.reloadFreeRADIUS(c)
	response.OK(c, "NAS saved", nas)
}

// reloadFreeRADIUS asks FreeRADIUS to re-read its SQL client list so a new or
// changed NAS takes effect immediately. Failures are logged, never fatal — the
// NAS row is already persisted; an operator can restart FreeRADIUS manually.
func (h *Handler) reloadFreeRADIUS(c *gin.Context) {
	if h.reloader == nil {
		return
	}
	if err := h.reloader.Reload(c.Request.Context()); err != nil {
		slog.Warn("FreeRADIUS reload failed; NAS saved but restart manually to apply",
			slog.Any("error", err))
	}
}

// DeleteNAS godoc
// @Summary  Remove a NAS client
// @Tags     NAS
// @Produce  json
// @Security ApiKeyAuth
// @Param    id path int true "NAS ID"
// @Success  200 {object} response.Envelope
// @Router   /nas/{id} [delete]
func (h *Handler) DeleteNAS(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, 400, "BAD_REQUEST", "Invalid id parameter")
		return
	}
	if err := h.svc.DeleteNAS(c.Request.Context(), uint(id)); err != nil {
		fail(c, err)
		return
	}
	h.reloadFreeRADIUS(c)
	response.NoContent(c, "NAS removed")
}
