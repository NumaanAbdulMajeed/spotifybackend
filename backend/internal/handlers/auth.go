package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"spotifybackend/internal/config"
	"spotifybackend/internal/models"
	"spotifybackend/internal/utils"
)

type registerReq struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	DisplayName string `json:"display_name"`
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func RegisterAuthRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg *config.Config) {
	auth := rg.Group("/auth")
	{
		auth.POST("/register", func(c *gin.Context) {
			var req registerReq
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
				return
			}
			u := &models.User{
				Email:        req.Email,
				PasswordHash: string(hash),
				DisplayName:  req.DisplayName,
			}
			if err := db.Create(u).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "email maybe already used: " + err.Error()})
				return
			}
			token, err := utils.GenerateAccessToken(cfg.JWTSecret, u.ID, cfg.JWTAccessTTLMin)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
				return
			}
			c.JSON(http.StatusCreated, gin.H{"user": u, "access_token": token})
		})

		auth.POST("/login", func(c *gin.Context) {
			var req loginReq
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			var u models.User
			if err := db.Where("email = ?", req.Email).First(&u).Error; err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
				return
			}
			if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
				return
			}
			token, err := utils.GenerateAccessToken(cfg.JWTSecret, u.ID, cfg.JWTAccessTTLMin)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"user": u, "access_token": token})
		})
	}
}
