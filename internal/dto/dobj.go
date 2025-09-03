// Package dto provides data transfer objects
package dto

import (
	"errors"
	"mime/multipart"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Transferer interface {
	Validate() error
}

type DTO struct {
}

type SurveyResponse struct {
	OptionID   uuid.UUID `json:"optionId" validate:"uuidv4,required"`
	UserID     uuid.UUID `json:"userId" validate:"uuidv4,required"`
	QuestionID uuid.UUID `json:"questionId" validate:"uuidv4,required"`
	TextAnwser string    `json:"textAnwser,omitempty" validate:"alphanum"`
}

func (s SurveyResponse) Validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	return v.Struct(s)
}

type SurveyResponses []SurveyResponse

func (s SurveyResponses) Validate() []error {
	errs := make([]error, 0)
	for _, res := range s {
		if err := res.Validate(); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

type FileUploadDto struct {
	File    multipart.File
	FileKey string
	Header  *multipart.FileHeader
}

func (f FileUploadDto) Validate() error {
	if f.File == nil {
		return errors.New("error file is required")
	}

	if f.FileKey == "" {
		return errors.New("error file key is empty")
	}

	if f.Header == nil {
		return errors.New("error file header is required")
	}

	return nil
}

type LoginDto struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type SignupDto struct {
	Username  string `json:"username" validate:"required,alphanum,min=1,max=75"`
	Password  string `json:"password" validate:"required,min=1,max=255"`
	Email     string `json:"email" validate:"required,alphanum,min=1,max=150"`
	FirstName string `json:"firstName" validate:"required,alpha,min=1,max=100"`
	LastName  string `json:"lastName" validate:"required,alpha,min=1,max=100"`
}
