// Package role entity
package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type RolePermission string

const (
	R     RolePermission = "read"
	W     RolePermission = "write"
	RW    RolePermission = "read-write"
	Admin RolePermission = "all"
)

type Role struct {
	bun.BaseModel `bun:"tabel:roles,alias:r"`
	ID            uuid.UUID      `bun:",pk,type:uuid" json:"id"`
	Name          string         `bun:"type:varchar(255),notnull,unique" json:"name" validate:"required,alpha,min=1,max=255"`
	Permissions   RolePermission `bun:"type:role_permission,notnull" json:"permissions" validate:"oneof='read' 'write' 'read-write' 'all'"`
	CreatedAt     time.Time      `bun:"type:timestamptz,notnull,nullzero,default=current_timestamp" json:"createdAt"`
	UpdatedAt     time.Time      `bun:"type:timestamptz,notnull,nullzero,default=current_timestamp" json:"updatedAt"`
}
