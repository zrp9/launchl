// Packag unit/work provides units of works for specific domain entities
package unit

import (
	"context"

	"github.com/zrp9/launchl/internal/domain"
)

type ReferalUnitOfWork interface {
	CreateUser(ctx context.Context, usr *domain.User) (*domain.User, error)
	GetReferer(ctx context.Context, username, urlID string) (*domain.User, error)
	RewardReferer(ctx context.Context, usr domain.User) error
	CreateReferal(ctx context.Context, usrID, refererID string) error
}
