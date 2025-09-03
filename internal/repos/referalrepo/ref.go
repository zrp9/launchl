// Package referalrepo implements a basic repo for referals
package referalrepo

import (
	"context"

	"github.com/zrp9/launchl/internal/database/store"
	"github.com/zrp9/launchl/internal/domain"
	"github.com/zrp9/launchl/internal/repos"
)

type ReferalRepo struct {
	repo *repos.BasicRepo[string, domain.Referal]
}

func NewReferalRepo(p store.Persister) ReferalRepo {
	return ReferalRepo{
		repo: repos.New[string, domain.Referal](p),
	}
}

func (r ReferalRepo) Get(ctx context.Context, id string) (*domain.Referal, error) {
	return r.repo.Get(ctx, id)
}

func (r ReferalRepo) GetAll(ctx context.Context) ([]*domain.Referal, error) {
	return r.repo.GetAll(ctx)
}

func (r ReferalRepo) Create(ctx context.Context, referal *domain.Referal) (*domain.Referal, error) {
	return r.repo.Create(ctx, referal)
}

func (r ReferalRepo) Update(ctx context.Context, referal *domain.Referal) error {
	return r.repo.Update(ctx, referal.ID.String(), referal)
}

func (r ReferalRepo) Delete(ctx context.Context, id string) error {
	return r.repo.Delete(ctx, id)
}
