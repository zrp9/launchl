// Package surveyrepo contains repositories for survey, survey question, question option, and survey response
package surveyrepo

import (
	"context"

	"github.com/zrp9/launchl/internal/database/store"
	"github.com/zrp9/launchl/internal/domain"
	"github.com/zrp9/launchl/internal/repos"
)

type SurveyRepo struct {
	repo *repos.BasicRepo[string, domain.Survey]
}

func NewSurveyRepo(p store.Persister) SurveyRepo {
	return SurveyRepo{
		repo: repos.New[string, domain.Survey](p),
	}
}

func (s SurveyRepo) Get(ctx context.Context, id string) (*domain.Survey, error) {
	return s.repo.Get(ctx, id)
}

func (s SurveyRepo) GetAll(ctx context.Context) ([]*domain.Survey, error) {
	return s.repo.GetAll(ctx)
}

func (s SurveyRepo) Create(ctx context.Context, survey *domain.Survey) (*domain.Survey, error) {
	return s.repo.Create(ctx, survey)
}

func (s SurveyRepo) Update(ctx context.Context, survey *domain.Survey) error {
	return s.repo.Update(ctx, survey.ID.String(), survey)
}

func (s SurveyRepo) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

type QuestionRepo struct {
	repo *repos.BasicRepo[string, domain.SurveyQuestion]
}

func NewSurveyQuestionRepo(p store.Persister) QuestionRepo {
	return QuestionRepo{
		repo: repos.New[string, domain.SurveyQuestion](p),
	}
}

func (s QuestionRepo) Get(ctx context.Context, id string) (*domain.SurveyQuestion, error) {
	return s.repo.Get(ctx, id)
}

func (s QuestionRepo) GetAll(ctx context.Context) ([]*domain.SurveyQuestion, error) {
	return s.repo.GetAll(ctx)
}

func (s QuestionRepo) Create(ctx context.Context, question *domain.SurveyQuestion) (*domain.SurveyQuestion, error) {
	return s.repo.Create(ctx, question)
}

func (s QuestionRepo) Update(ctx context.Context, question *domain.SurveyQuestion) error {
	return s.repo.Update(ctx, question.ID.String(), question)
}

func (s QuestionRepo) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

type QuestionOptionRepo struct {
	repo *repos.BasicRepo[string, domain.SurveyQuestionOption]
}

func NewQuestionOptionRepo(p store.Persister) QuestionOptionRepo {
	return QuestionOptionRepo{
		repo: repos.New[string, domain.SurveyQuestionOption](p),
	}
}

func (s QuestionOptionRepo) Get(ctx context.Context, id string) (*domain.SurveyQuestionOption, error) {
	return s.repo.Get(ctx, id)
}

func (s QuestionOptionRepo) GetAll(ctx context.Context) ([]*domain.SurveyQuestionOption, error) {
	return s.repo.GetAll(ctx)
}

func (s QuestionOptionRepo) Create(ctx context.Context, option *domain.SurveyQuestionOption) (*domain.SurveyQuestionOption, error) {
	return s.repo.Create(ctx, option)
}

func (s QuestionOptionRepo) Update(ctx context.Context, option *domain.SurveyQuestionOption) error {
	return s.repo.Update(ctx, option.ID.String(), option)
}

func (s QuestionOptionRepo) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

type ResponseRepo struct {
	repo *repos.BasicRepo[string, domain.SurveyResponse]
}

func NewResponseRepo(p store.Persister) ResponseRepo {
	return ResponseRepo{
		repo: repos.New[string, domain.SurveyResponse](p),
	}
}

func (s ResponseRepo) Get(ctx context.Context, id string) (*domain.SurveyResponse, error) {
	return s.repo.Get(ctx, id)
}

func (s ResponseRepo) GetAll(ctx context.Context) ([]*domain.SurveyResponse, error) {
	return s.repo.GetAll(ctx)
}

func (s ResponseRepo) Create(ctx context.Context, r *domain.SurveyResponse) (*domain.SurveyResponse, error) {
	return s.repo.Create(ctx, r)
}

func (s ResponseRepo) Update(ctx context.Context, r *domain.SurveyResponse) error {
	return s.repo.Update(ctx, r.ID.String(), r)
}

func (s ResponseRepo) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
