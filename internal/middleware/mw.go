// Package middleware provides a middleware type for http requests.
package middleware

import "net/http"

type Middleware func(http.Handler) http.HandlerFunc

func MiddlewareChain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.HandlerFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}

		return next.ServeHTTP
	}
}

type WrappedWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (w *WrappedWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.StatusCode = status
}
