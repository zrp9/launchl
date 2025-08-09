package domain

import (
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

	ID        uuid.UUID        `bun:"pk,type:uuid" json:"id" validate:"uuidv4"`
	Name      string           `bun:"type:varchar(255),notnull,nullzero" json:"name" validate:"alphanum"`
	Active    bool             `bun:"type:boolean,notnull,nullzero,default=false" json:"active" validate:"boolean"`
	Questions []SurveyQuestion `bun:"rel:has-many,join:id=survey_id" json:"questions"`
	CreatedAt time.Time        `bun:"type:timestamptz,notnull,nullzero,default=current_timestamp" json:"createdAt"`
	UpdatedAt time.Time        `bun:"type:timestamptz,notnull,nullzero,default=current_timestamp" json:"updatedAt"`
}

type SurveyQuestion struct {
	bun.BaseModel `bun:"table:survey_questions,alias:sq"`

	ID           uuid.UUID    `bun:"pk,type:uuid" json:"id" validate:"uuidv4"`
	SurveyID     uuid.UUID    `bun:"type:uuid,notnull" json:"surveyId" validate:"uuidv4"`
	QuestionType QuestionType `bun:"type:question_type,notnull,nullzero,default='check'" json:"questionType" validate:"oneof='check' 'multi-check' 'drop-down' 'text'"`
	Options      Options      `bun:"type:jsonb,notnull,nullzero" json:"options"`
	Active       bool         `bun:"type:boolean,notnull,nullzero,default=false" json:"active" validate:"boolean"`
	CreatedAt    time.Time    `bun:"type:timestamptz,notnull,nullzero,default=current_timestamp" json:"createdAt"`
	UpdatedAt    time.Time    `bun:"type:timestamptz,notnull,nullzero,default=current_timestamp" json:"updatedAt"`
}

type UserSurvey struct {
	bun.BaseModel `bun:"table:user_surveys,alias:us"`
	SurveyID      uuid.UUID `bun:",pk,type:uuid4" json:"surveyId" validate:"uuidv4"`
	Survey        *Survey   `bun:"rel:belongs-to,join:survey_id=id" json:"survey"`
	UserID        uuid.UUID `bun:",pk,type:uuid4" json:"userId" validate:"uuid4"`
	User          *User     `bun:"rel:belongs-to,join:user_id=id" json:"user"`
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
