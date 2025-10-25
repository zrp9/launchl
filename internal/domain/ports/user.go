// Package ports define repo and service port/interfaces for domain
package ports

import (
	"context"

	"github.com/zrp9/launchl/internal/domain/core"
)

type IUserRepo interface {
	Create(ctx context.Context, usr *core.User, role string) (*core.User, error)
	GetByEmail(ctx context.Context, email string) (*core.User, error)
	GetQuePosition(ctx context.Context, email string) (int, error)
	UpdateByReferer(ctx context.Context, referee core.User, refererID string) error
	DeleteByEmail(ctx context.Context, email string) error
	GetByUsername(ctx context.Context, usrname string) (core.User, error)
}

type IUserService interface {
	CreateUser(ctx context.Context, u *core.User) (*core.User, error)
	CheckQuePosition(ctx context.Context, email string) (int, error)
	SignupReferal(ctx context.Context, referee core.User, refererID string) error
	DeleteUser(ctx context.Context, email string) error
	GetRefLinkAndPosition(ctx context.Context, email string) (core.User, error)
}
