package seeder

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/zrp9/launchl/internal/database/store"
	"github.com/zrp9/launchl/internal/domain"
	"github.com/zrp9/launchl/internal/repos/configrepo"
	"github.com/zrp9/launchl/internal/services/feature"
)

type SeederAdapter struct {
	roleService    configrepo.RoleRepo
	featureService feature.FeatureService
}

func SeederFactory(s store.Persister) SeederAdapter {
	return SeederAdapter{
		roleService:    configrepo.NewRoleRepo(s),
		featureService: feature.New(s),
	}
}

func (s SeederAdapter) seedFeatures() error {
	log.Println("Starting feature seeder...")
	features := GetAppFeatures()
	feats := make([]domain.Feature, 0, len(features))
	for _, f := range features {
		feats = append(feats, domain.Feature{
			Title:            f.Title,
			Name:             f.Name,
			Details:          strings.Join(f.Details, ","),
			QuickDescription: f.QuickDescription,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		})
	}
	if err := s.featureService.BulkCreate(context.TODO(), feats); err != nil {
		return err
	}

	return nil
}

func (s SeederAdapter) seedRoles() error {
	log.Println("Starting role seeder...")
	return nil
}

func (s SeederAdapter) LoadDB() error {
	log.Println("Starting db seeding...")
	if err := s.seedFeatures(); err != nil {
		return err
	}

	if err := s.seedRoles(); err != nil {
		return err
	}
	return nil
}
