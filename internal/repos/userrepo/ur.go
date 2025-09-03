// Package userrepo uses basic repo to make a user repository
package userrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/zrp9/launchl/internal/database/store"
	"github.com/zrp9/launchl/internal/domain"
	"github.com/zrp9/launchl/internal/repos"
)

// NOTE: i should be able to overrite a method like get if i specify it here
// maybe i should so i can use uuid to do lookups

type UserRepo struct {
	repo *repos.BasicRepo[string, domain.User]
}

func New(p store.Persister) UserRepo {
	return UserRepo{
		repo: repos.New[string, domain.User](p),
	}
}

func (u UserRepo) Get(ctx context.Context, uid string) (*domain.User, error) {
	var usr domain.User
	err := u.repo.BnDB().NewSelect().Model(&usr).Where("? = ?", bun.Ident("id"), uid).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repos.ErrNoRecords
		}
		return nil, errors.Join(repos.ErrDBRead, err)
	}

	return &usr, nil
}

func (u UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var usr domain.User
	err := u.repo.BnDB().NewSelect().Model(&usr).Where("? = ?", bun.Ident("email"), email).Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repos.ErrNoRecords
		}
		return nil, errors.Join(repos.ErrDBRead, err)
	}
	return &usr, nil
}

func (u UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var usr domain.User
	err := u.repo.BnDB().NewSelect().Model(&usr).Where("? = ?", bun.Ident("username"), username).Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repos.ErrNoRecords
		}
		return nil, errors.Join(repos.ErrDBRead, err)
	}
	return &usr, nil
}

func (u UserRepo) GetAll(ctx context.Context) ([]*domain.User, error) {
	usrs, err := u.repo.GetAll(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repos.ErrNoRecords
		}
		return nil, errors.Join(repos.ErrDBRead, err)
	}

	return usrs, nil
}

func (u UserRepo) GetPaginated(ctx context.Context, pg, limit int) ([]*domain.User, error) {
	users, err := u.repo.GetPaginated(ctx, pg, limit)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (u UserRepo) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	var usr = user
	tx, err := u.repo.BnDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, errors.Join(repos.ErrFailedTransaction, err)
	}

	err = tx.NewInsert().Model(&usr).Returning("*").Scan(ctx, &usr)
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return nil, errors.Join(repos.ErrFailedTransaction, err)
		}
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return usr, nil
}

func (u UserRepo) Update(ctx context.Context, usr domain.User) (*domain.User, error) {
	var user domain.User
	tx, err := u.repo.BnDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, errors.Join(repos.ErrFailedTransaction, err)
	}

	err = tx.NewUpdate().Model(&user).Where("? = ?", bun.Ident("id"), usr.ID).Scan(ctx, &user)
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return nil, errors.Join(repos.ErrFailedRollback, err)
		}
		return nil, errors.Join(repos.ErrDBWrite, err)
	}

	return &user, nil
}

func (u UserRepo) Delete(ctx context.Context, id string) error {
	var usr domain.User
	tx, err := u.repo.BnDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return errors.Join(repos.ErrFailedTransaction, err)
	}

	if _, err = tx.NewDelete().Model(&usr).Where("? = ?", bun.Ident("id"), id).Exec(ctx); err != nil {
		return errors.Join(repos.ErrDBDelete, err)
	}

	return nil
}

func (u UserRepo) DeleteByEmail(ctx context.Context, email string) error {
	var usr domain.User
	tx, err := u.repo.BnDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return errors.Join(repos.ErrFailedTransaction, err)
	}

	if _, err := tx.NewDelete().Model(&usr).Where("? = ?", bun.Ident("email"), email).Exec(ctx); err != nil {
		return errors.Join(repos.ErrDBDelete, err)
	}
	return nil
}

func (u UserRepo) DeleteByUsername(ctx context.Context, usrname string) error {
	var usr domain.User
	tx, err := u.repo.BnDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return errors.Join(repos.ErrFailedTransaction, err)
	}

	if _, err := tx.NewDelete().Model(&usr).Where("? = ?", bun.Ident("username"), usrname).Exec(ctx); err != nil {
		return errors.Join(repos.ErrDBDelete, err)
	}
	return nil
}

func (u UserRepo) GetQuePosition(ctx context.Context, usrname string) (int64, error) {
	var usr domain.User
	tx, err := u.repo.BnDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return -1, errors.Join(repos.ErrFailedTransaction, err)
	}

	err = tx.NewSelect().Model(&usr).Where("? = ?", bun.Ident("username"), usrname).Scan(ctx, &usr)
	if err != nil {
		return -1, errors.Join(repos.ErrDBRead, err)
	}

	return usr.QuePosition, nil
}

func (u UserRepo) GetByRefererID(ctx context.Context, refID string) (domain.User, error) {
	var usr domain.User
	err := u.repo.BnDB().NewSelect().Model(&usr).Where("? = ?", bun.Ident("referer_id"), refID).Scan(ctx, &usr)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.User{}, fmt.Errorf("could not find user %v", refID)
		}
		return domain.User{}, err
	}

	return usr, nil
}

func (u UserRepo) GetReferer(ctx context.Context, usrname, refID string) (domain.User, error) {
	var usr domain.User
	err := u.repo.BnDB().NewSelect().Model(&usr).Where("? = ?", bun.Ident("referer_id"), refID).Where("? = ?", bun.Ident("username"), usrname).Scan(ctx, &usr)
	if err != nil {
		return domain.User{}, err
	}

	return usr, nil
}

func (u UserRepo) FetchByUsername(ctx context.Context, usrname string) (domain.User, error) {
	var usr domain.User
	err := u.repo.BnDB().NewSelect().Model(&domain.User{}).Where("? = ?", bun.Ident("username"), usrname).Scan(ctx, &usr)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.User{}, fmt.Errorf("could not find user %w", err)
		}
		return domain.User{}, err
	}

	return domain.User{}, nil
}
