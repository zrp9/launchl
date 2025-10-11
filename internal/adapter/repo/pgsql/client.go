// Package pgsql provides a postgresql repo
package pgsql

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/zrp9/launchl/internal/config"
	"github.com/zrp9/launchl/internal/domain/core"
)

type PGClient interface {
	DB() *sql.DB
	BnDB() *bun.DB
}

type PGStore struct {
	db  *sql.DB
	BdB bun.DB
}

func (s PGStore) DB() *sql.DB {
	return s.db
}

func (s PGStore) BnDB() *bun.DB {
	return &s.BdB
}

type StoreBuilder struct {
	db  *sql.DB
	bdb bun.DB
}

func NewBuilder() *StoreBuilder {
	return &StoreBuilder{}
}

func (b *StoreBuilder) SetDB(db *sql.DB) *StoreBuilder {
	b.db = db
	return b
}

func (b *StoreBuilder) SetBunDB() *StoreBuilder {
	if b.db == nil {
		panic("SetDb must be called first")
	}
	b.bdb = *bun.NewDB(b.db, pgdialect.New())
	b.bdb.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
	))
	return b
}

func (b *StoreBuilder) RegisterModels() *StoreBuilder {
	if b.db == nil {
		panic("Register models needs db")
	}

	// TODO: need to check if this is many 2 many
	b.bdb.RegisterModel((*core.SurveyResponse)(nil))
	return b
}

func (b *StoreBuilder) Build() PGClient {
	if b.db == nil {
		panic("Database connecciton must be set before building")
	}
	return &PGStore{
		db:  b.db,
		BdB: b.bdb,
	}
}

func InitStore(db *sql.DB) PGStore {
	return PGStore{
		db:  db,
		BdB: *bun.NewDB(db, pgdialect.New()),
	}
}

func Con() *sql.DB {
	dbc := "postgres://postgres:zroot_1119@18.226.170.114:5432/alessor?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dbc)))
	return sqldb
}

func (s PGStore) TestConnection() error {
	return s.db.Ping()
}

func DBConWithRetry(cfg config.DatabaseCfg) (*sql.DB, error) {
	backoff := []time.Duration{500 * time.Millisecond, 1 * time.Second, 2 * time.Second, 3 * time.Second, 5 * time.Second}
	var db *sql.DB
	var err error

	for _, d := range backoff {
		db, err = DBCon(cfg)
		if err == nil {
			return db, nil
		}
		// unwrap for pgdriver error code checks
		var pgErr *pgdriver.Error
		if errors.As(err, &pgErr) && (pgErr.Field('C') == "57P03") { // in recovery
			time.Sleep(d)
			continue
		}

		// also retry common startup errors
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "i/o timeout") {
			time.Sleep(d)
			continue
		}
		// any other error: fail fast
		return nil, err
	}
	return nil, err
}

func DBCon(dbConf config.DatabaseCfg) (*sql.DB, error) {
	var tlsConfig *tls.Config

	if dbConf.SSLMode != "disable" {
		rootCertPool := x509.NewCertPool()
		pem, err := os.ReadFile(dbConf.SSLRoot)
		if err != nil {
			log.Printf("failed to read CA certificate: %v", err)
			return nil, err
		}
		if !rootCertPool.AppendCertsFromPEM(pem) {
			log.Println("fialed to append CA certificate")
			return nil, err
		}
		tlsConfig = &tls.Config{
			RootCAs:            rootCertPool,
			InsecureSkipVerify: true, // might need to change if set to false current config will reject bc handshake
		}
	}

	pgOpts := []pgdriver.Option{
		pgdriver.WithNetwork("tcp"),
		pgdriver.WithAddr(fmt.Sprintf("%v:%v", dbConf.Host, dbConf.Port)),
		pgdriver.WithUser(dbConf.User),
		pgdriver.WithPassword(dbConf.Password),
		pgdriver.WithDatabase(dbConf.Name),
		pgdriver.WithDialTimeout(time.Second * time.Duration(dbConf.DialTimeout)),
		pgdriver.WithReadTimeout(time.Second * time.Duration(dbConf.ReadTimeout)),
		pgdriver.WithWriteTimeout(time.Second * time.Duration(dbConf.WriteTimeout)),
		pgdriver.WithTLSConfig(tlsConfig),
	}

	pgcon := pgdriver.NewConnector(pgOpts...)
	db := sql.OpenDB(pgcon)
	db.SetMaxOpenConns(dbConf.MaxOpenConns)
	db.SetMaxIdleConns(dbConf.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(dbConf.ConnTimeout) * time.Minute)

	if err := db.Ping(); err != nil {
		log.Printf("failed to connect to database: %v", err)
		return nil, err
	}

	return db, nil
}
