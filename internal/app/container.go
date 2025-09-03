// Package app proviced a container for repos and services
package app

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/zrp9/launchl/internal/crane"
	"github.com/zrp9/launchl/internal/database/store"
	"github.com/zrp9/launchl/internal/repos/configrepo"
	"github.com/zrp9/launchl/internal/repos/referalrepo"
	"github.com/zrp9/launchl/internal/repos/surveyrepo"
	"github.com/zrp9/launchl/internal/repos/userrepo"
	"github.com/zrp9/launchl/internal/services"
	"github.com/zrp9/launchl/internal/services/launch"
	"github.com/zrp9/launchl/internal/services/valkaree"
)

type Container struct {
	store     store.Persister
	logger    *crane.Zlogrus
	endpoints []services.Service
}

func (c Container) Endpoints() []services.Service {
	return c.endpoints
}

func New(s store.Persister, l *crane.Zlogrus) *Container {
	return &Container{
		store:  s,
		logger: l,
	}
}

func (c Container) RegisterServices(names []string) error {
	for _, name := range names {
		service, err := c.createService(name)
		if err != nil {
			return err
		}
		c.endpoints = append(c.endpoints, service)
	}

	return nil
}

func (c Container) createService(name string) (services.Service, error) {
	v := validator.New(validator.WithRequiredStructEnabled())
	configRepo := configrepo.NewRoleRepo(c.store)
	switch name {
	case "launch":
		userRepo := userrepo.New(c.store)
		questionRepo := surveyrepo.NewResponseRepo(c.store)
		refRepo := referalrepo.NewReferalRepo(c.store)
		s := valkaree.Stream{}
		sw := s.Writer()
		launchService := launch.New(userRepo, questionRepo, refRepo, configRepo, sw, v)
		return launch.Initialize(launchService, c.logger), nil
	default:
		return nil, fmt.Errorf("unknown service %v", name)
	}
}
