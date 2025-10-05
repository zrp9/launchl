// Package validator contains valdations from go-playground
package validator

import (
	"strings"
)

const MessageTagKey = "message"

type ValidError struct {
	Key     string
	Message string
}

type ValidErrors []*ValidError

func (v *ValidError) Error() string {
	return v.Message
}

func (v *ValidErrors) Errors() []string {
	errs := make([]string, 0)
	for _, err := range *v {
		errs = append(errs, err.Error())
	}
	return errs
}

func (v *ValidErrors) Error() string {
	return strings.Join(v.Errors(), ",")
}
