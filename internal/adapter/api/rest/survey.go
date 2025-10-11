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
			s.logger.MustError(err)
		}
	}
}

func (s SurveyHandler) RegisterRoutes(m *http.ServeMux) {
	log.Printf("MUX in survey hndler register:  %p", m)
	m.HandleFunc("GET /api/v1/survey", s.HandleLogging(s.HandleGetSurvey))
	m.HandleFunc("POST /api/v1/survey/{username}", s.HandleLogging(s.HandleSurveyResponse))
}

func (s SurveyHandler) HandleGetSurvey(w http.ResponseWriter, r *http.Request) error {
	if err := r.Context().Err(); err != nil {
		return APIErr{Status: http.StatusRequestTimeout, Err: err}
	}

	survey, err := s.s.GetSurvey(r.Context())
	if err != nil {
		return APIErr{Status: http.StatusRequestTimeout, Err: err}
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
		errs := make([]string, 0)
		for _, e := range err.(v.ValidationErrors) {
			errs = append(errs, fmt.Sprintf("field %v failed because %v", e.StructField(), e.Error()))
		}
		return APIErr{Status: http.StatusBadRequest, Err: errors.New(strings.Join(errs, ","))}
	}

	sResponse, err := converter.ConvertSurveyAnwsers(payload)
	if err != nil {
		return APIErr{Status: http.StatusBadRequest, Err: err}
	}

	if err := s.s.CreateSurveyResponse(r.Context(), sResponse); err != nil {
		s.logger.MustError(err)
	}

	res := request.JSON{
		"status": 201,
	}

	return request.WriteJSON(w, http.StatusCreated, res)
}
