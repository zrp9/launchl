package core

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Testimonial struct {
	bun.BaseModel `bun:"table:testimonials,alias:t"`

	ID        uuid.UUID `bun:"id,pk,type:uuid,notnull,unique" json:"id" validate:"uuid4"`
	Name      string    `bun:"type:varchar(150),notnull,unique" json:"name" validate:"asci"`
	Role      string    `bun:"type:varchar(150),notnull,nullzero" json:"role" validate:"ascii"`
	Rating    int       `bun:"type:integer,notnull" json:"rating" validate:"numeric"`
	Title     string    `bun:"type:varchar(100),notnull" json:"title" validate:"alpha,min=1,max=150"`
	Quote     string    `bun:"type:varchar(100),notnull" json:"quote" validate:"alpha,min=1,max=150"`
	Avatar    string    `bun:"type:varchar(150),notnull" json:"avatar"`
	Date      time.Time `bun:"type:timestamptz,notnull,nullzero,default:current_timestamp" json:"date"`
	CreatedAt time.Time `bun:"type:timestamptz,notnull,nullzero,default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:"type:timestamptz,notnull,nullzero,default:current_timestamp" json:"updatedAt"`
}
