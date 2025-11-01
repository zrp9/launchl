package service

import (
	"context"
	"encoding/json"

	"github.com/zrp9/launchl/internal/adapter/cache/valkaree"
	"github.com/zrp9/launchl/internal/adapter/log/crane"
	"github.com/zrp9/launchl/internal/adapter/repo/pgsql"
	"github.com/zrp9/launchl/internal/domain/core"
)

const (
	surveyKey = "x-serv"
)

type SurveyService struct {
	repo   pgsql.SurveyRepo
	cache  valkaree.Cache
	logger crane.Zlogrus
}

func NewSurveyService(r pgsql.SurveyRepo, c valkaree.Cache, l crane.Zlogrus) SurveyService {
	return SurveyService{
		repo:   r,
		cache:  c,
		logger: l,
	}
}

func (s SurveyService) Name() string {
	return "survey"
}

func (s SurveyService) GetSurvey(ctx context.Context) (*core.Survey, error) {
	sCache, err := s.cache.Get(ctx, surveyKey)
	if err != nil && err != valkaree.ErrEmptyCache {
		s.logger.MustTraceErr(err)
	}

	var survey *core.Survey
	if sCache != "" {
		err = json.Unmarshal([]byte(sCache), &survey)
		if err != nil {
			s.logger.MustTraceErr(err)
		}

		if survey != nil {
			return survey, nil
		}
	}

	if survey, err = s.repo.GetActiveSurvey(ctx); err != nil {
		return nil, err
	}

	return survey, nil
}

func (s SurveyService) CreateSurveyResponse(ctx context.Context, responses []core.SurveyResponse) error {
	return s.repo.CreateSurveyResponse(ctx, responses)
}
