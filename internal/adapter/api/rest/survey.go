// Package rest provides http handlers for services
package rest

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	v "github.com/go-playground/validator/v10"
	"github.com/zrp9/launchl/internal/adapter/converter"
	"github.com/zrp9/launchl/internal/adapter/dto"
	"github.com/zrp9/launchl/internal/adapter/log/crane"
	"github.com/zrp9/launchl/internal/adapter/repo/pgsql"
	"github.com/zrp9/launchl/internal/domain/service"
	"github.com/zrp9/launchl/internal/request"
)

type SurveyHandler struct {
	s         service.SurveyService
	logger    crane.Zlogrus
	validator v.Validate
}

func NewSurveyHandler(serv service.SurveyService, v v.Validate, l crane.Zlogrus) SurveyHandler {
	return SurveyHandler{
		s:         serv,
		logger:    l,
		validator: v,
	}
}

func (s SurveyHandler) HandleLogging(hn APIHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := hn(w, r); err != nil {
			if e, ok := err.(APIErr); ok {
				request.WriteErr(w, e.Status, e)
			}
			log.Printf("requet failed err %v", err)
			s.logger.MustError(err)
		}
	}
}

func (s SurveyHandler) RegisterRoutes(m *http.ServeMux) {
	log.Printf("MUX in survey hndler register:  %p", m)
	m.HandleFunc("GET /api/v1/survey", s.HandleLogging(s.HandleGetSurvey))
	m.HandleFunc("POST /api/v1/survey/respond", s.HandleLogging(s.HandleSurveyResponse))
}

func (s SurveyHandler) HandleGetSurvey(w http.ResponseWriter, r *http.Request) error {
	log.Printf("hit get survey api")
	if err := r.Context().Err(); err != nil {
		return APIErr{Status: http.StatusRequestTimeout, Err: err}
	}

	survey, err := s.s.GetSurvey(r.Context())
	if err != nil {
		if err == pgsql.ErrNoRecords {
			return request.WriteJSON(w, http.StatusOK, request.JSON{"survey": ""})
		}
		return APIErr{Status: http.StatusRequestTimeout, Err: err}
	}

	if survey == nil {
		return request.WriteJSON(w, http.StatusOK, request.JSON{
			"survey": nil,
		})
	}

	response := converter.MakeSurveyResponse(*survey)
	res := request.JSON{
		"survey": response,
	}

	return request.WriteJSON(w, http.StatusOK, res)
}

func (s SurveyHandler) HandleSurveyResponse(w http.ResponseWriter, r *http.Request) error {
	if err := r.Context().Err(); err != nil {
		return APIErr{Status: http.StatusRequestTimeout, Err: err}
	}

	var payload dto.SurveyAnwsers
	if err := request.ParseJSON(r, &payload); err != nil {
		return APIErr{Status: http.StatusBadRequest, Err: err}
	}

	if err := s.validator.Struct(payload); err != nil {
		log.Println("failed validation")
		errs := make([]string, 0)
		for _, e := range err.(v.ValidationErrors) {
			errs = append(errs, fmt.Sprintf("field %v failed because %v", e.StructField(), e.Error()))
		}
		return APIErr{Status: http.StatusBadRequest, Err: errors.New(strings.Join(errs, ","))}
	}

	responses, usr, err := converter.ConvertSurveyAnwsers(payload)
	if err != nil {
		log.Printf("failed conversion %v", err)
		return APIErr{Status: http.StatusBadRequest, Err: err}
	}
	log.Printf("responses %v", responses)

	if err := s.s.CreateSurveyResponse(r.Context(), usr, responses); err != nil {
		log.Printf("failed create %v", err)
		s.logger.MustError(err)
	}

	res := request.JSON{
		"status": 201,
	}

	return request.WriteJSON(w, http.StatusCreated, res)
}
