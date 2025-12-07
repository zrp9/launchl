package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/zrp9/launchl/internal/adapter/cache/valkaree"
	"github.com/zrp9/launchl/internal/adapter/log/crane"
	"github.com/zrp9/launchl/internal/adapter/noti"
	"github.com/zrp9/launchl/internal/adapter/repo/pgsql"
	"github.com/zrp9/launchl/internal/domain/core"
	"github.com/zrp9/launchl/internal/dto"
)

const (
	surveyKey = "ll-survey"
	notiSrc   = "survey-service"
)

type SurveyService struct {
	repo   pgsql.SurveyRepo
	cache  valkaree.Cache
	stream valkaree.StreamWriter
	logger crane.Zlogrus
}

func NewSurveyService(r pgsql.SurveyRepo, c valkaree.Cache, stream valkaree.Stream, l crane.Zlogrus) SurveyService {
	return SurveyService{
		repo:   r,
		cache:  c,
		stream: stream.Writer(),
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
			log.Printf("using survey %v", survey)
			return survey, nil
		}
	}

	if survey, err = s.repo.GetActiveSurvey(ctx); err != nil {
		return nil, err
	}

	return survey, nil
}

func (s SurveyService) CreateSurveyResponse(ctx context.Context, usr core.User, responses []core.SurveyResponse) error {
	err := s.repo.CreateSurveyResponse(ctx, &usr, responses)
	if err != nil {
		return err
	}

	// TODO: I dont think the email dto doesn't need a subject because that will be determined by template?
	go func() {
		log.Println("routine to send email...")
		nCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		data, err := dto.CreateEmailPayload(
			usr.Email,
			"Thank you For Taking Our Survey",
			"surveyThanks",
			streamCfg.SenderCfg.TemplateVersion,
			noti.SurveyThanksData{
				Name:        fmt.Sprintf("%v %v", usr.FirstName, usr.LastName),
				ReferralURL: usr.ReferalURL,
			},
		)

		log.Printf("sending email %v", data)
		if err != nil {
			s.logger.MustTraceErr(fmt.Errorf("failed to create email payload for survey notification %v", err))
			log.Printf("failed to create email payload for survey notification %v", err)
		}

		if _, err := s.stream.WriteEvent(nCtx, valkaree.Event{EventType: "survey", Target: streamCfg.StreamCfg.Group, Src: notiSrc}, data); err != nil {
			s.logger.MustTraceErr(fmt.Errorf("failed to write email stream event %v", err))
			log.Printf("failed to write email stream event %v", err)
		}

		log.Println("finished notification routine...")
	}()

	return nil
}
