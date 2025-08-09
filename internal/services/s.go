// Package services defines service interface and other functionality
package services

import "net/http"

type Service interface {
	RegisterRoutes(mr *http.ServeMux)
}
