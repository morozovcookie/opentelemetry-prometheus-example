package percona

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
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
	exist, err := svc.checkUserAccountExistent(ctx, tx, ua.Username)
	if err != nil {
		return err
	}

	if exist {
		return &otelexample.Error{
			Code:    otelexample.ErrorCodeConflict,
			Message: fmt.Sprintf(`user account with username "%s" already exist`, ua.Username),
			Err:     nil,
		}
	}

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

func (svc *UserAccountService) checkUserAccountExistent(ctx context.Context, tx Tx, username string) (bool, error) {
	stmt, err := tx.PrepareContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_accounts ua WHERE ua.username = ?) `+
		`AS is_exists`)
	if err != nil {
		return false, err
	}

	defer func(ctx context.Context, stmt Stmt, err *error) {
		if closeErr := stmt.Close(ctx); closeErr != nil {
			*err = closeErr
		}
	}(ctx, stmt, &err)

	var exists bool
	if err := stmt.QueryRowContext(ctx, username).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
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
	*otelexample.FindUserAccountsResult,
	error,
) {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second))
	defer cancel()

	var (
		result = new(otelexample.FindUserAccountsResult)

		err error
	)

	result.Options = opts
	if result.Total, err = svc.findUserAccountsCountTotal(ctx); err != nil {
		return nil, fmt.Errorf("find user accounts: %w", err)
	}

	if result.Data, err = svc.findUserAccounts(ctx, opts); err != nil {
		return nil, fmt.Errorf("find user accounts: %w", err)
	}

	if result.HasNext, err = svc.findUserAccountsHasNext(ctx, opts.Offset()+opts.Limit()); err != nil {
		return nil, fmt.Errorf("find user accounts: %w", err)
	}

	return result, nil
}

func (svc *UserAccountService) findUserAccountsCountTotal(ctx context.Context) (uint64, error) {
	stmt, err := svc.preparer.PrepareContext(ctx, `SELECT count(1) FROM user_accounts`)
	if err != nil {
		return 0, err
	}

	defer func(ctx context.Context, stmt Stmt, err *error) {
		if closeErr := stmt.Close(ctx); closeErr != nil {
			*err = closeErr
		}
	}(ctx, stmt, &err)

	var total uint64

	if err := stmt.QueryRowContext(ctx).Scan(&total); err != nil {
		return 0, err
	}

	return total, nil
}

func (svc *UserAccountService) findUserAccounts(
	ctx context.Context,
	opts otelexample.FindOptions,
) (
	[]*otelexample.UserAccount,
	error,
) {
	stmt, err := svc.preparer.PrepareContext(ctx, `SELECT * FROM (SELECT ROW_NUMBER() OVER `+
		`(ORDER BY ua.row_id) as row_num, ua.user_account_id, ua.username, ua.created_at AS ua_created_at, `+
		`u.user_id, u.first_name, u.last_name, u.created_at AS u_created_at FROM user_accounts ua JOIN users u ON `+
		`ua.user_id = u.user_id) AS subquery WHERE row_num > ? LIMIT ?`)
	if err != nil {
		return nil, err
	}

	defer func(ctx context.Context, stmt Stmt, err *error) {
		if closeErr := stmt.Close(ctx); closeErr != nil {
			*err = closeErr
		}
	}(ctx, stmt, &err)

	rows, err := stmt.QueryContext(ctx, opts.Offset(), opts.Limit())
	if err != nil {
		return nil, err
	}

	defer func(closer io.Closer, err *error) {
		if closeErr := rows.Close(); closeErr != nil {
			*err = closeErr
		}
	}(rows, &err)

	uaa := make([]*otelexample.UserAccount, 0, opts.Limit())

	for rows.Next() {
		var (
			rowNumber     int64
			createdAt     int64
			userCreatedAt int64
		)

		ua := new(otelexample.UserAccount)
		ua.User = new(otelexample.User)

		err = rows.Scan(&rowNumber, &ua.ID, &ua.Username, &createdAt, &ua.User.ID, &ua.User.FirstName,
			&ua.User.LastName, &userCreatedAt)
		if err != nil {
			return nil, err
		}

		ua.CreatedAt, ua.User.CreatedAt = time.UnixMilli(createdAt), time.UnixMilli(userCreatedAt)

		uaa = append(uaa, ua)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return uaa, nil
}

func (svc *UserAccountService) findUserAccountsHasNext(ctx context.Context, offset uint64) (bool, error) {
	stmt, err := svc.preparer.PrepareContext(ctx, `SELECT EXISTS(SELECT 1 FROM (SELECT ROW_NUMBER() `+
		`OVER (ORDER BY ua.row_id) as row_num FROM user_accounts ua) AS subquery WHERE row_num > ? LIMIT 1) `+
		`AS has_next`)
	if err != nil {
		return false, err
	}

	defer func(ctx context.Context, stmt Stmt, err *error) {
		if closeErr := stmt.Close(ctx); closeErr != nil {
			*err = closeErr
		}
	}(ctx, stmt, &err)

	var hasNext bool
	if err = stmt.QueryRowContext(ctx, offset).Scan(&hasNext); err != nil {
		return false, err
	}

	return hasNext, nil
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

	stmt, err := svc.preparer.PrepareContext(ctx, `SELECT ua.username, ua.created_at AS ua_created_at, `+
		`u.user_id, u.first_name, u.last_name, u.created_at AS u_created_at FROM user_accounts as ua JOIN users as u `+
		`ON ua.user_id = u.user_id WHERE ua.user_account_id = ?`)
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
