package pgsql

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/zrp9/launchl/internal/domain/core"
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
		Relation("SurveyQuestion", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Where("sq.active = TRUE").OrderExpr("sq.position ASC")
		}).
		Relation("Questions.Options", func(sqo *bun.SelectQuery) *bun.SelectQuery { return sqo.Where("sqo.position ASC") }).
		Where("? = ?", bun.Ident("active"), true).
		Limit(1).
		Scan(ctx, &survey)
	if err != nil {
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

func (s SurveyRepo) GetQuestion(ctx context.Context, id string) (*core.SurveyQuestion, error) {
	var question core.SurveyQuestion
	if err := s.repo.BnDB().NewSelect().Model(&question).Where("? = ?", bun.Ident("id"), id).Scan(ctx, &question); err != nil {
		return nil, err
	}

	return &question, nil
}

func (s SurveyRepo) GetAllQuestions(ctx context.Context) ([]*core.SurveyQuestion, error) {
	var questions []*core.SurveyQuestion
	if err := s.repo.BnDB().NewSelect().Model(&questions).Scan(ctx, &questions); err != nil {
		return nil, err
	}

	return questions, nil
}

func (s SurveyRepo) CreateQuestion(ctx context.Context, question *core.SurveyQuestion) error {
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

func (s SurveyRepo) UpdateQuestion(ctx context.Context, question *core.SurveyQuestion) error {
	return s.repo.BnDB().NewUpdate().Model(&question).Scan(ctx, &question)
}

func (s SurveyRepo) DeleteQuestion(ctx context.Context, id string) error {
	if err := s.repo.BnDB().RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if err := tx.NewDelete().Model(&core.SurveyQuestionOption{}).Where("? = ?", bun.Ident("question_id"), id).Scan(ctx, core.SurveyQuestionOption{}); err != nil {
			return err
		}

		if err := tx.NewDelete().Model(&core.SurveyQuestion{}).Where("? = ?", bun.Ident("id"), id).Scan(ctx, &core.SurveyQuestion{}); err != nil {
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

func (s SurveyRepo) CreateSurveyResponse(ctx context.Context, responses []core.SurveyResponse) error {
	if err := s.repo.BnDB().RunInTx(ctx, &sql.TxOptions{Isolation: 0, ReadOnly: false}, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().Model(&responses).Exec(ctx); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
