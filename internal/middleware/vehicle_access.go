package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"carhistory/internal/repository"
	"carhistory/pkg/response"
)

func VehicleAccessMiddleware(vehicleRepo *repository.VehicleRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok {
			response.Unauthorized(c, "unauthorized")
			c.Abort()
			return
		}

		vehicleIDStr := c.Param("vehicleId")
		vehicleID, err := uuid.Parse(vehicleIDStr)
		if err != nil {
			response.BadRequest(c, "invalid vehicle id")
			c.Abort()
			return
		}

		member, err := vehicleRepo.GetMember(c.Request.Context(), vehicleID, userID)
		if err != nil {
			response.InternalError(c, "failed to check vehicle access")
			c.Abort()
			return
		}
		if member == nil {
			response.Forbidden(c, "you do not have access to this vehicle")
			c.Abort()
			return
		}

		c.Set(ContextUserRole, member.Role)
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextUserRole)
		if !exists {
			response.Forbidden(c, "access denied")
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			response.Forbidden(c, "access denied")
			c.Abort()
			return
		}

		for _, r := range roles {
			if r == userRole {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "insufficient permissions for this action")
		c.Abort()
	}
}

func GetVehicleRole(c *gin.Context) string {
	role, _ := c.Get(ContextUserRole)
	r, _ := role.(string)
	return r
}
