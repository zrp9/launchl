// Package repos provides a generic interface for basic crud operations.
// Each method either returns a json strong to be marshalled into a type or a error or both.
package repos

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/uptrace/bun"
	"github.com/zrp9/launchl/internal/database/store"
)

// when using have to do something like this
// type UserRepo struct {
// 	*BasicRepo[string, User]
// }

// func NewUserRepo(db *sql.DB) *UserRepo {
// 	return &UserRepo{
// 		BasicRepo: New[string, User](db),
// 	}
// }

var ErrNoRecords = errors.New("no records found")
var ErrDBWrite = errors.New("failed to write to db")
var ErrDBRead = errors.New("failed to read db")
var ErrDBDelete = errors.New("failed to delete db record")
var ErrFailedTransaction = errors.New("an issue occurred with the transaction")
var ErrFailedRollback = errors.New("failed to rollback db")

type identifier interface {
	~int | ~string
}

type Saver interface {
	Name() string
}

type Repoer[T identifier, M any] interface {
	Get(ctx context.Context, key T) (*M, error)
	List(ctx context.Context) ([]*M, error)
	Save(ctx context.Context, o M) (*M, error)
	Update(ctx context.Context, key T) error
	Delete(ctx context.Context, o M) error
}

type BasicRepo[T identifier, M any] struct {
	store.Persister
}

func New[T identifier, M any](store store.Persister) *BasicRepo[T, M] {
	return &BasicRepo[T, M]{
		store,
	}
}

func (br BasicRepo[T, M]) Get(ctx context.Context, key T) (*M, error) {
	var domObj M
	err := br.BnDB().NewSelect().Model(&domObj).Where("? = ?", bun.Ident("id"), key).Scan(ctx, &domObj)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoRecords
		}
		return nil, err
	}

	return &domObj, nil
}

func (br BasicRepo[T, M]) GetAll(ctx context.Context) ([]*M, error) {
	var domObj []M
	err := br.BnDB().NewSelect().Model(&domObj).Scan(ctx, &domObj)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoRecords
		}
		return nil, err
	}

	objs := make([]*M, len(domObj))
	for i := range domObj {
		objs[i] = &domObj[i]
	}

	return objs, nil
}

func (br BasicRepo[T, M]) GetPaginated(ctx context.Context, page, limit int) ([]*M, error) {
	var domObj []M
	err := br.BnDB().NewSelect().Model(&domObj).Offset(page).Limit(limit).Scan(ctx, &domObj)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoRecords
		}
		return nil, err
	}

	objs := make([]*M, len(domObj))
	for i := range domObj {
		objs[i] = &domObj[i]
	}

	return objs, nil
}

func (br BasicRepo[T, M]) Create(ctx context.Context, m *M) (*M, error) {

	tx, err := br.BnDB().BeginTx(ctx, &sql.TxOptions{})

	if err != nil {
		return nil, err
	}

	// this below gives rowsEffected not the new user
	//rslt, err := tx.NewInsert().Model(user).Returning("*").Exec(ctx)
	err = tx.NewInsert().Model(m).
		Returning("*").Scan(ctx, m)

	if err != nil {
		emsg := fmt.Errorf("failed write operation %s", err)
		if err = tx.Rollback(); err != nil {
			log.Printf("db rollback err: %v", err)
			return nil, err
		}
		return nil, emsg
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return m, nil
}

func (br BasicRepo[T, M]) Update(ctx context.Context, k T, m *M) error {
	tx, err := br.BnDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	if _, err = tx.NewUpdate().Model(m).OmitZero().Where("? = ?", bun.Ident("id"), k).Exec(ctx, m); err != nil {
		emsg := fmt.Errorf("failed write operation %s", err)
		if err = tx.Rollback(); err != nil {
			return fmt.Errorf("failed to rollback after error %v", err)
		}
		return emsg
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (br BasicRepo[T, M]) Delete(ctx context.Context, k T) error {
	var domObj M
	tx, err := br.BnDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	_, err = tx.NewDelete().Model(&domObj).Where("? = ?", bun.Ident("id"), k).Exec(ctx)
	if err != nil {
		emsg := fmt.Errorf("failed to perform delete operation %v", err)
		if err = tx.Rollback(); err != nil {
			return err
		}
		return emsg
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
