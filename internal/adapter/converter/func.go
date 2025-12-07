// Package converter provides a converter interface and impls for converters to convert dtos to domain types and vice versa
package converter

import (
	"github.com/google/uuid"
	"github.com/zrp9/launchl/internal/adapter/dto"
	"github.com/zrp9/launchl/internal/domain/core"
)

type Converter interface {
	ToDTO(data any) (any, error)
	FromDTO(data any) (any, error)
}

func MakeSubscribeUser(data dto.UserSubscribeRequest) *core.User {
	return &core.User{
		Email:     data.Email,
		FirstName: data.FirstName,
		LastName:  data.LastName,
	}
}

func MakeCreateUser(data dto.UserCreateRequest) *core.User {
	return &core.User{
		Email:       data.Email,
		FirstName:   data.FirstName,
		LastName:    data.LastName,
		CompanyName: data.CompanyName,
	}
}

func ConvertSurveyAnwser(data dto.SurveyAnwser) (core.SurveyResponse, error) {
	id := uuid.New()
	qID, err := uuid.Parse(data.QuestionID)
	if err != nil {
		return core.SurveyResponse{}, err
	}

	if data.QuestionType != "text" {
		opID, err := uuid.Parse(data.OptionID)
		if err != nil {
			return core.SurveyResponse{}, err
		}

		return core.SurveyResponse{
			ID:         id,
			QuestionID: qID,
			OptionID:   opID,
		}, nil
	}

	return core.SurveyResponse{
		ID:              id,
		QuestionID:      qID,
		WrittenResponse: data.WrittenResponse,
	}, nil
}

func ConvertSurveyAnwsers(data dto.SurveyAnwsers) ([]core.SurveyResponse, core.User, error) {
	anwsers := make([]core.SurveyResponse, 0, len(data.Answers))
	for _, anwsDTO := range data.Answers {
		anws, err := ConvertSurveyAnwser(anwsDTO)
		if err != nil {
			return nil, core.User{}, err
		}
		anwsers = append(anwsers, anws)
	}

	user := core.User{
		FirstName:   data.FirstName,
		LastName:    data.LastName,
		Email:       data.UserEmail,
		CompanyName: data.CompanyName,
	}

	return anwsers, user, nil
}

func MakeSurveyResponse(data core.Survey) dto.SurveyResponse {
	response := dto.SurveyResponse{
		ID:      data.ID,
		Name:    data.Name,
		Version: data.Version,
	}

	questions := make([]dto.QuestionResponse, 0, len(data.Questions))
	for _, q := range data.Questions {
		options := make([]dto.OptionResponse, 0, len(q.Options))
		for _, o := range q.Options {
			options = append(options, dto.OptionResponse{
				ID:         o.ID,
				QuestionID: o.QuestionID,
				Position:   o.Position,
				Label:      o.Label,
				Value:      o.Value,
			})
		}
		questions = append(questions, dto.QuestionResponse{
			ID:           q.ID,
			QuestionType: string(q.QuestionType),
			Prompt:       q.Prompt,
			Position:     q.Position,
			Required:     q.Required,
			MetaData:     q.MetaData,
			Options:      options,
		})
	}

	response.Questions = questions
	return response
}
