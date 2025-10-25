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

	ID          uuid.UUID `bun:"id,pk,type:uuid,notnull,unique" json:"id" validate:"uuid4"`
	Email       string    `bun:"type:varchar(150),notnull,unique" json:"email" validate:"asci"`
	Username    string    `bun:"type:varchar(150),notnull,nullzero" json:"username" validate:"ascii"`
	Phone       string    `bun:"type:varchar(12),notnull" json:"phone" validate:"numeric"`
	FirstName   string    `bun:"type:varchar(100),notnull" json:"firstName" validate:"alpha,min=1,max=150"`
	LastName    string    `bun:"type:varchar(100),notnull" json:"lastName" validate:"alpha,min=1,max=150"`
	RoleID      uuid.UUID `bun:"type:uuid,notnull" json:"roleId" validate:"uuid4"`
	Role        *Role     `bun:"rel:belongs-to,join:role_id=id" json:"role"`
	WouldUse    bool      `bun:"type:boolean,notnull,nullzero,default:false" json:"wouldUse" validate:"boolean"`
	Comments    string    `bun:"type:text,nullzero" json:"comments" validate:"alphanum"`
	CompanyName string    `bun:"type:varchar(150),notnull,nullzero" json:"companyName" validate:"alphanum"`
	QuePosition int64     `bun:"type:integer,notnull,nullzero" json:"quePosition" validate:"number,min=1,"`
	// TODO: this survey ref needs to be updated because user_survey table removed
	Surveys    []Survey  `bun:"m2m:user_survey,join:User=Survey" json:"surveys"`
	Referals   []Referal `bun:"rel:has-many,join:id=referer_id" json:"referals"`
	ReferedBys []Referal `bun:"rel:has-many,join:id=referee_id" json:"referedBys"`
	ReferalURL string    `bun:"type:varchar(255),notnull,nullzero" json:"referalURL"`
	CreatedAt  time.Time `bun:"type:timestamptz,notnull,nullzero,default:current_timestamp" json:"createdAt"`
	UpdatedAt  time.Time `bun:"type:timestamptz,notnull,nullzero,default:current_timestamp" json:"updatedAt"`
}

func NewUser(uid, email, phne, company, fname, lname string, role Role, would bool) (*User, error) {
	UID, err := uuid.Parse(uid)
	if err != nil {
		return nil, err
	}
	return &User{
		ID:          UID,
		Email:       email,
		Phone:       phne,
		FirstName:   fname,
		LastName:    lname,
		RoleID:      role.ID,
		WouldUse:    would,
		CompanyName: company,
	}, nil
}

func (u User) Position() int {
	return int(u.QuePosition)
}

func (u User) RefLink() string {
	return fmt.Sprintf("https://www.estatehub.z3.com/refer/%v", u.ReferalURL)
}

func (u *User) Validate() error {
	v := validator.New(validator.WithPrivateFieldValidation())
	return v.Struct(u)
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
