package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/zrp9/launchl/internal/adapter/cache/valkaree"
	"github.com/zrp9/launchl/internal/adapter/log/crane"
	"github.com/zrp9/launchl/internal/adapter/noti"
	"github.com/zrp9/launchl/internal/adapter/repo/pgsql"
	"github.com/zrp9/launchl/internal/config"
	"github.com/zrp9/launchl/internal/domain/core"
	"github.com/zrp9/launchl/internal/dto"
	"github.com/zrp9/launchl/internal/hash"
)

var notificationSrc = "user-service"
var streamCfg = config.LoadVkENotifierConfig()

type UserService struct {
	repo   pgsql.UserRepo
	logger crane.Zlogrus
	cache  valkaree.Cache
	stream valkaree.StreamWriter
}

func (u UserService) Name() string {
	return "user"
}

func NewUserService(repo pgsql.UserRepo, logger crane.Zlogrus, cache valkaree.Cache, stream valkaree.Stream) UserService {
	return UserService{
		repo:   repo,
		logger: logger,
		cache:  cache,
		stream: stream.Writer(),
	}
}

func (u UserService) CreateUser(ctx context.Context, usr *core.User) (*core.User, error) {
	var err error
	usr.ID, err = uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	usr.SetRefLink(hash.GenerateHashLink(
		fmt.Sprintf("%vlessor-%v", usr.Email, time.Now()),
	))

	nUser, err := u.repo.Create(ctx, usr, "subscriber")
	if err != nil {
		return nil, err
	}

	go func() {
		cUser, err := json.Marshal(nUser)
		if err != nil {
			u.logger.MustTrace(fmt.Sprintf("failed to marshal user, %v", err))
		}

		if cUser != nil {
			if err = u.cache.Set(ctx, nUser.Email, string(cUser), valkaree.WithTTL(10*time.Minute)); err != nil {
				u.logger.MustTrace(fmt.Sprintf("failed to cache user data %v", err))
			}
		}

		data, err := dto.CreateEmailPayload(
			usr.Email,
			"Welcome to launch list",
			"welcome",
			streamCfg.SenderCfg.TemplateVersion,
			noti.WelcomeData{
				Name:        fmt.Sprintf("%v %v", usr.FirstName, usr.LastName),
				ReferralURL: usr.ReferalURL,
			},
		)

		if err != nil {
			u.logger.MustError(err)
		}

		if _, err := u.stream.WriteEvent(ctx, valkaree.Event{EventType: "welcome", Target: streamCfg.StreamCfg.Group, Src: notificationSrc}, data); err != nil {
			u.logger.MustTrace(fmt.Sprintf("failed to write event to stream %v", err))
		}
		u.logger.MustDebug("email notification wrote to stream successfully")
	}()

	return nUser, nil
}

func (u UserService) CheckQuePosition(ctx context.Context, email string) (int, error) {
	cacheVal, err := u.cache.Get(ctx, email)
	if err != nil && err != valkaree.ErrEmptyCache {
		u.logger.MustTrace(fmt.Sprintf("a cache error occurred: %v", err))
	}

	if cacheVal != "" {
		var usr core.User
		if err = json.Unmarshal([]byte(cacheVal), &usr); err != nil {
			u.logger.MustTrace(fmt.Sprintf("failed to unmarshal cache value for user %v", err))
		}

		// if usr == (core.User{}) {
		// }
		if usr.ID != uuid.Nil {
			return usr.Position(), nil
		}
	}

	usr, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		return -1, err
	}

	return usr.Position(), nil
}

func (u UserService) DeleteUser(ctx context.Context, email string) error {
	if err := u.cache.Delete(ctx, email); err != nil {
		u.logger.MustTrace(fmt.Sprintf("error could not delete user cache %v", err))
	}

	if err := u.cache.Delete(ctx, email); err != nil {
		u.logger.MustTrace(fmt.Sprintf("error occurred trying to delete usr cache %v", err))
	}
	return u.repo.DeleteByEmail(ctx, email)
}

func (u UserService) SignupReferal(ctx context.Context, referee core.User, refererID string) error {
	return u.repo.UpdateByReferer(ctx, referee, refererID)
}

func (u UserService) GetRefLinkAndPosition(ctx context.Context, usrname string) (string, int, error) {
	usrCached, err := u.cache.Get(ctx, usrname)
	if err != nil && err != valkaree.ErrEmptyCache {
		return "", -1, err
	}

	if usrCached != "" {
		var cUsr core.User
		if err = json.Unmarshal([]byte(usrCached), &cUsr); err != nil {
			return "", -1, err
		}
		return cUsr.ReferalURL, cUsr.Position(), nil
	}

	usr, err := u.repo.GetByUsername(ctx, usrname)
	if err != nil {
		return "", -1, err
	}

	return usr.ReferalURL, usr.Position(), nil
}

func (u UserService) createEmailPayload(usr *core.User, notificationType, subject string) ([]byte, error) {
	emailCfg := config.LoadEmailConfig()
	to := []string{usr.Email}
	evnt := dto.EmailDTO{
		To:              to,
		Template:        notificationType,
		TemplateVersion: strconv.Itoa(emailCfg.TemplateVersion),
		Subject:         subject,
	}

	data, err := json.Marshal(evnt)
	if err != nil {
		return nil, err
	}

	return data, nil
}
