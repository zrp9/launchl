// Package core referal entity
package core

import (
	"context"
	"os/user"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Referal struct {
	bun.BaseModel `bun:"table:referals,alias:rf"`
	ID            uuid.UUID  `bun:"id,pk,type:uuid,notnull" json:"id" validate:"required,uuidv4"`
	RefererID     uuid.UUID  `bun:"type:uuid,notnull" json:"refererId" validate:"required,uuidv4"`
	RefereeID     uuid.UUID  `bun:"type:uuid,notnull" json:"refereeId" validate:"required,uuidv4"`
	Referer       *user.User `bun:"rel:belongs-to,join:referer_id=id" json:"referer"`
	Referee       *user.User `bun:"rel:belongs-to,join:referee_id=id" json:"referee"`
}

type ReferalPort interface {
	Get(ctx context.Context, id string) (*Referal, error)
}
