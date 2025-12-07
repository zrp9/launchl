// Package core survey contains survey domain
package core

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type QuestionType string

const (
	CHECK      QuestionType = "check"
	MULTICHECK QuestionType = "multi-check"
	DROPDOWN   QuestionType = "drop-down"
	TEXT       QuestionType = "text"
)

type Survey struct {
	bun.BaseModel `bun:"table:surveys,alias:s"`
	ID            uuid.UUID  `bun:"id,pk,type:uuid" json:"id" validate:"uuidv4"`
	Questions     []Question `bun:"rel:has-many,join:id=survey_id" json:"questions"`
	Version       string     `bun:"type:varchar(75),notnull,nullzero" json:"version" validate:"numeric"`
	Name          string     `bun:"type:varchar(255),notnull,nullzero" json:"name" validate:"alphanum"`
	Intro         string     `bun:"type:varchar(255),notnull,nullzero" json:"intro" validate:"alphanum"`
	ClosingText   string     `bun:"type:varchar(255),notnull,nullzero" json:"closingText"`
	Active        bool       `bun:"type:boolean,notnull,nullzero,default:false" json:"active" validate:"boolean"`
	CreatedAt     time.Time  `bun:"type:timestamptz,notnull,nullzero,default:current_timestamp" json:"createdAt"`
	UpdatedAt     time.Time  `bun:"type:timestamptz,notnull,nullzero,default:current_timestamp" json:"updatedAt"`
}

type Question struct {
	bun.BaseModel `bun:"table:survey_questions,alias:sq"`

	ID           uuid.UUID              `bun:"id,pk,type:uuid" json:"id" validate:"uuidv4"`
	SurveyID     uuid.UUID              `bun:"type:uuid,notnull" json:"surveyId" validate:"uuidv4"`
	QuestionType QuestionType           `bun:"type:question_type,notnull,nullzero,default:'check'" json:"questionType" validate:"oneof='check' 'multi-check' 'drop-down' 'text'"`
	Options      []SurveyQuestionOption `bun:"rel:has-many,join:id=question_id" json:"options"`
	Prompt       string                 `bun:"type:text,notnull,nullzero" json:"prompt" validate:"alphanum"`
	Position     int                    `bun:"type:integer,notnull,nullzero,default:0" json:"position" validate:"numeric"`
	Active       bool                   `bun:"type:boolean,notnull,nullzero,default:false" json:"active" validate:"boolean"`
	Required     bool                   `bun:"type:boolean,notnull,nullzero,default:true" json:"required" validate:"boolean"`
	MetaData     json.RawMessage        `bun:"type:jsonb,nullzero" json:"metaData" validate:"json"`
	Name         string                 `bun:"type:varchar(255),notnull,nullzero" json:"name" validate:"alphanum"`
	CreatedAt    time.Time              `bun:"type:timestamptz,notnull,nullzero,default:current_timestamp" json:"createdAt"`
	UpdatedAt    time.Time              `bun:"type:timestamptz,notnull,nullzero,default:current_timestamp" json:"updatedAt"`
}

type SurveyQuestionOption struct {
	bun.BaseModel `bun:"table:survey_question_options,alias:sqo"`
	ID            uuid.UUID `bun:"id,pk,type:uuid,nullzero" json:"id" validate:"uuidv4"`
	QuestionID    uuid.UUID `bun:"type:uuid,nullzero" json:"questionId" validate:"uuidv4"`
	Position      int       `bun:"type:integer,notnull,nullzero,default:0" json:"position" validate:"numeric"`
	Label         string    `bun:"type:varchar(255),notnull,nullzero" json:"label" validate:"alphanum"`
	// value can be empty to support text responses
	Value     string    `bun:"type:varchar(255),notnull,nullzero" json:"value" validate:"aplhanum"`
	CreatedAt time.Time `bun:"type:timestamptz,notnull,nullzero,default:current_timestamp" json:"createdAt"`
	UpdatedAt time.Time `bun:"type:timestamptz,notnull,nullzero,default:current_timestamp" json:"updatedAt"`
}

type SurveyResponse struct {
	bun.BaseModel  `bun:"table:user_survey_responses,alias:sur"`
	ID             uuid.UUID             `bun:"id,pk,type:uuid,nullzero" json:"id" validate:"uuidv4"`
	QuestionID     uuid.UUID             `bun:"type:uuid,nullzero" json:"questionId" validate:"uuidv4"`
	UserID         uuid.UUID             `bun:"type:uuid,nullzero" json:"userId" validate:"uuidv4"`
	OptionID       uuid.UUID             `bun:"type:uuid,nullzero" json:"optionId,omitempty" validate:"uuidv4"`
	Question       *Question             `bun:"rel:belongs-to,join:question_id=id" json:"question,omitempty"`
	User           *User                 `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
	QuestionOption *SurveyQuestionOption `bun:"rel:belongs-to,join:option_id=id" json:"questionOption,omitempty"`
	// WrittenResponse holds the response to text questions
	WrittenResponse string    `bun:"type:text,nullzero" json:"writtenResponse,omitempty" validate:"alphanum"`
	CreatedAt       time.Time `bun:"type:timestamptz,notnull,nullzero,default:current_timestamp" json:"createdAt"`
	UpdatedAt       time.Time `bun:"type:timestamptz,notnull,nullzero,default:current_timestamp" json:"updatedAt"`
}

// TODO: add the object like point for options for db scanning/inserting

type Option struct {
	Key   string `json:"key" validate:"alpha"`
	Value string `json:"value" validate:"alphanum"`
}

type Options []Option

func (o Options) Value() (driver.Value, error) {
	return json.Marshal(o)
}

func (o *Options) Scan(src any) error {
	if src == nil {
		*o = nil
		return nil
	}

	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, o)
	case string:
		return json.Unmarshal([]byte(v), o)
	default:
		return fmt.Errorf("unsupported src type %T for option", src)

	}
}

var _ bun.BeforeAppendModelHook = (*SurveyResponse)(nil)

func (m *SurveyResponse) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		m.CreatedAt = time.Now()
	case *bun.UpdateQuery:
		m.UpdatedAt = time.Now()
	}
	return nil
}
