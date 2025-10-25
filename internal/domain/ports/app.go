package ports

import (
	"context"

	"github.com/zrp9/launchl/internal/domain/core"
)

type FeaturePort interface {
	GetFeature(ctx context.Context, id string) (*core.Feature, error)
	GetAllFeatures(ctx context.Context) ([]core.Feature, error)
	BulkCreateFeatures(ctx context.Context, feats []core.Feature) error
	CreateFeature(ctx context.Context, feat *core.Feature) (*core.Feature, error)
	UpdateFeature(ctx context.Context, feat *core.Feature) error
	DeleteFeature(ctx context.Context, id string) error
}

type RolePort interface {
	GetRole(ctx context.Context, id string) (*core.Role, error)
	GetAllRoles(ctx context.Context) ([]core.Role, error)
	GetRoleByID(ctx context.Context, id string) (core.Role, error)
	BulkCreateRoles(ctx context.Context, roles []core.Role) error
}

type AppService interface {
	CreateFeatures(ctx context.Context, feats []*core.Feature) error
	CreateRoles(ctx context.Context, roles []*core.Role) error
	GetFeature(ctx context.Context, id string) (*core.Feature, error)
	GetFeatures(ctx context.Context, pg, limit int) (*core.Feature, error)
	GetRole(ctx context.Context, name string) (*core.Role, error)
	GetRoles(ctx context.Context) ([]*core.Role, error)
	CreateSurvey(ctx context.Context, survey *core.Survey)
}
