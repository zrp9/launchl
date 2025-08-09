// Package userservice provides a interface for interacting with a user repo and http handlers.
package userservice

import (
	"context"
	"errors"

	v "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/zrp9/launchl/internal/auth"
	"github.com/zrp9/launchl/internal/domain"
	"github.com/zrp9/launchl/internal/dto"
	cfgr "github.com/zrp9/launchl/internal/repos/configrepo"
	usr "github.com/zrp9/launchl/internal/repos/userrepo"
)

type UserService struct {
	repo      usr.UserRepo
	cfgRepo   cfgr.ConfigRepo
	validator *v.Validate
}

func New(r usr.UserRepo, cfg cfgr.ConfigRepo, v *v.Validate) UserService {
	return UserService{
		repo:      r,
		cfgRepo:   cfg,
		validator: v,
	}
}

func (us UserService) Create(ctx context.Context, usr *domain.User) (*domain.User, error) {
	var err error
	usr.ID, err = uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	u, err := us.repo.Create(ctx, usr)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (us UserService) Update(ctx context.Context, usr domain.User) (*domain.User, error) {
	u, err := us.repo.Update(ctx, usr)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (us UserService) Get(ctx context.Context, id string) (*domain.User, error) {
	u, err := us.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (us UserService) GetAll(ctx context.Context) ([]*domain.User, error) {
	usrs, err := us.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return usrs, nil
}

func (us UserService) Delete(ctx context.Context, id string) error {
	if err := us.repo.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}

func (us UserService) Authenticate(ctx context.Context, usrname, pwd string) (bool, domain.User, error) {
	usr, err := us.repo.FetchByUsername(ctx, usrname)
	if err != nil {
		return false, domain.User{}, err
	}

	isMatch, err := auth.VerifyHash(usr.Password, pwd)
	if err != nil {
		return false, domain.User{}, err
	}

	return isMatch, usr, nil
}

func (us UserService) ValidateClaims(ctx context.Context, token string) (domain.User, error) {
	claims := auth.ParseAuthToken(token)
	usr, err := us.Get(ctx, claims.ID)

	if err != nil {
		return domain.User{}, err
	}

	if usr.Username != claims.Username {
		return domain.User{}, errors.New("claims do not match username")
	}

	if usr.Role.Name != claims.Role {
		return domain.User{}, errors.New("claims do not match role")
	}

	return *usr, nil
}

func (us UserService) SignupAdapter(signupDto dto.SignupDto) *domain.User {
	return &domain.User{
		Username:  signupDto.Username,
		Password:  signupDto.Password,
		Email:     signupDto.Email,
		FirstName: signupDto.FirstName,
		LastName:  signupDto.LastName,
	}
}
