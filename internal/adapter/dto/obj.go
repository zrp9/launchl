// Package dto provides data transfer objects for different domain types
package dto

import (
	"encoding/json"

	"github.com/google/uuid"
)

type UserCreateRequest struct {
	Email string `json:"email" validate:"email,required"`
	// Username is generated on usr type
	Phone       string `json:"phone" validate:"ascii,min=12,max=12"`
	FirstName   string `json:"firstName" validate:"required,max=150"`
	LastName    string `json:"lastName" validate:"required,max=150"`
	WouldUse    bool   `json:"wouldUse" validate:"boolean"`
	CompanyName string `json:"companyName" validate:"alphanum"`
}

type SurveyAnwser struct {
	QuestionID    string `json:"questionId" validate:"required,uuid4"`
	UserID        string `json:"userId" validate:"required,uuid4"`
	UserEmail     string `json:"userEmail" validate:"required,ascii"`
	OptionID      string `json:"optionId" validate:"required,uuid4"`
	WrittenAnwser string `json:"writtenAnwser" validate:"alphanum"`
}

type SurveyAnwsers struct {
	Answers   []SurveyAnwser
	UserEmail string `json:"userEmail" validate:"required,ascii"`
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
