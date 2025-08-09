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
	Redis      RedisCfg
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

type RedisCfg struct {
	Host     string
	Port     string
	Password string
	DB       int
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
	appEnv, err := initializeEnv()
	if err != nil {
		log.Println(err.Error())
	}

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
		Redis: RedisCfg{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
		},
		Jwt: JWTCfg{
			Secret:     mustGetEnv("JWT_SECRET"),
			Expiration: getDurationEnv("JWT_EXPIRATION", 24*time.Hour),
			Salty:      mustGetEnv("SALTY"),
		},
	}, nil
}

func LoadDBConfig() (DatabaseCfg, error) {
	_, err := initializeEnv()
	if err != nil {
		return DatabaseCfg{}, err
	}

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

func GetAuthToken() ([]byte, error) {
	_, err := initializeEnv()
	if err != nil {
		log.Println(err.Error())
	}

	authKey := mustGetEnv("AUTH_KEY")

	return []byte(authKey), nil
}

func initializeEnv() (string, error) {
	appEnv := getEnv("ENV", "dev")
	if err := godotenv.Load(fmt.Sprintf(".env.%s", appEnv)); err != nil {
		return "", fmt.Errorf("%s env does not exist", appEnv)
	}

	if err := godotenv.Load(".env"); err != nil {
		return "", fmt.Errorf("warning could not load .env file %v", err)
	}

	return appEnv, nil
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
