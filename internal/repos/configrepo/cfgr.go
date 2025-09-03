// Package configrepo repo is for types specific to app settings or config like roles
package configrepo

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/zrp9/launchl/internal/database/store"
	"github.com/zrp9/launchl/internal/domain"
	"github.com/zrp9/launchl/internal/repos"
)

type RoleRepo struct {
	repo *repos.BasicRepo[string, domain.Role]
}

func NewRoleRepo(p store.Persister) RoleRepo {
	return RoleRepo{
		repo: repos.New[string, domain.Role](p),
	}
}

func (c RoleRepo) New(ctx context.Context, name string) (domain.Role, error) {
	r := domain.Role{
		Name: name,
	}

	role, err := c.repo.Create(ctx, &r)
	if err != nil {
		return domain.Role{}, err
	}

	return *role, nil
}

func (c RoleRepo) Get(ctx context.Context, name string) (domain.Role, error) {
	var role domain.Role
	err := c.repo.BnDB().NewSelect().Model(&role).Where("? = ?", bun.Ident("name"), name).Scan(ctx, &role)

	if err != nil {
		return domain.Role{}, err
	}

	return role, nil
}

func (c RoleRepo) GetAll(ctx context.Context) ([]*domain.Role, error) {
	roles, err := c.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (c RoleRepo) GetByID(ctx context.Context, id string) (domain.Role, error) {
	role, err := c.repo.Get(ctx, id)
	if err != nil {
		return domain.Role{}, err
	}

	return *role, nil
}

type FeatureRepo struct {
	repo *repos.BasicRepo[string, domain.Feature]
}

func NewFeatureRepo(p store.Persister) FeatureRepo {
	return FeatureRepo{
		repo: repos.New[string, domain.Feature](p),
	}
}

func (f FeatureRepo) Get(ctx context.Context, id string) (*domain.Feature, error) {
	return f.repo.Get(ctx, id)
}

func (f FeatureRepo) GetAll(ctx context.Context) ([]*domain.Feature, error) {
	return f.repo.GetAll(ctx)
}

func (f FeatureRepo) Create(ctx context.Context, feat domain.Feature) (*domain.Feature, error) {
	return f.repo.Create(ctx, &feat)
}

func (f FeatureRepo) Update(ctx context.Context, feat domain.Feature) error {
	return f.repo.Update(ctx, feat.ID.String(), &feat)
}

func (f FeatureRepo) Delete(ctx context.Context, id string) error {
	return f.repo.Delete(ctx, id)
}
