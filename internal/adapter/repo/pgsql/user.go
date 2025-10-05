package pgsql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"
	"github.com/zrp9/launchl/internal/domain/core"
)

// NOTE: i should be able to overrite a method like get if i specify it here
// maybe i should so i can use uuid to do lookups

type UserRepo struct {
	repo *BasicRepo[string, core.User]
}

func NewUserRepo(p PGClient) UserRepo {
	return UserRepo{
		repo: NewBasicRepo[string, core.User](p),
	}
}

func (u UserRepo) GetByEmail(ctx context.Context, email string) (*core.User, error) {
	var usr core.User
	err := u.repo.BnDB().NewSelect().Model(&usr).Where("? = ?", bun.Ident("email"), email).Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoRecords
		}
		return nil, errors.Join(ErrDBRead, err)
	}
	return &usr, nil
}

func (u UserRepo) Create(ctx context.Context, user *core.User, rolename string) (*core.User, error) {
	var usr = user
	tx, err := u.repo.BnDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, errors.Join(ErrFailedTransaction, err)
	}

	var role core.Role
	if err = tx.NewSelect().Model(&role).Where("? = ?", bun.Ident("name"), rolename).Scan(ctx, &role); err != nil {
		return nil, err
	}

	user.Role = &role
	err = tx.NewInsert().Model(&user).Returning("*").Scan(ctx, &usr)
	if err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return nil, errors.Join(ErrFailedTransaction, err)
		}
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return usr, nil
}

func (u UserRepo) DeleteByEmail(ctx context.Context, email string) error {
	var usr core.User
	tx, err := u.repo.BnDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return errors.Join(ErrFailedTransaction, err)
	}

	if _, err := tx.NewDelete().Model(&usr).Where("? = ?", bun.Ident("email"), email).Exec(ctx); err != nil {
		return errors.Join(ErrDBDelete, err)
	}
	return nil
}

func (u UserRepo) GetQuePosition(ctx context.Context, usrname string) (int64, error) {
	var usr core.User
	tx, err := u.repo.BnDB().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return -1, errors.Join(ErrFailedTransaction, err)
	}

	err = tx.NewSelect().Model(&usr).Where("? = ?", bun.Ident("username"), usrname).Scan(ctx, &usr)
	if err != nil {
		return -1, errors.Join(ErrDBRead, err)
	}

	return usr.QuePosition, nil
}

func (u UserRepo) UpdateByReferer(ctx context.Context, referee core.User, refererID string) error {
	if err := u.repo.BnDB().RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var role core.Role
		if err := tx.NewSelect().Model(&role).Where("? = ?", bun.Ident("name"), "subscriber").Scan(ctx, &role); err != nil {
			return err
		}

		var nwUsr core.User
		referee.Role = &role
		if err := tx.NewInsert().Model(&referee).Scan(ctx, nwUsr); err != nil {
			return err
		}

		var refereer core.User
		if err := tx.NewSelect().Model(&core.User{}).Where("? = ?", bun.Ident("referer_id"), refererID).Scan(ctx, &refereer); err != nil {
			return err
		}

		if err := tx.NewUpdate().Model(&core.User{}).Where("? = ?", bun.Ident("id"), refereer.ID).Scan(ctx, &core.User{}); err != nil {
			return err
		}

		if _, err := tx.NewInsert().Model(&core.Referal{RefererID: refereer.ID, RefereeID: nwUsr.ID}).Exec(ctx); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (u UserRepo) GetByUsername(ctx context.Context, uname string) (core.User, error) {
	var usr core.User
	if err := u.repo.BnDB().NewSelect().Model(&usr).Where("? = ?", bun.Ident("username"), uname); err != nil {
		return core.User{}, nil
	}

	return usr, nil
}
