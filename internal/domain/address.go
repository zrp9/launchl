// Package domain is address entitiy
package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

type Address struct {
	bun.BaseModel `bun:"table:addresses,alias=adrs"`
	ID            int64     `bun:"column:id,pk,autoincrement" json:"id"`
	Address       string    `bun:"type:varchar(255),notnull,nullzero" json:"address" validate:"required,alphanumunicode,min=1,max=255"`
	City          string    `bun:"type:varchar(150),notnull,nullzero" json:"city" validate:"required,alpha,min=1,max=150"`
	State         string    `bun:"type:varchar(2),notnull,nullzero" json:"state" validate:"required,alpha,min=1,max=75"`
	Country       string    `bun:"type:varchar(75),notnull,nullzero" json:"country" validate:"required,alpha,min=1,max=75"`
	Zipcode       string    `bun:"type:varchar(12),notnull,nullzero" json:"zipcode" validate:"required,min=12,max=12"`
	Coordinates   Point     `bun:"type:point,notnull" json:"coordinates"`
	CreatedAt     time.Time `bun:"type:timestamptz,notnull,nullzero,default=current_timestamp"`
}

func NewAddress(addrs, city, state, country, zip string) *Address {
	return &Address{

		Address: addrs,
		City:    city,
		State:   state,
		Country: country,
		Zipcode: zip,
	}
}

func (a *Address) Value() (driver.Value, error) {
	adr, err := json.Marshal(a)
	return adr, err
}

func (a *Address) Scan(value interface{}) error {
	if value == nil {
		*a = Address{}
		return nil
	}

	adrs, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed for address")
	}

	err := json.Unmarshal(adrs, a)
	return err
}

// Point custom type implementing sql.valuer and sql.scanner so bun can use the custom type and save it as a point a postgres db type
type Point struct {
	X float64 `json:"x" validate:"longitude"` // lng
	Y float64 `json:"y" validate:"latitude"`  // lat
}

func (p *Point) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *Point) Scan(value interface{}) error {
	if value == nil {
		*p = Point{}
		return nil
	}

	s, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into Point", value)
	}

	// takes postgres point type (longitidue,latitude) and converts to point type Point.x and Point.y
	trimmed := strings.Trim(string(s), "()")
	parts := strings.Split(trimmed, ",")
	if len(parts) != 2 {
		return fmt.Errorf("invalid point format %s", s)
	}

	x, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return fmt.Errorf("invalid X (longitude) coordinate %v", err)
	}
	y, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return fmt.Errorf("invalid Y (latitude) coordinate %v", err)
	}

	p.X = x
	p.Y = y
	return nil
}
