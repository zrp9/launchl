// Package store contains for creating an connection to db
package store

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/zrp9/launchl/internal/config"
	"github.com/zrp9/launchl/internal/domain"
)

type Persister interface {
	DB() *sql.DB
	BnDB() *bun.DB
}

type Store struct {
	db  *sql.DB
	BdB bun.DB
}

func (s Store) DB() *sql.DB {
	return s.db
}

func (s Store) BnDB() *bun.DB {
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

	b.bdb.RegisterModel((*domain.UserSurvey)(nil))
	return b
}

func (b *StoreBuilder) Build() Persister {
	if b.db == nil {
		panic("Database connecciton must be set before building")
	}
	return &Store{
		db:  b.db,
		BdB: b.bdb,
	}
}

func InitStore(db *sql.DB) Store {
	return Store{
		db:  db,
		BdB: *bun.NewDB(db, pgdialect.New()),
	}
}

func Con() *sql.DB {
	dbc := "postgres://postgres:zroot_1119@18.226.170.114:5432/alessor?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dbc)))
	return sqldb
}

func (s Store) TestConnection() error {
	return s.db.Ping()
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
