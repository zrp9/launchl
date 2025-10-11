package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zrp9/launchl/internal/adapter/cache/valkaree"
	"github.com/zrp9/launchl/internal/adapter/repo/pgsql"
	"github.com/zrp9/launchl/internal/domain/core"
	"github.com/zrp9/launchl/internal/domain/repo"
)

type AppService struct {
	appRepo pgsql.AppRepo
	cache   valkaree.Cache
}

func NewAppService(r pgsql.AppRepo, c valkaree.Cache) AppService {
	return AppService{
		appRepo: r,
		cache:   c,
	}
}

func (a AppService) Name() string {
	return "app"
}

func (a AppService) CreateFeatures(ctx context.Context, feats []*core.Feature) error {
	fts, err := a.appRepo.BulkCreateFeatures(ctx, feats)
	if err != nil {
		return err
	}

	fs, err := core.Serialize(fts)
	if err != nil {
		return err
	}

	if err = a.cache.Set(ctx, "features", string(fs)); err != nil {
		return err
	}

	return nil
}

func (a AppService) CreateRoles(ctx context.Context, roles []*core.Role) error {
	return a.appRepo.BulkCreateRoles(ctx, roles)
}

func (a AppService) GetFeature(ctx context.Context, id string) (*core.Feature, error) {
	return a.appRepo.GetFeature(ctx, id)
}

func (a AppService) GetFeatures(ctx context.Context, pg, limit int) ([]core.Feature, error) {
	featCache, err := a.cache.Get(ctx, fmt.Sprintf("features:%v:%v", pg, limit))
	if err != nil && err != valkaree.ErrEmptyCache {
		return nil, err
	}

	if featCache != "" {
		var features []core.Feature
		if err = json.Unmarshal([]byte(featCache), &features); err != nil {
			return nil, err
		}
		return features, nil
	}

	features, err := a.appRepo.GetAllFeatures(ctx, repo.Pager{Page: pg, Limit: limit})
	if err != nil {
		return nil, err
	}

	jfeats, err := json.Marshal(features)
	if err != nil {
		return nil, err
	}

	if err = a.cache.Set(ctx, fmt.Sprintf("features:%v:%v", pg, limit), string(jfeats)); err != nil {
		return nil, err
	}

	return features, nil
}

func (a AppService) GetRole(ctx context.Context, name string) (core.Role, error) {
	cCache, err := a.cache.Get(ctx, fmt.Sprintf("role:%v", name))
	if err != nil && err != valkaree.ErrEmptyCache {
		return core.Role{}, err
	}

	if cCache != "" {
		var role core.Role
		if err = json.Unmarshal([]byte(cCache), &role); err != nil {
			return core.Role{}, err
		}

		return role, nil
	}

	role, err := a.appRepo.GetRole(ctx, name)
	if err != nil {
		return core.Role{}, err
	}

	jrole, err := json.Marshal(role)
	if err != nil {
		return core.Role{}, err
	}

	if err := a.cache.Set(ctx, fmt.Sprintf("role:%v", name), string(jrole)); err != nil {
		return core.Role{}, err
	}

	return role, nil
}

func (a AppService) GetRoles(ctx context.Context) ([]core.Role, error) {
	rCache, err := a.cache.Get(ctx, "roles")
	if err != nil && err != valkaree.ErrEmptyCache {
		return nil, fmt.Errorf("cache error %w", err)
	}

	if rCache != "" {
		var roles []core.Role
		if err = json.Unmarshal([]byte(rCache), &roles); err != nil {
			return nil, err
		}
		return roles, nil
	}

	roles, err := a.appRepo.GetAllRoles(ctx)
	if err != nil {
		return nil, err
	}

	return roles, nil
}
