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

func AuthMiddleware(userModel *models.UserModel) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		fmt.Println("Authorization header:", authHeader) // Debug log

		if authHeader == "" {
			fmt.Println("Missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			fmt.Println("Invalid authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenStr := bearerToken[1]
		fmt.Println("Token:", tokenStr) // Debug log

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return config.JWTSecret, nil
		})

		if err != nil {
			fmt.Println("Token parse error:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if !token.Valid {
			fmt.Println("Token is invalid")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			fmt.Println("Invalid token claims")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			fmt.Println("User ID not found in token claims")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			c.Abort()
			return
		}

		fmt.Println("User ID from token claims:", uint(userID)) // Debug log

		user, err := userModel.FindByID(uint(userID))
		if err != nil {
			fmt.Println("User not found:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		if !user.IsActive {
			fmt.Println("User is inactive")
			c.JSON(http.StatusForbidden, gin.H{"error": "Account is deactivated"})
			c.Abort()
			return
		}

		// Set both "user" and "user_id" in context
		c.Set("user", user)
		c.Set("user_id", user.ID)
		fmt.Println("User successfully authenticated:", user)
		c.Next()
	}
}

func SuperAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		currentUser, ok := user.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user data"})
			c.Abort()
			return
		}

		if !currentUser.IsSuperAdmin() {
			c.JSON(http.StatusForbidden, gin.H{"error": "Super admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func SysAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		currentUser, ok := user.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user data"})
			c.Abort()
			return
		}

		if !currentUser.IsSysAdmin() && !currentUser.IsSuperAdmin() {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
