package middleware

import (
	"net/http"

	"github.com/Faithful001/aegis/internal/domain/organization"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TenantAuthorizer struct {
	orgRepo organization.Repository
}

func NewTenantAuthorizer(orgRepo organization.Repository) *TenantAuthorizer {
	return &TenantAuthorizer{orgRepo: orgRepo}
}

// RequireOrgMembership verifies that the JWT user belongs to the requested org_id URL param
func (a *TenantAuthorizer) RequireOrgMembership() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
			c.Abort()
			return
		}

		orgIDStr := c.Param("org_id")
		if orgIDStr == "" {
			orgIDStr = c.Param("id")
		}

		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid organization id"})
			c.Abort()
			return
		}

		member, err := a.orgRepo.GetMember(c.Request.Context(), orgID, userID)
		if err != nil || member == nil {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "you are not a member of this organization",
			})
			c.Abort()
			return
		}

		c.Set("userRole", member.Role)
		c.Next()
	}
}

// RequireOrgRole verifies that the user has at least one of the permitted roles
func (a *TenantAuthorizer) RequireOrgRole(allowedRoles ...organization.OrgRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "forbidden"})
			c.Abort()
			return
		}

		userRole := roleVal.(organization.OrgRole)
		for _, allowed := range allowedRoles {
			if userRole == allowed {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "insufficient permissions in this organization",
		})
		c.Abort()
	}
}
