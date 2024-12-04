// middleware/auth_middleware.go
package middleware

import (
	"fmt"
	"net/http"
	"safesplit/config"
	"safesplit/models"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware verifies the JWT token and sets user info in context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return config.JWTSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Add user info to context
		c.Set("user_id", uint(claims["user_id"].(float64)))
		c.Set("user_role", claims["role"].(string))
		c.Next()
	}
}

// RequirePremiumUser middleware checks if user has premium role
func RequirePremiumUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != models.RolePremiumUser {
			c.JSON(http.StatusForbidden, gin.H{"error": "Premium access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireSysAdmin middleware checks if user has system admin role
func RequireSysAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != models.RoleSysAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "System admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireSuperAdmin middleware checks if user has super admin role
func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != models.RoleSuperAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Super admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
