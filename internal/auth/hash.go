// Package auth provides jwts and hashing.
package auth

import (
	"github.com/zrp9/launchl/internal/config"
	"golang.org/x/crypto/bcrypt"
)

func HashString(s string) (string, error) {
	cfg, err := loadCfg()
	if err != nil {
		return "", err
	}
	saltedNuts := cfg.Jwt.Salty + s
	bytes, err := bcrypt.GenerateFromPassword([]byte(saltedNuts), 14)
	return string(bytes), err
}

func VerifyHash(hash string, plainTxt string) (bool, error) {
	cfg, err := loadCfg()
	if err != nil {
		return false, err
	}

	saltedNuts := cfg.Jwt.Salty + plainTxt
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(saltedNuts))
	return err == nil, nil
}

func loadCfg() (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
