package middleware

import (
	"github.com/gin-gonic/gin"

	"carhistory/internal/repository"
	"carhistory/pkg/response"
)

func AdminMiddleware(userRepo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok {
			response.Unauthorized(c, "unauthorized")
			c.Abort()
			return
		}

		user, err := userRepo.FindByID(c.Request.Context(), userID)
		if err != nil || user == nil || !user.IsAdmin {
			response.Forbidden(c, "admin access required")
			c.Abort()
			return
		}

		c.Next()
	}
}
