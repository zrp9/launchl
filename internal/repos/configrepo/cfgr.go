// Package configrepo repo is for types specific to app settings or config like roles
package configrepo

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/zrp9/launchl/internal/database/store"
	"github.com/zrp9/launchl/internal/domain"
	"github.com/zrp9/launchl/internal/repos"
)

type ConfigRepo struct {
	repo *repos.BasicRepo[string, domain.Role]
}

func New(p store.Persister) ConfigRepo {
	return ConfigRepo{
		repo: repos.New[string, domain.Role](p),
	}
}

func (c ConfigRepo) NewRole(ctx context.Context, name string) (domain.Role, error) {
	r := domain.Role{
		Name: name,
	}

	role, err := c.repo.Create(ctx, &r)
	if err != nil {
		return domain.Role{}, err
	}

	return *role, nil
}

func (c ConfigRepo) GetRole(ctx context.Context, name string) (domain.Role, error) {
	var role domain.Role
	err := c.repo.BnDB().NewSelect().Model(&role).Where("? = ?", bun.Ident("name"), name).Scan(ctx, &role)

	if err != nil {
		return domain.Role{}, err
	}

	return role, nil
}

func (c ConfigRepo) GetRoles(ctx context.Context) ([]*domain.Role, error) {
	roles, err := c.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (c ConfigRepo) GetRoleByID(ctx context.Context, id string) (domain.Role, error) {
	role, err := c.repo.Get(ctx, id)
	if err != nil {
		return domain.Role{}, err
	}

	return *role, nil
}
