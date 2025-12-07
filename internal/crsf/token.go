// Package crsf provides crsf tokens for http request
package crsf

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/zrp9/launchl/internal/middleware"
	req "github.com/zrp9/launchl/internal/request"
)

const (
	crsfToken  = "crsf_token"
	crsfHeader = "X-CRSF_Token"
)

func randomToken(n int) (string, error) {
	b := make([]byte, 0, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func EnsureCRSFCookie(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(crsfToken); err == nil && c.Value != "" {
		return
	}

	token, err := randomToken(32)
	if err != nil {
		log.Printf("could not create crsf token %v", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     crsfToken,
		Value:    token,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
		HttpOnly: false,
	})
}

func ApplyCRSFToken(isDev bool) middleware.Middleware {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log.Println("apply crfs middleware...")
			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions:
				if c, err := r.Cookie(crsfToken); err != nil || c.Value == "" {
					token, err := randomToken(32)
					log.Printf("token %v", token)
					if err != nil {
						log.Printf("failed to create CRSF token %v", err)
					}

					if token != "" && err == nil {
						log.Println("no token header setting...")
						http.SetCookie(w, &http.Cookie{
							Name:     crsfToken,
							Value:    token,
							Path:     "/",
							SameSite: http.SameSiteLaxMode,
							Secure:   isDev,
							HttpOnly: false,
						})
					}

				}
			}
			next.ServeHTTP(w, r)
		}
	}
}

func ValidateCRSFToken(isDev bool, paths ...string) middleware.Middleware {
	// protected := make(map[string]struct{}, len(paths))
	// for _, p := range paths {
	// 	protected[p] = struct{}{}
	// }

	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log.Println("validate crfs middleware...")
			isWrite := isWriteMethod(r)
			//_, protectedPath := protected[r.URL.Path]
			var protectedPath = false
			for _, path := range paths {
				if strings.Contains(r.URL.Path, path) {
					protectedPath = true
				}
			}

			if !isWrite || !protectedPath {
				log.Println("not a write method skip...")
				next.ServeHTTP(w, r)
				return
			}

			cookie, err := r.Cookie(crsfToken)
			if err != nil || cookie.Value == "" {
				log.Printf("error getting cookie %v", err)
				req.WriteErr(w, http.StatusForbidden, errors.New("missing CRSF token"))
				return
			}

			header := r.Header.Get(crsfHeader)
			log.Printf("csrf header %v", header)
			if header == "" {
				log.Println("no header err")
				req.WriteErr(w, http.StatusForbidden, errors.New("missing CRSF header"))
				return
			}

			if subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(header)) != 1 {
				log.Println("constant time compare fail")
				req.WriteErr(w, http.StatusForbidden, errors.New("invalid CRSF token"))
				return
			}

			if origin := r.Header.Get("Origin"); origin != "" && origin != "https://zrp3.dev" {
				log.Println("invalid origin err")
				req.WriteErr(w, http.StatusForbidden, errors.New("invalid origin"))
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

func isWriteMethod(r *http.Request) bool {
	return r.Method == http.MethodPost ||
		r.Method == http.MethodPut ||
		r.Method == http.MethodPatch ||
		r.Method == http.MethodDelete
}
