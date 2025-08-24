// Package config holds methods for working with config file
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env        string
	Server     ServerCfg
	Database   DatabaseCfg
	AWS        AWSCfg
	OpenSearch OpenSearchCfg
	Valkey     ValkeyCfg
	Jwt        JWTCfg
}

type ServerCfg struct {
	Port         string
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseCfg struct {
	Provider        string
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	SSLRoot         string
	DialTimeout     time.Duration
	WriteTimeout    time.Duration
	ReadTimeout     time.Duration
	MaxConns        int
	MaxOpenConns    int
	MaxIdleConns    int
	ConnTimeout     time.Duration
	MigrationsTable string
	MigrationsLock  string
}

type AWSCfg struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	S3Bucket        string
	SQSQueueURL     string
}

type OpenSearchCfg struct {
	Host     string
	Port     string
	Username string
	Password string
	Index    string
	UseSSL   bool
}

type ValkeyCfg struct {
	Host string
	Port string
}

type EmailCfg struct {
	Sender          string
	Attempts        int64
	TemplateVersion int
}

type JWTCfg struct {
	Secret     string
	Expiration time.Duration
	Salty      string
}

func Load() (*Config, error) {
	// appEnv := getEnv("ENV", "dev")
	// if err := godotenv.Load(fmt.Sprintf(".env.%s", appEnv)); err != nil {
	// 	log.Printf("%s env does not exist", appEnv)
	// }

	// if err := godotenv.Load(".env"); err != nil {
	// 	log.Printf("warning could not load .env file %v", err)
	// }
	appEnv := initializeEnv()

	return &Config{
		Env: appEnv,
		Server: ServerCfg{
			//change port dflt to 443 for prod
			Port:         getEnv("PORT", "8090"),
			Host:         getEnv("HOST", "0.0.0.0"),
			ReadTimeout:  getDurationEnv("READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 10*time.Second),
		},
		Database: DatabaseCfg{
			Provider:        mustGetEnv("DB_PROVIDER"),
			Host:            mustGetEnv("DB_HOST"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            mustGetEnv("DB_USER"),
			Password:        mustGetEnv("DB_PASSWORD"),
			Name:            mustGetEnv("DB_NAME"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			SSLRoot:         getEnv("DB_SSL_ROOT", "none"),
			MaxConns:        getIntEnv("DB_MAX_CONNS", 20),
			MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 20),
			MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 20),
			DialTimeout:     getDurationEnv("DB_DIAL_TIMEOUT", 15),
			WriteTimeout:    getDurationEnv("DB_WRITE_TIMEOUT", 15),
			ReadTimeout:     getDurationEnv("DB_READ_TIMEOUT", 15),
			ConnTimeout:     getDurationEnv("DB_CONN_TIMEOUT", 15),
			MigrationsTable: getEnv("DB_MIGRATIONS_TABLE", "bun_migrations"),
			MigrationsLock:  getEnv("DB_MIGRATIONS_LOCK", "bun_migration_lock"),
		},
		AWS: AWSCfg{
			Region:          getEnv("AWS_REGION", "us-east-2"),
			AccessKeyID:     mustGetEnv("AWS_ACCESS_KEY_ID"),
			SecretAccessKey: mustGetEnv("AWS_SECRET_ACCESS_KEY"),
			S3Bucket:        mustGetEnv("AWS_S3_BUCKET"),
			SQSQueueURL:     getEnv("AWS_SQS_QUEUE_URL", ""),
		},
		OpenSearch: OpenSearchCfg{
			Host:     mustGetEnv("OPENSEARCH_HOST"),
			Port:     getEnv("OPENSEARCH_PORT", "9200"),
			Username: getEnv("OPENSEARCH_USERNAME", ""),
			Password: getEnv("OPENSEARCH_PASSWORD", ""),
			Index:    getEnv("OPENSEARCH_INDEX", "documents"),
			UseSSL:   getBoolEnv("OPENSEARCH_USE_SSL", true),
		},
		Valkey: ValkeyCfg{
			Host: getEnv("VALKEY_HOST", "localhost"),
			Port: getEnv("VALKEY_PORT", "6379"),
		},
		Jwt: JWTCfg{
			Secret:     mustGetEnv("JWT_SECRET"),
			Expiration: getDurationEnv("JWT_EXPIRATION", 24*time.Hour),
			Salty:      mustGetEnv("SALTY"),
		},
	}, nil
}

func LoadDBConfig() (DatabaseCfg, error) {
	_ = initializeEnv()

	return DatabaseCfg{
		Provider:        mustGetEnv("DB_PROVIDER"),
		Host:            mustGetEnv("DB_HOST"),
		Port:            getEnv("DB_PORT", "5432"),
		User:            mustGetEnv("DB_USER"),
		Password:        mustGetEnv("DB_PASSWORD"),
		Name:            mustGetEnv("DB_NAME"),
		SSLMode:         getEnv("DB_SSLMODE", "disable"),
		SSLRoot:         getEnv("DB_SSL_ROOT", "none"),
		MaxConns:        getIntEnv("DB_MAX_CONNS", 20),
		MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 20),
		MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 20),
		DialTimeout:     getDurationEnv("DB_DIAL_TIMEOUT", 15),
		WriteTimeout:    getDurationEnv("DB_WRITE_TIMEOUT", 15),
		ReadTimeout:     getDurationEnv("DB_READ_TIMEOUT", 15),
		ConnTimeout:     getDurationEnv("DB_CONN_TIMEOUT", 15),
		MigrationsTable: getEnv("DB_MIGRATIONS_TABLE", "bun_migrations"),
		MigrationsLock:  getEnv("DB_MIGRATIONS_LOCK", "bun_migration_lock"),
	}, nil
}

func LoadValkey() ValkeyCfg {
	_ = initializeEnv()
	return ValkeyCfg{
		Host: mustGetEnv("VALKEY_HOST"),
		Port: mustGetEnv("VALKEY_HOST"),
	}
}

func LoadEmailConfig() EmailCfg {
	_ = initializeEnv()
	return EmailCfg{
		Sender:          mustGetEnv("EMAIL_SENDER"),
		Attempts:        getInt64Env("EMAIL_ATTEMPTS", 1),
		TemplateVersion: getIntEnv("EMAIL_TEMPLATE_VERSION", 1),
	}
}

func GetAuthToken() ([]byte, error) {
	_ = initializeEnv()
	authKey := mustGetEnv("AUTH_KEY")
	return []byte(authKey), nil
}

func initializeEnv() string {
	appEnv := getEnv("ENV", "dev")
	if err := godotenv.Load(fmt.Sprintf(".env.%s", appEnv)); err != nil {
		log.Printf("%s env does not exist", appEnv)
	}

	if err := godotenv.Load(".env"); err != nil {
		log.Printf("warning could not load .env file %v", err)
	}

	return appEnv
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("environment variable %s is required", key))
	}
	return value
}

func getIntEnv(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err != nil {
			return i
		}
	}
	return fallback
}

func getInt64Env(key string, fallback int64) int64 {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.ParseInt(value, 10, 64); err != nil {
			return i
		}
	}

	return fallback
}

func getBoolEnv(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}
