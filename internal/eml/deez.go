// Package eml contains functions for working with emails
package eml

import "strings"

func StripDomain(email string) string {
	return strings.Split(email, "@")[0]
}
