// Package app proviced a container for repos and services
package app

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/zrp9/launchl/internal/adapter/api/rest"
	"github.com/zrp9/launchl/internal/adapter/cache/valkaree"
	"github.com/zrp9/launchl/internal/adapter/log/crane"
	"github.com/zrp9/launchl/internal/adapter/repo/pgsql"
	"github.com/zrp9/launchl/internal/config"
	"github.com/zrp9/launchl/internal/domain/service"
)

type Container struct {
	store    pgsql.PGClient
	logger   *crane.Zlogrus
	services map[string]service.Servicer
	handlers []rest.Handler
}

func (c Container) Endpoints() map[string]service.Servicer {
	return c.services
}

func (c *Container) Handlers() []rest.Handler {
	return c.handlers
}

func New(s pgsql.PGClient, l *crane.Zlogrus) *Container {
	return &Container{
		store:    s,
		logger:   l,
		services: make(map[string]service.Servicer, 0),
		handlers: make([]rest.Handler, 0),
	}
}

func (c *Container) RegisterServices(ctx context.Context, names []string) error {
	vconf := config.LoadCacheAndStreamCfg()
	valkeyClient, err := valkaree.NewClient(ctx, vconf.Cache)
	if err != nil {
		return err
	}

	cache := valkaree.NewCache(valkeyClient)
	log.Printf("creating stream %v", vconf.Stream.Name)
	stream := valkaree.NewStream(valkeyClient, vconf.Stream.Name, vconf.Stream.WriteRetries, vconf.Stream.Threshold, *c.logger)

	for _, name := range names {
		service, err := c.createService(cache, *stream, name)
		if err != nil {
			return err
		}
		c.services[service.Name()] = service
	}

	log.Printf("services in container %v\n", len(c.services))

	if err := c.createHandlers(); err != nil {
		return err
	}

	return nil
}

func (c *Container) createHandlers() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	for _, service := range c.services {
		if err := c.handlerFactory(service, *v); err != nil {
			return err
		}
	}
	log.Printf("handlers in container %v\n", len(c.handlers))
	return nil
}

func (c *Container) handlerFactory(serv service.Servicer, v validator.Validate) error {
	switch serv.Name() {
	case "user":
		usrService, ok := serv.(service.UserService)
		if !ok {
			return c.ServiceErr(serv.Name())
		}
		c.handlers = append(c.handlers, rest.NewUserHandler(usrService, v, *c.logger))
		return nil
	case "survey":
		survService, ok := serv.(service.SurveyService)
		if !ok {
			return c.ServiceErr(serv.Name())
		}
		c.handlers = append(c.handlers, rest.NewSurveyHandler(survService, v, *c.logger))
		return nil
	case "app":
		aService, ok := serv.(service.AppService)
		if !ok {
			log.Printf("not type app service")
			return c.ServiceErr(serv.Name())
		}
		c.handlers = append(c.handlers, rest.NewAppHandler(aService, v, *c.logger))
		return nil
	default:
		log.Printf("hit default case")
		return errors.New("unsupported service name")
	}
}

func (c Container) createService(cache valkaree.Cache, stream valkaree.Stream, name string) (service.Servicer, error) {
	switch name {
	case "user":
		userRepo := pgsql.NewUserRepo(c.store)
		return service.NewUserService(userRepo, *c.logger, cache, stream), nil
	case "survey":
		surveyRepo := pgsql.NewSurveyRepo(c.store)
		return service.NewSurveyService(surveyRepo, cache, stream, *c.logger), nil
	case "app":
		appRepo := pgsql.NewAppRepo(c.store)
		return service.NewAppService(appRepo, cache), nil
	default:
		return nil, fmt.Errorf("unknown service %v", name)
	}
}

func (c Container) ServiceErr(service string) error {
	return fmt.Errorf("%v is invalid service type for handler", service)
}
