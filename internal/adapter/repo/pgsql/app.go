// Package pgsql - this file is for app level entities like feature/roles
package pgsql

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/zrp9/launchl/internal/domain/core"
	"github.com/zrp9/launchl/internal/domain/repo"
)

type AppRepo struct {
	repo *BasicRepo[string, core.Feature]
}

func NewAppRepo(p PGClient) AppRepo {
	return AppRepo{
		repo: NewBasicRepo[string, core.Feature](p),
	}
}

func (f AppRepo) BulkCreateFeatures(ctx context.Context, feats []*core.Feature) (features []*core.Feature, err error) {
	tx, err := f.repo.BnDB().BeginTx(ctx, nil)

	if err != nil {
		return nil, err
	}
	defer func() {
		if e := tx.Rollback(); e != nil {
			err = e
		}
	}()

	_, err = tx.NewInsert().Model(feats).Exec(ctx, &features)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return features, nil
}

func (f AppRepo) GetFeature(ctx context.Context, id string) (*core.Feature, error) {
	return f.repo.Get(ctx, id)
}

func (f AppRepo) GetAllFeatures(ctx context.Context, page repo.Pager) ([]core.Feature, error) {
	var features []core.Feature
	if err := f.repo.BnDB().NewSelect().Model(&features).Offset(page.Page).Limit(page.Limit).Scan(ctx, &features); err != nil {
		return nil, err
	}

	return features, nil
}

func (f AppRepo) CreateFeature(ctx context.Context, feat core.Feature) (*core.Feature, error) {
	return f.repo.Create(ctx, &feat)
}

func (f AppRepo) UpdateFeature(ctx context.Context, feat core.Feature) error {
	return f.repo.Update(ctx, feat.ID.String(), &feat)
}

func (f AppRepo) DeleteFeature(ctx context.Context, id string) error {
	return f.repo.Delete(ctx, id)
}

func (f AppRepo) New(ctx context.Context, name string) error {
	r := core.Role{
		Name: name,
	}

	return f.repo.BnDB().NewInsert().Model(&r).Scan(ctx)
}

func (f AppRepo) BulkCreateRoles(ctx context.Context, roles []*core.Role) (err error) {
	tx, err := f.repo.BnDB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if e := tx.Rollback(); e != nil {
			err = e
		}
	}()

	if _, err := tx.NewInsert().Model(roles).Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (f AppRepo) GetRole(ctx context.Context, name string) (core.Role, error) {
	var role core.Role
	err := f.repo.BnDB().NewSelect().Model(&role).Where("? = ?", bun.Ident("name"), name).Scan(ctx, &role)

	if err != nil {
		return core.Role{}, err
	}

	return role, nil
}

func (f AppRepo) GetAllRoles(ctx context.Context) ([]core.Role, error) {
	var roles []core.Role
	if err := f.repo.BnDB().NewSelect().Model(&roles).Scan(ctx, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}

func (f AppRepo) GetRoleByID(ctx context.Context, id string) (core.Role, error) {
	var role core.Role
	if err := f.repo.BnDB().NewSelect().Model(&role).Where("? = ?", bun.Ident("id"), id).Scan(ctx, &role); err != nil {
		return core.Role{}, err
	}

	return role, nil
}
