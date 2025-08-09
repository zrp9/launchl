// Package api has server config, common middleware and base route definitions
package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zrp9/launchl/internal/auth"
	"github.com/zrp9/launchl/internal/config"
	"github.com/zrp9/launchl/internal/crane"
	"github.com/zrp9/launchl/internal/middleware"
	"github.com/zrp9/launchl/internal/request"
	"github.com/zrp9/launchl/internal/services"
)

func NewServer(cfg config.ServerCfg, apis []services.Service) *http.Server {
	mux := http.NewServeMux()
	mwChain := middleware.MiddlewareChain(handlePanic, loggerMiddleware, headerMiddleware, contextMiddleware)
	registerRoutes(mux, apis)
	server := &http.Server{
		Addr:         cfg.Host,
		ReadTimeout:  time.Second * time.Duration(cfg.ReadTimeout),
		WriteTimeout: time.Second * time.Duration(cfg.WriteTimeout),
		Handler:      mwChain(mux),
	}
	return server
}

func registerRoutes(mux *http.ServeMux, apis []services.Service) {
	for _, api := range apis {
		api.RegisterRoutes(mux)
	}
}

// TODO: update these middlewares to "better paradigms or what you've learned"

func Authenticate(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cook, err := r.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				request.WriteErr(w, http.StatusUnauthorized, err)
				return
			}
			request.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		token := cook.Value
		claims := &auth.UserClaims{}
		tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return auth.GetJwtKey()
		})

		if err != nil || !tkn.Valid {
			request.WriteErr(w, http.StatusUnauthorized, errors.New("unathorized"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func headerMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		crane.DefaultLogger.MustDebug(fmt.Sprintf("orign: %v", origin))
		// log.Printf("🔍 Incoming request: %s %s", r.Method, r.URL.Path)
		// log.Printf("Headers: %v", r.Header)

		if isValidOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		//w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func contextMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeout := 10 * time.Minute
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func loggerMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &middleware.WrappedWriter{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}
		next.ServeHTTP(wrapped, r)
		crane.DefaultLogger.MustDebug(fmt.Sprintf("Method: %s, URI: %s, IP: %s, Duration: %v, Status: %v", r.Method, r.RequestURI, r.RemoteAddr, start, wrapped.StatusCode))
	})
}

func HandleShutdown(server *http.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	crane.DefaultLogger.MustDebug("Shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		crane.DefaultLogger.MustFatal(fmt.Sprintf("Server forced shutdown: %v", err))
	}
	crane.DefaultLogger.MustDebug("Server exited")
}

func handlePanic(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				crane.DefaultLogger.MustDebug(fmt.Sprintf("panic recovered %v", rec))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func isValidOrigin(origin string) bool {
	return true
	//	validOrigin := map[string]bool{
	//		"http://localhost:3000":      true,
	//		"https://localhost:3000":     true,
	//		"http://zrp3.dev":      true,
	//		"https://zrp3.dev":     true,
	//		"http://www.zrp3.dev":  true,
	//		"https://www.zrp3.dev": true,
	//		"www.zrp3.dev":         true,
	//	}
	//
	// return validOrigin[origin]
}
