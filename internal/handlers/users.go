package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"spotifybackend/internal/config"
	"spotifybackend/internal/models"
	"spotifybackend/internal/utils"
)

func authMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing auth"})
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid auth header"})
			return
		}
		tok := parts[1]
		claims, err := utils.ParseToken(cfg.JWTSecret, tok)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token: " + err.Error()})
			return
		}
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

func RegisterUserRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg *config.Config) {
	users := rg.Group("/users")
	{
		users.GET("/me", authMiddleware(cfg), func(c *gin.Context) {
			uid := c.GetString("user_id")
			var u models.User
			if err := db.First(&u, "id = ?", uid).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"user": u})
		})
	}
}
