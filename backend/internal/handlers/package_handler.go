package handlers

import (
	"github.com/dirhamt/billing-hotspot/backend/internal/dto"
	"github.com/dirhamt/billing-hotspot/backend/internal/response"
	"github.com/dirhamt/billing-hotspot/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// PackageHandler exposes admin package CRUD.
type PackageHandler struct {
	svc *services.PackageService
}

// NewPackageHandler builds a PackageHandler.
func NewPackageHandler(svc *services.PackageService) *PackageHandler {
	return &PackageHandler{svc: svc}
}

// List godoc
// @Summary  List packages
// @Tags     Packages
// @Produce  json
// @Security BearerAuth
// @Param    page     query int    false "Page"
// @Param    per_page query int    false "Per page"
// @Param    search   query string false "Search by name"
// @Success  200 {object} response.Envelope{data=[]models.Package,meta=response.Meta}
// @Router   /packages [get]
func (h *PackageHandler) List(c *gin.Context) {
	var q dto.PageQuery
	if !bindQuery(c, &q) {
		return
	}
	q.Normalize()
	packages, total, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		fail(c, err)
		return
	}
	response.Paginated(c, "OK", packages, response.NewMeta(q.Page, q.PerPage, total))
}

// Get godoc
// @Summary  Get a package
// @Tags     Packages
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "Package ID"
// @Success  200 {object} response.Envelope{data=models.Package}
// @Router   /packages/{id} [get]
func (h *PackageHandler) Get(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	p, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "OK", p)
}

// Create godoc
// @Summary  Create a package
// @Tags     Packages
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    payload body dto.PackageRequest true "Package"
// @Success  201 {object} response.Envelope{data=models.Package}
// @Router   /packages [post]
func (h *PackageHandler) Create(c *gin.Context) {
	var req dto.PackageRequest
	if !bindJSON(c, &req) {
		return
	}
	p, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	response.Created(c, "Package created", p)
}

// Update godoc
// @Summary  Update a package
// @Tags     Packages
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "Package ID"
// @Param    payload body dto.PackageRequest true "Package"
// @Success  200 {object} response.Envelope{data=models.Package}
// @Router   /packages/{id} [put]
func (h *PackageHandler) Update(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var req dto.PackageRequest
	if !bindJSON(c, &req) {
		return
	}
	p, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Package updated", p)
}

// Delete godoc
// @Summary  Delete a package
// @Tags     Packages
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "Package ID"
// @Success  200 {object} response.Envelope
// @Router   /packages/{id} [delete]
func (h *PackageHandler) Delete(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		fail(c, err)
		return
	}
	response.NoContent(c, "Package deleted")
}
