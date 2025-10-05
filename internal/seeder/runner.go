package seeder

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/zrp9/launchl/internal/domain/core"
	"github.com/zrp9/launchl/internal/domain/service"
)

type SeederAdapter struct {
	service service.AppService
}

func SeederFactory(s service.AppService) SeederAdapter {
	return SeederAdapter{
		service: s,
	}
}

func (s SeederAdapter) seedFeatures() error {
	log.Println("Starting feature seeder...")
	features := GetAppFeatures()
	feats := make([]*core.Feature, 0, len(features))
	for _, f := range features {
		feats = append(feats, &core.Feature{
			Title:            f.Title,
			Name:             f.Name,
			Details:          strings.Join(f.Details, ","),
			QuickDescription: f.QuickDescription,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		})
	}
	if err := s.service.CreateFeatures(context.TODO(), feats); err != nil {
		return err
	}

	return nil
}

func (s SeederAdapter) seedRoles() error {
	log.Println("Starting role seeder...")
	roles := GetAppRoles()
	appRoles := make([]*core.Role, 0, len(roles))
	for _, role := range roles {
		appRoles = append(appRoles, &core.Role{
			ID:          role.ID,
			Name:        role.Name,
			Permissions: core.RolePermission(role.Permissions),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		})
	}

	if err := s.service.CreateRoles(context.TODO(), appRoles); err != nil {
		return err
	}
	return nil
}

func (s SeederAdapter) LoadDBData() error {
	log.Println("Starting db seeding...")
	if err := s.seedFeatures(); err != nil {
		return err
	}

	if err := s.seedRoles(); err != nil {
		return err
	}
	return nil
}
