package launch

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
	"github.com/zrp9/launchl/internal/eml"
	"github.com/zrp9/launchl/internal/repos/configrepo"
	"github.com/zrp9/launchl/internal/repos/referalrepo"
	"github.com/zrp9/launchl/internal/repos/surveyrepo"
	usr "github.com/zrp9/launchl/internal/repos/userrepo"
	"github.com/zrp9/launchl/internal/services/noti"
	"github.com/zrp9/launchl/internal/services/valkaree"
)

var (
	notificationType   = "email"
	notificationTarget = "email-consumer"
	notificationSrc    = "user-service"
)

type LaunchService struct {
	usrRepo      usr.UserRepo
	questnRepo   surveyrepo.ResponseRepo
	refRepo      referalrepo.ReferalRepo
	log          crane.Zlogrus
	cfgRepo      configrepo.RoleRepo
	streamWriter valkaree.StreamWriter
	validator    *v.Validate
}

func New(u usr.UserRepo, q surveyrepo.ResponseRepo, r referalrepo.ReferalRepo, cfg configrepo.RoleRepo, writer valkaree.StreamWriter, v *v.Validate) LaunchService {
	return LaunchService{
		usrRepo:      u,
		questnRepo:   q,
		refRepo:      r,
		cfgRepo:      cfg,
		streamWriter: writer,
		validator:    v,
	}
}

func (ls LaunchService) CreateUser(ctx context.Context, usr *domain.User) (*domain.User, error) {
	var err error
	usr.ID, err = uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	emailBase := eml.StripDomain(usr.Email)
	if emailBase == "" {
		return nil, fmt.Errorf("email address is required to generate username")
	}

	usr.Username = emailBase

	u, err := ls.usrRepo.Create(ctx, usr)
	if err != nil {
		return nil, err
	}

	// TODO: start here
	go func() {
		// returns message id, err
		data, err := ls.createEmailPayload(usr, "welcome", "Welcome to launch list")
		if err != nil {
			ls.log.MustTrace("could not create email json payload for notification stream")
		}
		if _, err := ls.streamWriter.WriteJob(ctx, notificationType, notificationTarget, notificationSrc, data); err != nil {
			ls.log.MustTrace(fmt.Sprintf("failed to write job to stream %v", err))
		}
		// TODO: remove this log before deploying
		ls.log.MustDebug("notification successfuly wrote to stream")
	}()

	return u, nil
}

func (ls LaunchService) UpdateUser(ctx context.Context, usr domain.User) (*domain.User, error) {
	u, err := ls.usrRepo.Update(ctx, usr)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (ls LaunchService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, err := ls.usrRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (ls LaunchService) GetUserByUsername(ctx context.Context, usrname string) (*domain.User, error) {
	u, err := ls.usrRepo.GetByUsername(ctx, usrname)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (ls LaunchService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	u, err := ls.usrRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (ls LaunchService) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	usrs, err := ls.usrRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return usrs, nil
}

func (ls LaunchService) CheckQue(ctx context.Context, usrname string) (int64, error) {
	pos, err := ls.usrRepo.GetQuePosition(ctx, usrname)
	if err != nil {
		return -1, err
	}

	return pos, nil
}

func (ls LaunchService) DeleteUser(ctx context.Context, id string) error {
	if err := ls.usrRepo.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}

func (ls LaunchService) DeleteUserByEmail(ctx context.Context, email string) error {
	if err := ls.usrRepo.DeleteByEmail(ctx, email); err != nil {
		return err
	}

	return nil
}

func (ls LaunchService) DeleteUserByUsername(ctx context.Context, usrname string) error {
	if err := ls.usrRepo.DeleteByUsername(ctx, usrname); err != nil {
		return err
	}

	return nil
}

func (ls LaunchService) AddSurveyResponses(ctx context.Context, usrID, questionID, optionID uuid.UUID, text string) error {
	if _, err := ls.questnRepo.Create(ctx, &domain.SurveyResponse{QuestionID: questionID, UserID: usrID, OptionID: optionID, WrittenResponse: text}); err != nil {
		return err
	}

	return nil
}

func (ls LaunchService) RewardReferer(ctx context.Context, referer domain.User) error {
	referer.QuePosition += 1
	if _, err := ls.UpdateUser(ctx, referer); err != nil {
		return err
	}
	return nil
}

func (ls LaunchService) GetReferer(ctx context.Context, usrname, urlID string) (domain.User, error) {
	referer, err := ls.usrRepo.GetReferer(ctx, usrname, urlID)
	if err != nil {
		return domain.User{}, err
	}

	return referer, nil
}

func (ls LaunchService) CreateReferal(ctx context.Context, refererID, refereeID uuid.UUID) error {
	ref := domain.Referal{
		RefereeID: refereeID,
		RefererID: refererID,
	}
	if _, err := ls.refRepo.Create(ctx, &ref); err != nil {
		return err
	}

	return nil
}

func (ls LaunchService) createEmailPayload(usr *domain.User, notificationType, subject string) ([]byte, error) {
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
