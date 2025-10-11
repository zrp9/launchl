// Package core role entity
package core

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type RolePermission string

const (
	R          RolePermission = "read"
	W          RolePermission = "write"
	RW         RolePermission = "read-write"
	A          RolePermission = "all"
	Admin      RolePermission = "admin"
	SuperAdmin RolePermission = "super-admin"
)

type Role struct {
	bun.BaseModel `bun:"table:roles,alias:r"`
	ID            uuid.UUID      `bun:"id,pk,type:uuid" json:"id"`
	Name          string         `bun:"type:varchar(255),notnull,unique" json:"name" validate:"required,alpha,min=1,max=255"`
	Permissions   RolePermission `bun:"type:role_permission,notnull" json:"permissions" validate:"oneof='read' 'write' 'read-write' 'all'"`
	CreatedAt     time.Time      `bun:"type:timestamptz,notnull,nullzero,default:current_timestamp" json:"createdAt"`
	UpdatedAt     time.Time      `bun:"type:timestamptz,notnull,nullzero,default:current_timestamp" json:"updatedAt"`
}
