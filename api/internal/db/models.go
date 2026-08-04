package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Tenant struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Tier        string    `gorm:"type:varchar(50);default:'free'" json:"tier"`
	MaxMonitors int       `gorm:"default:5" json:"max_monitors"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Users    []User    `gorm:"foreignKey:TenantID" json:"users,omitempty"`
	Monitors []Monitor `gorm:"foreignKey:TenantID" json:"monitors,omitempty"`
}

func (t *Tenant) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	TenantID     uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	Role         string    `gorm:"type:varchar(50);default:'admin'" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}

type Monitor struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	TenantID        uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Name            string    `gorm:"type:varchar(255);not null" json:"name"`
	URL             string    `gorm:"type:varchar(1024);not null" json:"url"`
	IntervalSeconds int32     `gorm:"default:60" json:"interval_seconds"`
	ExpectedStatus  int32     `gorm:"default:200" json:"expected_status"`
	TimeoutSeconds  int32     `gorm:"default:10" json:"timeout_seconds"`
	K8sName         string    `gorm:"type:varchar(255);not null" json:"k8s_name"`
	K8sNamespace    string    `gorm:"type:varchar(255);not null" json:"k8s_namespace"`
	Status          string    `gorm:"type:varchar(50);default:'UNKNOWN'" json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	CheckResults []CheckResult `gorm:"foreignKey:MonitorID;constraint:OnDelete:CASCADE" json:"check_results,omitempty"`
}

func (m *Monitor) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}

type CheckResult struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	MonitorID  uuid.UUID `gorm:"type:uuid;not null;index" json:"monitor_id"`
	StatusCode int32     `json:"status_code"`
	LatencyMs  int64     `json:"latency_ms"`
	Result     string    `gorm:"type:varchar(20)" json:"result"`
	Error      string    `gorm:"type:text" json:"error,omitempty"`
	Timestamp  time.Time `gorm:"index" json:"timestamp"`
}

type WebhookConfig struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	TenantID  uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	URL       string    `gorm:"type:varchar(1024);not null" json:"url"`
	Secret    string    `gorm:"type:varchar(255)" json:"secret,omitempty"`
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

func (w *WebhookConfig) BeforeCreate(tx *gorm.DB) (err error) {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return
}

