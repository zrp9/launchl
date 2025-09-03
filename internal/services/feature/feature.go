// Package feature contains feature service
package feature

import (
	"context"

	"github.com/zrp9/launchl/internal/database/store"
	"github.com/zrp9/launchl/internal/domain"
	"github.com/zrp9/launchl/internal/repos"
)

type FeatureService struct {
	repo *repos.BasicRepo[string, domain.Feature]
}

func New(p store.Persister) FeatureService {
	return FeatureService{
		repo: repos.New[string, domain.Feature](p),
	}
}

func (f FeatureService) Get(ctx context.Context, id string) (*domain.Feature, error) {
	return f.repo.Get(ctx, id)
}

func (f FeatureService) GetAll(ctx context.Context) ([]*domain.Feature, error) {
	return f.repo.GetAll(ctx)
}

func (f FeatureService) Create(ctx context.Context, feat domain.Feature) (*domain.Feature, error) {
	return f.repo.Create(ctx, &feat)
}

func (f FeatureService) Update(ctx context.Context, feat domain.Feature) error {
	return f.repo.Update(ctx, feat.ID.String(), &feat)
}

func (f FeatureService) BulkCreate(ctx context.Context, feats []domain.Feature) (err error) {
	tx, err := f.repo.BnDB().BeginTx(ctx, nil)

	if err != nil {
		return err
	}
	defer func() {
		if e := tx.Rollback(); e != nil {
			err = e
		}
	}()

	_, err = f.repo.BnDB().NewInsert().Model(&feats).Exec(ctx)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (f FeatureService) Delete(ctx context.Context, id string) error {
	return f.repo.Delete(ctx, id)
}
