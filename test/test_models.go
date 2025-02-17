package test

import (
	"time"
)

// Tenant represents a tenant in the system.
type Tenant struct {
	TenantID  int32     `json:"tenant_id"`
	Name      string    `json:"name"`
	Plan      string    `json:"plan"`
	CreatedAt time.Time `json:"created_at"`
}

// Resource represents a resource owned by a tenant.
type Resource struct {
	TenantID   int32     `json:"tenant_id"`
	ResourceID int32     `json:"resource_id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	CreatedAt  time.Time `json:"created_at"`
}
