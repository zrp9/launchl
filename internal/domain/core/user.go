// Package core contains all the domain objects
package core

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID          uuid.UUID `bun:"id,pk,type:uuid,notnull,nullzero,default:uuid_generate_v4" json:"id" validate:"uuid4"`
	Email       string    `bun:"type:varchar(150),notnull,unique" json:"email" validate:"ascii"`
	FirstName   string    `bun:"type:varchar(100),notnull" json:"firstName,omitempty" validate:"alpha,min=1,max=150"`
	LastName    string    `bun:"type:varchar(100),notnull" json:"lastName,omitempty" validate:"alpha,min=1,max=150"`
	RoleID      uuid.UUID `bun:"type:uuid,notnull,nullzero" json:"roleId,omitempty" validate:"uuid4"`
	Role        *Role     `bun:"rel:belongs-to,join:role_id=id" json:"role,omitempty"`
	CompanyName string    `bun:"type:varchar(150),nullzero" json:"companyName,omitempty" validate:"ascii"`
	QuePosition int64     `bun:"type:integer,notnull,nullzero" json:"quePosition,omitempty" validate:"number,min=0"`
	Referals    []Referal `bun:"rel:has-many,join:id=referer_id" json:"referals,omitempty"`
	ReferedBys  []Referal `bun:"rel:has-many,join:id=referee_id" json:"referedBys,omitempty"`
	ReferalURL  string    `bun:"type:varchar(255),notnull,nullzero" json:"referalURL,omitempty"`
	CreatedAt   time.Time `bun:"type:timestamptz,notnull,nullzero,default:current_timestamp" json:"createdAt"`
	UpdatedAt   time.Time `bun:"type:timestamptz,notnull,nullzero,default:current_timestamp" json:"updatedAt"`
}

func NewUser(uid, email, phne, company, fname, lname string, role Role, would bool) (*User, error) {
	UID, err := uuid.Parse(uid)
	if err != nil {
		return nil, err
	}
	return &User{
		ID:          UID,
		Email:       email,
		FirstName:   fname,
		LastName:    lname,
		RoleID:      role.ID,
		CompanyName: company,
	}, nil
}

func (u User) Position() int {
	return int(u.QuePosition)
}

func (u User) RefLink() string {
	return u.ReferalURL
}

func (u *User) Validate() error {
	v := validator.New(validator.WithPrivateFieldValidation())
	return v.Struct(u)
}

func (u *User) SetRefLink(link string) {
	u.ReferalURL = fmt.Sprintf("https://www.estatehub.z3.com/refer/%v", link)
}

func (u *User) SetRoleID(id uuid.UUID) {
	u.RoleID = id
}

func (u *User) SetQuePosition(pos int64) {
	u.QuePosition = pos
}

var _ bun.BeforeAppendModelHook = (*User)(nil)

func (u *User) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		u.CreatedAt = time.Now()
	case *bun.UpdateQuery:
		u.UpdatedAt = time.Now()
	}
	return nil
}

func (u User) Info() string {
	return fmt.Sprintf("%#v\n", u)
}
