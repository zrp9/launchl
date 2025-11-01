package seeder

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/zrp9/launchl/internal/domain/core"
	"github.com/zrp9/launchl/internal/domain/service"
)

type SeederAdapter struct {
	service service.AppService
	args    []string
}

func SeederFactory(s service.AppService, args []string) SeederAdapter {
	return SeederAdapter{
		service: s,
		args:    args,
	}
}

func (s SeederAdapter) seedFeatures() error {
	features := GetAppFeatures()
	feats := make([]*core.Feature, 0, len(features))
	for _, f := range features {
		feats = append(feats, &core.Feature{
			ID:               f.ID,
			Title:            f.Title,
			Name:             f.Name,
			Details:          strings.Join(f.Details, ","),
			QuickDescription: f.QuickDescription,
			Img:              f.Img,
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

func (s SeederAdapter) seedSurvey() error {
	survey := GetAppSurvey()
	log.Printf("questions  before calling service %v", survey.Questions)
	if err := s.service.CreateSurvey(context.TODO(), survey); err != nil {
		return err
	}

	return nil
}

func (s SeederAdapter) seedTestimonials() error {
	testimonials := GetTestimonials()
	if err := s.service.CreateTestimonials(context.TODO(), testimonials); err != nil {
		return err
	}

	return nil
}

func (s SeederAdapter) registry() map[string]func() error {
	return map[string]func() error{
		"features":     s.seedFeatures,
		"roles":        s.seedRoles,
		"survey":       s.seedSurvey,
		"testimonials": s.seedTestimonials,
	}
}

func (s SeederAdapter) LoadDBData() error {
	log.Println("Starting db seeding...")
	reg := s.registry()
	targets := normalizeArgs(s.args)
	log.Printf("normalized args %v", targets)
	if len(targets) == 1 && targets[0] == "all" {
		targets = keysInOrder(reg)
	}

	for _, name := range targets {
		fn, ok := reg[name]
		if !ok {
			return fmt.Errorf("unknown model %q (available: %s)", name, strings.Join(keysInOrder(reg), ", "))
		}
		log.Printf("seeding %s ...", name)
		if err := fn(); err != nil {
			return fmt.Errorf("seeding %s failed: %w", name, err)
		}
		log.Printf("seeding %s complete", name)
	}
	// if err := s.seedFeatures(); err != nil {
	// 	return err
	// }

	// if err := s.seedRoles(); err != nil {
	// 	return err
	// }
	return nil
}

func normalizeArgs(args []string) []string {
	var out []string
	for _, arg := range args {
		for _, p := range strings.Split(arg, ",") {
			p = strings.TrimSpace(strings.ToLower(p))
			if p != "" {
				out = append(out, p)
			}
		}
	}
	return out
}

func keysInOrder(m map[string]func() error) []string {
	order := []string{"roles", "features", "survey"}
	var out []string
	seen := map[string]bool{}
	for _, k := range order {
		if _, ok := m[k]; ok && !seen[k] {
			out = append(out, k)
			seen[k] = true
		}
	}

	for k := range m {
		if !seen[k] {
			out = append(out, k)
		}
	}

	return out
}
