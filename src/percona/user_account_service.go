package percona

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	otelexample "github.com/morozovcookie/opentelemetry-prometheus-example"
)

// UserAccountService represents a service for managing UserAccount data.
type UserAccountService struct {
	txBeginner TxBeginner
	preparer   Preparer

	identifierGenerator otelexample.IdentifierGenerator
	timer               otelexample.Timer
}

// NewUserAccountService returns a new instance of UserAccountService.
func NewUserAccountService(
	beginner TxBeginner,
	preparer Preparer,
	identifierGenerator otelexample.IdentifierGenerator,
	timer otelexample.Timer,
) *UserAccountService {
	return &UserAccountService{
		txBeginner: beginner,
		preparer:   preparer,

		identifierGenerator: identifierGenerator,
		timer:               timer,
	}
}

// CreateUserAccount creates a new user account.
func (svc *UserAccountService) CreateUserAccount(ctx context.Context, ua *otelexample.UserAccount) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second))
	defer cancel()

	tx, err := svc.txBeginner.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("create user account: %w", err)
	}

	if err = svc.createUserAccount(ctx, tx, ua); err == nil {
		return nil
	}

	if rollbackErr := tx.Rollback(); rollbackErr != nil {
		return fmt.Errorf("create user account: %w", rollbackErr)
	}

	return fmt.Errorf("create user account: %w", err)
}

func (svc *UserAccountService) createUserAccount(ctx context.Context, tx Tx, ua *otelexample.UserAccount) error {
	if err := svc.createUserRow(ctx, tx, ua.User); err != nil {
		return err
	}

	if err := svc.createUserAccountRow(ctx, tx, ua); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (svc *UserAccountService) createUserRow(ctx context.Context, tx Tx, user *otelexample.User) error {
	var (
		createdAt = svc.timer.Time(ctx)
		userID    = svc.identifierGenerator.GenerateIdentifier(ctx)
	)

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO users (user_id, first_name, last_name, created_at) VALUES `+
		`(?,?,?,?)`)
	if err != nil {
		return err
	}

	defer func(ctx context.Context, stmt Stmt, err *error) {
		if closeErr := stmt.Close(ctx); closeErr != nil {
			*err = closeErr
		}
	}(ctx, stmt, &err)

	_, err = stmt.ExecContext(ctx, userID.String(), user.FirstName, user.LastName, createdAt.UnixMilli())
	if err != nil {
		return err
	}

	user.ID, user.CreatedAt = userID, createdAt

	return nil
}

func (svc *UserAccountService) createUserAccountRow(ctx context.Context, tx Tx, ua *otelexample.UserAccount) error {
	var (
		createdAt = svc.timer.Time(ctx)
		uaID      = svc.identifierGenerator.GenerateIdentifier(ctx)
	)

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO user_accounts (user_account_id,username,user_id,`+
		`created_at) VALUES (?,?,?,?)`)
	if err != nil {
		return err
	}

	defer func(ctx context.Context, stmt Stmt, err *error) {
		if closeErr := stmt.Close(ctx); closeErr != nil {
			*err = closeErr
		}
	}(ctx, stmt, &err)

	_, err = stmt.ExecContext(ctx, uaID.String(), ua.Username, ua.User.ID.String(), createdAt.UnixMilli())
	if err != nil {
		return err
	}

	ua.ID, ua.CreatedAt = uaID, createdAt

	return nil
}

// FindUserAccounts returns a list of user accounts.
func (svc *UserAccountService) FindUserAccounts(
	ctx context.Context,
	opts otelexample.FindOptions,
) (
	[]*otelexample.UserAccount,
	error,
) {
	return nil, nil
}

// FindUserAccountByID returns user account by unique identifier.
func (svc *UserAccountService) FindUserAccountByID(
	ctx context.Context,
	id otelexample.ID,
) (
	*otelexample.UserAccount,
	error,
) {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second))
	defer cancel()

	stmt, err := svc.preparer.PrepareContext(ctx, `SELECT ua.username, ua.created_at, u.user_id, u.first_name, `+
		`u.last_name, u.created_at FROM user_accounts as ua JOIN users as u ON ua.user_id = u.user_id `+
		`WHERE ua.user_account_id = ?`)
	if err != nil {
		return nil, fmt.Errorf("find user account by id: %w", err)
	}

	defer func(ctx context.Context, stmt Stmt, err *error) {
		if closeErr := stmt.Close(ctx); closeErr != nil {
			*err = closeErr
		}
	}(ctx, stmt, &err)

	var (
		userAccountCreatedAt int64
		userCreatedAt        int64

		ua = new(otelexample.UserAccount)
	)

	ua.ID, ua.User = id, new(otelexample.User)

	err = stmt.QueryRowContext(ctx, id.String()).Scan(&ua.Username, &userAccountCreatedAt, &ua.User.ID,
		&ua.User.FirstName, &ua.User.LastName, &userCreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("find user account by id: %w", &otelexample.Error{
			Code:    otelexample.ErrorCodeNotFound,
			Message: "user account does not exist",
			Err:     nil,
		})
	}

	if err != nil {
		return nil, fmt.Errorf("find user account by id: %w", err)
	}

	ua.CreatedAt, ua.User.CreatedAt = time.UnixMilli(userAccountCreatedAt), time.UnixMilli(userCreatedAt)

	return ua, nil
}
