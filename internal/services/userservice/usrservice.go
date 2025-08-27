// Package userservice provides a interface for interacting with a user repo and http handlers.
package userservice

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	v "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/zrp9/launchl/internal/config"
	"github.com/zrp9/launchl/internal/crane"
	"github.com/zrp9/launchl/internal/domain"
	"github.com/zrp9/launchl/internal/dto"
	cfgr "github.com/zrp9/launchl/internal/repos/configrepo"
	usr "github.com/zrp9/launchl/internal/repos/userrepo"
	"github.com/zrp9/launchl/internal/services/noti"
	"github.com/zrp9/launchl/internal/services/valkaree"
)

type UserService struct {
	repo         usr.UserRepo
	log          crane.Zlogrus
	cfgRepo      cfgr.ConfigRepo
	streamWriter valkaree.StreamWriter
	validator    *v.Validate
}

func New(r usr.UserRepo, cfg cfgr.ConfigRepo, writer valkaree.StreamWriter, v *v.Validate) UserService {
	return UserService{
		repo:         r,
		cfgRepo:      cfg,
		streamWriter: writer,
		validator:    v,
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

	// TODO: start here
	go func() {
		// returns message id, err
		data, err := us.createEmailPayload(usr, "welcome", "Welcome to launch list")
		if err != nil {
			us.log.MustTrace("could not create email json payload for notification stream")
		}
		if _, err := us.streamWriter.WriteJob(ctx, "email-notification", "email-consumer", "user-service", data); err != nil {
			us.log.MustTrace(fmt.Sprintf("failed to write job to stream %v", err))
		}
		// TODO: remove this log before deploying
		us.log.MustDebug("notification successfuly wrote to stream")
	}()

	return u, nil
}

func (us UserService) Update(ctx context.Context, usr domain.User) (*domain.User, error) {
	u, err := us.repo.Update(ctx, usr)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (us UserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, err := us.repo.GetByEmail(ctx, email)
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

func (us UserService) CheckQue(ctx context.Context, email string) (int64, error) {
	pos, err := us.repo.GetQuePosition(ctx, email)
	if err != nil {
		return -1, err
	}

	return pos, nil
}

func (us UserService) Delete(ctx context.Context, id string) error {
	if err := us.repo.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}

func (us UserService) DeleteByEmail(ctx context.Context, email string) error {
	if err := us.repo.DeleteByEmail(ctx, email); err != nil {
		return err
	}

	return nil
}

func (us UserService) SignupAdapter(signupDto dto.SignupDto) *domain.User {
	return &domain.User{
		Email:     signupDto.Email,
		FirstName: signupDto.FirstName,
		LastName:  signupDto.LastName,
	}
}

func (us UserService) createEmailPayload(usr *domain.User, notificationType, subject string) ([]byte, error) {
	emailCfg := config.LoadEmailConfig()
	to := []string{usr.Email}
	ejob := noti.EmailJob{
		To:              to,
		From:            emailCfg.Sender,
		Template:        notificationType,
		TemplateVersion: strconv.Itoa(emailCfg.TemplateVersion),
		Subject:         subject,
	}

	data, err := json.Marshal(ejob)
	if err != nil {
		return data, err
	}

	return data, nil
}
