// Package domain contains all the domain objects
package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	// TODO: check that making json for avatar be avatarUrl doesnt' cause problems
	ID          uuid.UUID `bun:"pk,type:uuid,notnull,unique" json:"uid" validate:"uuid4"`
	Email       string    `bun:"type:varchar(150),notnull,unique" json:"email" validate:"asci"`
	Phone       string    `bun:"type:varchar(12),notnull" json:"phone" validate:"numeric"`
	Username    string    `bun:"type:varchar(75),notnull" json:"username" validate:"alphanum"`
	Password    string    `bun:"type:varchar(255),notnull" json:"password"`
	FirstName   string    `bun:"type:varchar(100),notnull" json:"firstName" validate:"alpha,min=1,max=150"`
	LastName    string    `bun:"type:varchar(100),notnull" json:"lastName" validate:"alpha,min=1,max=150"`
	RoleID      uuid.UUID `bun:"type:uuid,notnull" json:"roleId"`
	Role        *Role     `bun:"rel:belongs-to,join:role_id=id" json:"role"`
	WouldUse    bool      `bun:"type:boolean,notnull,nullzero,default=false" json:"wouldUse" validate:"boolean"`
	Comments    string    `bun:"type:text,null,nullzero" json:"comments"`
	Address     *Address  `bun:"type:jsonb,notnull,nullzero" json:"address"`
	CompanyName string    `bun:"type:varchar(150),null,nullzero" json:"companyName" validate:"alphanum"`
	Surveys     []Survey  `bun:"m2m:user_survey,join:User=Survey" json:"surveys"`
	CreatedAt   time.Time `bun:"type:timestamptz,notnull,nullzero,default=current_timestamp" json:"createdAt"`
	UpdatedAt   time.Time `bun:"type:timestamptz,notnull,nullzero,default=current_timestamp" json:"updatedAt"`
}

func NewUser(uid, avatar, email, phne, usrname, fname, lname, company string, role Role, would bool, adrs Address) (*User, error) {
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
		Address:     &adrs,
		CompanyName: company,
	}, nil
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
