package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	CtxUserIDKey   = "userID"
	CtxTenantIDKey = "tenantID"
	CtxRoleKey     = "role"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		claims, err := ValidateToken(parts[1], jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Set(CtxUserIDKey, claims.UserID)
		c.Set(CtxTenantIDKey, claims.TenantID)
		c.Set(CtxRoleKey, claims.Role)

		c.Next()
	}
}
