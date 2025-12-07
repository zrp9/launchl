package pgsql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/uptrace/bun"
	"github.com/zrp9/launchl/internal/domain/core"
	"github.com/zrp9/launchl/internal/hash"
)

type SurveyRepo struct {
	repo *BasicRepo[string, core.Survey]
}

func NewSurveyRepo(p PGClient) SurveyRepo {
	return SurveyRepo{
		repo: NewBasicRepo[string, core.Survey](p),
	}
}

func (s SurveyRepo) GetSurvey(ctx context.Context, id string) (*core.Survey, error) {
	return s.repo.Get(ctx, id)
}

func (s SurveyRepo) GetActiveSurvey(ctx context.Context) (*core.Survey, error) {
	var survey core.Survey
	err := s.repo.BnDB().NewSelect().
		Model(&survey).
		Relation("Questions", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Where("sq.active = TRUE").OrderExpr("sq.position ASC")
		}).
		Relation("Questions.Options", func(sqo *bun.SelectQuery) *bun.SelectQuery { return sqo.Order("position ASC") }).
		Where("? = ?", bun.Ident("active"), true).
		Limit(1).
		Scan(ctx)
	if err != nil {
		log.Println("query err here ")
		if err == sql.ErrNoRows {
			log.Println("no rows err")
			return nil, ErrNoRecords
		}
		return nil, err
	}

	return &survey, nil
}

func (s SurveyRepo) GetAllSurveys(ctx context.Context) ([]*core.Survey, error) {
	return s.repo.GetAll(ctx)
}

func (s SurveyRepo) CreateSurvey(ctx context.Context, survey *core.Survey) (*core.Survey, error) {
	return s.repo.Create(ctx, survey)
}

func (s SurveyRepo) UpdateSurvey(ctx context.Context, survey *core.Survey) error {
	return s.repo.Update(ctx, survey.ID.String(), survey)
}

func (s SurveyRepo) DeleteSurvey(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s SurveyRepo) GetQuestion(ctx context.Context, id string) (*core.Question, error) {
	var question core.Question
	if err := s.repo.BnDB().NewSelect().Model(&question).Where("? = ?", bun.Ident("id"), id).Scan(ctx, &question); err != nil {
		return nil, err
	}

	return &question, nil
}

func (s SurveyRepo) GetAllQuestions(ctx context.Context) ([]*core.Question, error) {
	var questions []*core.Question
	if err := s.repo.BnDB().NewSelect().Model(&questions).Scan(ctx, &questions); err != nil {
		return nil, err
	}

	return questions, nil
}

func (s SurveyRepo) CreateQuestion(ctx context.Context, question *core.Question) error {
	if err := s.repo.BnDB().RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().Model(&question).Exec(ctx); err != nil {
			return err
		}

		if _, err := tx.NewInsert().Model(&question.Options).Exec(ctx); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s SurveyRepo) UpdateQuestion(ctx context.Context, question *core.Question) error {
	return s.repo.BnDB().NewUpdate().Model(&question).Scan(ctx, &question)
}

func (s SurveyRepo) DeleteQuestion(ctx context.Context, id string) error {
	if err := s.repo.BnDB().RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if err := tx.NewDelete().Model(&core.SurveyQuestionOption{}).Where("? = ?", bun.Ident("question_id"), id).Scan(ctx, core.SurveyQuestionOption{}); err != nil {
			return err
		}

		if err := tx.NewDelete().Model(&core.Question{}).Where("? = ?", bun.Ident("id"), id).Scan(ctx, &core.Question{}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s SurveyRepo) CreateQuestionOption(ctx context.Context, option *core.SurveyQuestionOption) error {
	return s.repo.BnDB().NewInsert().Model(&option).Scan(ctx)
}

func (s SurveyRepo) UpdateQuestionOption(ctx context.Context, option *core.SurveyQuestionOption) error {
	return s.repo.BnDB().NewInsert().Model(&option).Scan(ctx)
}

func (s SurveyRepo) DeleteOption(ctx context.Context, id string) error {
	return s.repo.BnDB().NewDelete().Model(&core.SurveyQuestionOption{}).Scan(ctx)
}

func (s SurveyRepo) GetUserResponse(ctx context.Context, userID string) (*core.SurveyResponse, error) {
	var response core.SurveyResponse
	if err := s.repo.BnDB().NewSelect().Model(&response).Where("? = ?", bun.Ident("user_id"), userID).Scan(ctx, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (s SurveyRepo) GetAllSurveyResponses(ctx context.Context) ([]*core.SurveyResponse, error) {
	var responses []*core.SurveyResponse
	if err := s.repo.BnDB().NewSelect().Model(&responses).Scan(ctx, &responses); err != nil {
		return nil, err
	}

	return responses, nil
}

func (s SurveyRepo) CreateSurveyResponse(ctx context.Context, user *core.User, responses []core.SurveyResponse) error {

	if err := s.repo.BnDB().RunInTx(ctx, &sql.TxOptions{ReadOnly: false}, func(ctx context.Context, tx bun.Tx) error {
		if err := tx.NewSelect().Model(user).Where("email = ?", user.Email).Scan(ctx); err != nil {

			if err != sql.ErrNoRows {
				return err
			}

			var role core.Role
			err = tx.NewSelect().Model(&role).Where("? = ?", bun.Ident("name"), "subscriber").Scan(ctx, &role)
			if err != nil {
				return err
			}

			user.SetRoleID(role.ID)
			usrCount, err := tx.NewSelect().Model((*core.User)(nil)).Count(ctx)
			if err != nil {
				return err
			}

			user.SetQuePosition(int64(usrCount) + 1)
			key := fmt.Sprintf("%vlessor-%v", user.Email, time.Now())
			hashLink := hash.GenerateHashLink(key)
			user.SetRefLink(hashLink)
			err = tx.NewInsert().Model(user).Scan(ctx, user)
			if err != nil {
				return err
			}
		}

		for i := range responses {
			responses[i].UserID = user.ID
		}

		if _, err := tx.NewInsert().Model(&responses).Exec(ctx); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
