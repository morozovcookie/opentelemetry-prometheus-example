package zap

import (
	"context"

	otelexample "github.com/morozovcookie/opentelemetry-prometheus-example"
	"go.uber.org/zap"
)

var _ otelexample.UserAccountService = (*UserAccountService)(nil)

// UserAccountService represents a service for managing UserAccount data.
type UserAccountService struct {
	wrapped otelexample.UserAccountService
	logger  *zap.Logger
}

// NewUserAccountService returns a new instance of UserAccountService.
func NewUserAccountService(svc otelexample.UserAccountService, logger *zap.Logger) *UserAccountService {
	return &UserAccountService{
		wrapped: svc,
		logger:  logger,
	}
}

// CreateUserAccount creates a new user account.
func (svc *UserAccountService) CreateUserAccount(ctx context.Context, ua *otelexample.UserAccount) error {
	var (
		before = ua.Clone()

		err error
	)

	start, end, elapsed := trackOfTime(func() {
		err = svc.wrapped.CreateUserAccount(ctx, ua)
	})

	ff := []zap.Field{
		zap.Stringer("start", start), zap.Stringer("end", end), zap.Stringer("elapsed", elapsed),
		zap.Any("before", before), zap.Any("after", ua), zap.Error(err),
	}

	svc.logger.Debug("create user account", ff...)

	if err != nil {
		svc.logger.Error("create user account", ff...)

		return err // nolint:wrapcheck
	}

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
	var (
		result *otelexample.FindUserAccountsResult
		err    error
	)

	start, end, elapsed := trackOfTime(func() {
		result, err = svc.wrapped.FindUserAccounts(ctx, opts)
	})

	var count int
	if result != nil {
		count = len(result.Data)
	}

	ff := []zap.Field{
		zap.Stringer("start", start), zap.Stringer("end", end), zap.Stringer("elapsed", elapsed),
		zap.Any("options", opts), zap.Any("result", result), zap.Int("dataSize", count),
		zap.Error(err),
	}

	svc.logger.Debug("find user accounts", ff...)

	if err != nil {
		svc.logger.Error("find user accounts", ff...)

		return nil, err // nolint:wrapcheck
	}

	return result, nil
}

// FindUserAccountByID returns user account by unique identifier.
func (svc *UserAccountService) FindUserAccountByID(
	ctx context.Context,
	id otelexample.ID,
) (
	*otelexample.UserAccount,
	error,
) {
	var (
		ua  *otelexample.UserAccount
		err error
	)

	start, end, elapsed := trackOfTime(func() {
		ua, err = svc.wrapped.FindUserAccountByID(ctx, id)
	})

	ff := []zap.Field{
		zap.Stringer("start", start), zap.Stringer("end", end), zap.Stringer("elapsed", elapsed),
		zap.Stringer("accountId", id), zap.Any("account", ua), zap.Error(err),
	}

	svc.logger.Debug("find user account by id", ff...)

	if err != nil {
		svc.logger.Error("find user account by id", ff...)

		return nil, err // nolint:wrapcheck
	}

	return ua, nil
}
