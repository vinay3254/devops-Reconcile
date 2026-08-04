package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/statuspage/api/internal/auth"
	"github.com/statuspage/api/internal/db"
	"github.com/statuspage/api/internal/k8s"
	"gorm.io/gorm"
)

type MonitorHandler struct {
	DB        *gorm.DB
	K8sClient *k8s.K8sClient
}

func NewMonitorHandler(database *gorm.DB, client *k8s.K8sClient) *MonitorHandler {
	return &MonitorHandler{
		DB:        database,
		K8sClient: client,
	}
}

type CreateMonitorRequest struct {
	Name            string `json:"name" binding:"required"`
	URL             string `json:"url" binding:"required,url"`
	IntervalSeconds int32  `json:"interval_seconds"`
	ExpectedStatus  int32  `json:"expected_status"`
	TimeoutSeconds  int32  `json:"timeout_seconds"`
}

func (h *MonitorHandler) CreateMonitor(c *gin.Context) {
	tenantIDVal, exists := c.Get(auth.CtxTenantIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context missing"})
		return
	}
	tenantID := tenantIDVal.(uuid.UUID)

	var req CreateMonitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.IntervalSeconds < 30 {
		req.IntervalSeconds = 60
	}
	if req.ExpectedStatus == 0 {
		req.ExpectedStatus = 200
	}
	if req.TimeoutSeconds == 0 {
		req.TimeoutSeconds = 10
	}

	// Check Tenant monitor tier limits
	var tenant db.Tenant
	if err := h.DB.First(&tenant, "id = ?", tenantID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify tenant tier"})
		return
	}

	var count int64
	h.DB.Model(&db.Monitor{}).Where("tenant_id = ?", tenantID).Count(&count)
	if int(count) >= tenant.MaxMonitors {
		c.JSON(http.StatusForbidden, gin.H{
			"error": fmt.Sprintf("Monitor limit reached for %s tier (%d max)", tenant.Tier, tenant.MaxMonitors),
		})
		return
	}

	monitorID := uuid.New()
	k8sNamespace := fmt.Sprintf("tenant-%s", tenantID.String()[:8])
	k8sName := fmt.Sprintf("mon-%s", monitorID.String()[:8])

	// Create K8s CR via Client-go
	err := h.K8sClient.CreateMonitorCR(
		c.Request.Context(),
		k8sNamespace,
		k8sName,
		req.URL,
		req.IntervalSeconds,
		req.ExpectedStatus,
		req.TimeoutSeconds,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to sync monitor to Kubernetes: %v", err)})
		return
	}

	// Create DB Record
	monitor := db.Monitor{
		ID:              monitorID,
		TenantID:        tenantID,
		Name:            req.Name,
		URL:             req.URL,
		IntervalSeconds: req.IntervalSeconds,
		ExpectedStatus:  req.ExpectedStatus,
		TimeoutSeconds:  req.TimeoutSeconds,
		K8sName:         k8sName,
		K8sNamespace:    k8sNamespace,
		Status:          "UNKNOWN",
	}

	if err := h.DB.Create(&monitor).Error; err != nil {
		// Rollback K8s CR if DB fails
		_ = h.K8sClient.DeleteMonitorCR(c.Request.Context(), k8sNamespace, k8sName)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to persist monitor"})
		return
	}

	c.JSON(http.StatusCreated, monitor)
}

func (h *MonitorHandler) ListMonitors(c *gin.Context) {
	tenantIDVal, _ := c.Get(auth.CtxTenantIDKey)
	tenantID := tenantIDVal.(uuid.UUID)

	var monitors []db.Monitor
	if err := h.DB.Where("tenant_id = ?", tenantID).Order("created_at desc").Find(&monitors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch monitors"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"monitors": monitors})
}

func (h *MonitorHandler) GetMonitor(c *gin.Context) {
	tenantIDVal, _ := c.Get(auth.CtxTenantIDKey)
	tenantID := tenantIDVal.(uuid.UUID)
	monitorIDStr := c.Param("id")

	monitorID, err := uuid.Parse(monitorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monitor ID format"})
		return
	}

	var monitor db.Monitor
	if err := h.DB.Preload("CheckResults", func(db *gorm.DB) *gorm.DB {
		return db.Order("timestamp desc").Limit(50)
	}).Where("id = ? AND tenant_id = ?", monitorID, tenantID).First(&monitor).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Monitor not found"})
		return
	}

	c.JSON(http.StatusOK, monitor)
}

func (h *MonitorHandler) DeleteMonitor(c *gin.Context) {
	tenantIDVal, _ := c.Get(auth.CtxTenantIDKey)
	tenantID := tenantIDVal.(uuid.UUID)
	monitorIDStr := c.Param("id")

	monitorID, err := uuid.Parse(monitorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid monitor ID format"})
		return
	}

	var monitor db.Monitor
	if err := h.DB.Where("id = ? AND tenant_id = ?", monitorID, tenantID).First(&monitor).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Monitor not found"})
		return
	}

	// Delete K8s CR
	if err := h.K8sClient.DeleteMonitorCR(c.Request.Context(), monitor.K8sNamespace, monitor.K8sName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to remove monitor from Kubernetes: %v", err)})
		return
	}

	// Delete DB Record
	if err := h.DB.Delete(&monitor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete monitor record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Monitor deleted successfully"})
}
