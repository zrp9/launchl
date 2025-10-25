package ports

import (
	"context"

	"github.com/zrp9/launchl/internal/domain/core"
)

type ISurveyRepo interface {
	GetSurvey(ctx context.Context, id string) (*core.Survey, error)
	GetActiveSurvey(ctx context.Context) (*core.Survey, error)
	CreateSurvey(ctx context.Context, usr *core.Survey) (*core.Survey, error)
	UpdateSurvey(ctx context.Context, usr *core.Survey) error
	DeleteSurvey(ctx context.Context, id string) error
	GetQuestion(ctx context.Context, id string) (*core.Question, error)
	GetAllQuestions(ctx context.Context) ([]*core.Question, error)
	CreateQuestion(ctx context.Context, q *core.Question) (*core.Question, error)
	UpdateQuestion(ctx context.Context, q *core.Question) error
	DeleteQuestion(ctx context.Context, id string) error
	CreateQuestionOption(ctx context.Context, option *core.SurveyQuestionOption) error
	UpdateQuestionOption(ctx context.Context, op *core.SurveyQuestionOption) error
	DeleteOption(ctx context.Context, id string) error
	GetAllSurveyResponses(ctx context.Context) ([]*core.SurveyResponse, error)
	CreateSurveyResponse(ctx context.Context, responses []core.SurveyResponse) error
}

type ISurveyService interface {
	GetSurvey(ctx context.Context) (*core.Survey, error)
	CreateSurveyResponse(ctx context.Context, respones core.SurveyResponse) error
}

type IAdminSurveyService interface {
	GetSurvey(ctx context.Context, id string) (*core.Survey, error)
	CreateSurvey(ctx context.Context, usr *core.Survey) error
	UpdateSurvey(ctx context.Context, usr *core.Survey) error
	DeleteSurvey(ctx context.Context, id string) error
}
