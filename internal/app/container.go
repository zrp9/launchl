// Package app proviced a container for repos and services
package app

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/zrp9/launchl/internal/crane"
	"github.com/zrp9/launchl/internal/database/store"
	"github.com/zrp9/launchl/internal/repos/configrepo"
	"github.com/zrp9/launchl/internal/repos/userrepo"
	"github.com/zrp9/launchl/internal/services"
	"github.com/zrp9/launchl/internal/services/userservice"
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
	configRepo := configrepo.New(c.store)
	switch name {
	case "user":
		userRepo := userrepo.New(c.store)
		usrService := userservice.New(userRepo, configRepo, v)
		return userservice.Initialize(usrService, c.logger), nil
	default:
		return nil, fmt.Errorf("unknown service %v", name)
	}
}
