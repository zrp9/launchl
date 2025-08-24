package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type AccessListQueue struct {
	bun.BaseModel `bun:"table:access_queue,alias:aq"`
	ID            uuid.UUID `bun:",pk,type:uuid" json:"id" validate:"uuidv4"`
	CurrentCount  int64     `bun:"type:bigint,notnull,nullzero,default=0" json:"currentCount" validate:"numeric,min=0"`
}

type Feature struct {
	bun.BaseModel `bun:"table:features,alias:f"`
	ID            uuid.UUID `bun:",pk,type:uuid" json:"id" validate:"uuidv4"`
	Name          string    `bun:"type:varchar(150),notnull,nullzero" json:"name" validate:"alphanum"`
	Details       string    `bun:"type:text,notnull,nullzero" json:"details" validate:"alphanum"`
	CreatedAt     time.Time `bun:"type:timestamptz,notnull,nullzero,default=current_timestamp" json:"createdAt"`
	UpdatedAt     time.Time `bun:"type:timestamptz,notnull,nullzero,default=current_timestamp" json:"updatedAt"`
	Images        []string  `json:"images"`
}
