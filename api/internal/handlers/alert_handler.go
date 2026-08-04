package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/statuspage/api/internal/auth"
	"github.com/statuspage/api/internal/db"
	"gorm.io/gorm"
)

type AlertHandler struct {
	DB *gorm.DB
}

func NewAlertHandler(database *gorm.DB) *AlertHandler {
	return &AlertHandler{DB: database}
}

type CreateWebhookRequest struct {
	URL    string `json:"url" binding:"required,url"`
	Secret string `json:"secret"`
}

func (h *AlertHandler) CreateWebhook(c *gin.Context) {
	tenantIDVal, exists := c.Get(auth.CtxTenantIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context missing"})
		return
	}
	tenantID := tenantIDVal.(uuid.UUID)

	var req CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	webhook := db.WebhookConfig{
		TenantID: tenantID,
		URL:      req.URL,
		Secret:   req.Secret,
		Enabled:  true,
	}

	if err := h.DB.Create(&webhook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save webhook config"})
		return
	}

	c.JSON(http.StatusCreated, webhook)
}

func (h *AlertHandler) ListWebhooks(c *gin.Context) {
	tenantIDVal, _ := c.Get(auth.CtxTenantIDKey)
	tenantID := tenantIDVal.(uuid.UUID)

	var webhooks []db.WebhookConfig
	if err := h.DB.Where("tenant_id = ?", tenantID).Find(&webhooks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch webhooks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"webhooks": webhooks})
}
