// Package dto provides data transfer objects for different domain types
package dto

import (
	"encoding/json"

	"github.com/google/uuid"
)

type UserSubscribeRequest struct {
	Email     string `json:"email" validate:"email,required"`
	FirstName string `json:"firstName" validate:"required,max=150"`
	LastName  string `json:"lastName" validate:"required,max=150"`
}

type UserCreateRequest struct {
	Email     string `json:"email" validate:"email,required"`
	FirstName string `json:"firstName" validate:"required,max=150"`
	LastName  string `json:"lastName" validate:"required,max=150"`
	// Username is generated on usr type
	Phone       string `json:"phone" validate:"ascii,min=12,max=12"`
	WouldUse    bool   `json:"wouldUse" validate:"boolean"`
	CompanyName string `json:"companyName" validate:"alphanum"`
}

type SurveyAnwser struct {
	QuestionID      string `json:"questionId" validate:"required,uuid4"`
	QuestionType    string `json:"questionType" validate:"oneof='text' 'multi-check' 'check' 'drop-down'"`
	OptionID        string `json:"optionId" `
	OptionValue     string `json:"optionValue"`
	Prompt          string `json:"prompt" validate:"ascii"`
	WrittenResponse string `json:"writtenResponse" validate:"alphanum"`
}

type SurveyAnwsers struct {
	Answers     []SurveyAnwser `json:"anwsers" validate:"required"`
	UserEmail   string         `json:"userEmail" validate:"required,ascii"`
	CompanyName string         `json:"companyName" validate:"required,ascii"`
	FirstName   string         `json:"firstName" validate:"alphanum"`
	LastName    string         `json:"lastName" validate:"required,alphanum"`
}

type SurveyResponse struct {
	ID        uuid.UUID          `json:"id"`
	Name      string             `json:"name"`
	Version   string             `json:"version"`
	Questions []QuestionResponse `json:"questions"`
}

type QuestionResponse struct {
	ID           uuid.UUID        `json:"id"`
	QuestionType string           `json:"questionType"`
	Prompt       string           `json:"prompt"`
	Position     int              `json:"position"`
	Required     bool             `json:"required"`
	MetaData     json.RawMessage  `json:"metadata"`
	Options      []OptionResponse `json:"options"`
}

type OptionResponse struct {
	ID         uuid.UUID `json:"id"`
	QuestionID uuid.UUID `json:"questionId"`
	Position   int       `json:"position"`
	Label      string    `json:"label"`
	Value      string    `json:"value"`
}
