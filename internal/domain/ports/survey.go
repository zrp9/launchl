package ports

import (
	"context"

	"github.com/zrp9/launchl/internal/domain"
	"github.com/zrp9/launchl/internal/domain/core"
)

type ISurveyRepo interface {
	GetSurvey(ctx context.Context, id string) (*core.Survey, error)
	GetActiveSurvey(ctx context.Context) (*core.Survey, error)
	CreateSurvey(ctx context.Context, usr *core.Survey) (*core.Survey, error)
	UpdateSurvey(ctx context.Context, usr *core.Survey) error
	DeleteSurvey(ctx context.Context, id string) error
	GetQuestion(ctx context.Context, id string) (*core.SurveyQuestion, error)
	GetAllQuestions(ctx context.Context) ([]*core.SurveyQuestion, error)
	CreateQuestion(ctx context.Context, q *core.SurveyQuestion) (*core.SurveyQuestion, error)
	UpdateQuestion(ctx context.Context, q *core.SurveyQuestion) error
	DeleteQuestion(ctx context.Context, id string) error
	CreateQuestionOption(ctx context.Context, option *core.SurveyQuestionOption) error
	UpdateQuestionOption(ctx context.Context, op *core.SurveyQuestionOption) error
	DeleteOption(ctx context.Context, id string) error
	GetAllSurveyResponses(ctx context.Context) ([]*core.SurveyResponse, error)
	CreateSurveyResponse(ctx context.Context, responses []core.SurveyResponse) error
}

type ISurveyService interface {
	GetSurvey(ctx context.Context) (*domain.Survey, error)
	CreateSurveyResponse(ctx context.Context, respones domain.SurveyResponse) error
}

type IAdminSurveyService interface {
	GetSurvey(ctx context.Context, id string) (*core.Survey, error)
	CreateSurvey(ctx context.Context, usr *core.Survey) error
	UpdateSurvey(ctx context.Context, usr *core.Survey) error
	DeleteSurvey(ctx context.Context, id string) error
}
