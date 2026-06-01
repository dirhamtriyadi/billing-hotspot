package handlers

import (
	"net/http"
	"strconv"

	"github.com/dirhamt/billing-hotspot/backend/internal/response"
	"github.com/gin-gonic/gin"
)

// idParam parses a positive uint path parameter named "id", writing a 400 and
// returning false when it is missing or invalid.
func idParam(c *gin.Context) (uint, bool) {
	n, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || n == 0 {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid id parameter")
		return 0, false
	}
	return uint(n), true
}
