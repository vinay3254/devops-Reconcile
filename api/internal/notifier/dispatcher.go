package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/statuspage/api/internal/db"
	"gorm.io/gorm"
)

type NotificationPayload struct {
	Event       string    `json:"event"`        // "MONITOR_DOWN" or "MONITOR_RECOVERED"
	MonitorID   uuid.UUID `json:"monitor_id"`
	MonitorName string    `json:"monitor_name"`
	TargetURL   string    `json:"target_url"`
	Status      string    `json:"status"`
	StatusCode  int32     `json:"status_code"`
	Timestamp   time.Time `json:"timestamp"`
}

type Notifier struct {
	DB         *gorm.DB
	HTTPClient *http.Client
}

func NewNotifier(database *gorm.DB) *Notifier {
	return &Notifier{
		DB: database,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// DispatchAlert sends webhook notifications for a status transition event
func (n *Notifier) DispatchAlert(tenantID, monitorID uuid.UUID, monitorName, targetURL, newStatus string, statusCode int32) {
	var webhooks []db.WebhookConfig
	if err := n.DB.Where("tenant_id = ? AND enabled = true", tenantID).Find(&webhooks).Error; err != nil || len(webhooks) == 0 {
		return
	}

	event := "MONITOR_RECOVERED"
	if newStatus == "DOWN" {
		event = "MONITOR_DOWN"
	}

	payload := NotificationPayload{
		Event:       event,
		MonitorID:   monitorID,
		MonitorName: monitorName,
		TargetURL:   targetURL,
		Status:      newStatus,
		StatusCode:  statusCode,
		Timestamp:   time.Now().UTC(),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Notifier error: failed to marshal payload: %v", err)
		return
	}

	for _, hook := range webhooks {
		go func(targetURL string) {
			req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(jsonData))
			if err != nil {
				return
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("User-Agent", "StatusPage-NotifierWorker/1.0")

			resp, err := n.HTTPClient.Do(req)
			if err != nil {
				log.Printf("Notifier dispatch failed for %s: %v", targetURL, err)
				return
			}
			_ = resp.Body.Close()
			log.Printf("Successfully dispatched alert %s to %s (HTTP %d)", event, targetURL, resp.StatusCode)
		}(hook.URL)
	}
}
