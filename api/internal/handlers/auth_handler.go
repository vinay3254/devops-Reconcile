package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/statuspage/api/internal/auth"
	"github.com/statuspage/api/internal/db"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB        *gorm.DB
	JWTSecret string
}

func NewAuthHandler(database *gorm.DB, secret string) *AuthHandler {
	return &AuthHandler{
		DB:        database,
		JWTSecret: secret,
	}
}

type SignupRequest struct {
	TenantName string `json:"tenant_name" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token  string   `json:"token"`
	User   db.User  `json:"user"`
	Tenant db.Tenant `json:"tenant"`
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	var existingUser db.User
	if err := h.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// Create Tenant & Admin User in a single transaction
	tenant := db.Tenant{
		Name: req.TenantName,
		Tier: "free",
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&tenant).Error; err != nil {
			return err
		}

		user := db.User{
			TenantID:     tenant.ID,
			Email:        req.Email,
			PasswordHash: hash,
			Role:         "admin",
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant and user"})
		return
	}

	var user db.User
	h.DB.Where("email = ?", req.Email).First(&user)

	token, err := auth.GenerateToken(user.ID, tenant.ID, user.Role, h.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Token:  token,
		User:   user,
		Tenant: tenant,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user db.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	var tenant db.Tenant
	h.DB.First(&tenant, "id = ?", user.TenantID)

	token, err := auth.GenerateToken(user.ID, user.TenantID, user.Role, h.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token:  token,
		User:   user,
		Tenant: tenant,
	})
}
