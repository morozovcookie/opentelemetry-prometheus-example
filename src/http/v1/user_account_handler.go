package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	otelexample "github.com/morozovcookie/opentelemetry-prometheus-example"
)

const (
	UserAccountHandlerPathPrefix = "/api/v1/user-accounts"

	CreateUserAccountPathPrefix = "/"
	FindUserAccountsPathPrefix  = "/"
	FindUserAccountPathPrefix   = "/{id}"
)

var _ http.Handler = (*UserAccountHandler)(nil)

// UserAccountHandler represents a controller for handling
// operations with otelexample.UserAccount via HTTP requests.
type UserAccountHandler struct {
	http.Handler

	baseURL *url.URL

	userAccountService otelexample.UserAccountService
}

// NewUserAccountHandler returns a new instance of UserAccountHandler.
func NewUserAccountHandler(baseURL *url.URL, userAccountService otelexample.UserAccountService) *UserAccountHandler {
	var (
		router  = chi.NewRouter()
		handler = &UserAccountHandler{
			Handler: router,

			baseURL: baseURL,

			userAccountService: userAccountService,
		}
	)

	router.Post(CreateUserAccountPathPrefix, handler.handleCreateUserAccount)
	router.Get(FindUserAccountsPathPrefix, handler.handleFindUserAccounts)
	router.Get(FindUserAccountPathPrefix, handler.handleFindUserAccount)

	return handler
}

// CreateUserAccountRequest is the request body
// for creating otelexample.UserAccount.
type CreateUserAccountRequest struct {
	// Username is the username.
	Username string `json:"username"`

	// FirstName is the user first name.
	FirstName string `json:"firstName"`

	// LastName is the user last name.
	LastName string `json:"lastName"`
}

func decodeCreateUserAccount(reader io.Reader) (*CreateUserAccountRequest, error) {
	decoded := new(CreateUserAccountRequest)

	if err := json.NewDecoder(reader).Decode(decoded); err != nil {
		return nil, fmt.Errorf("decode CreateUserAccountRequest: %w", &otelexample.Error{
			Code:    otelexample.ErrorCodeInvalid,
			Message: "failed to decode request",
			Err:     err,
		})
	}

	for _, field := range []struct {
		name string
		val  string
	}{
		{
			name: decoded.Username,
			val:  "username",
		},
		{
			name: decoded.FirstName,
			val:  "firstName",
		},
		{
			name: decoded.LastName,
			val:  "lastName",
		},
	} {
		if err := checkOnEmptyString(field.val, field.name); err != nil {
			return nil, fmt.Errorf("decode CreateUserAccountRequest: %w", err)
		}
	}

	return decoded, nil
}

func checkOnEmptyString(val, name string) error {
	if val != "" {
		return nil
	}

	return &otelexample.Error{
		Code:    otelexample.ErrorCodeInvalid,
		Message: fmt.Sprintf(`"%s" could not be empty`, name),
		Err:     nil,
	}
}

func (h *UserAccountHandler) handleCreateUserAccount(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	decoded, err := decodeCreateUserAccount(request.Body)
	if err != nil {
		encodeErrorResponse(writer, err)

		return
	}

	account := &otelexample.UserAccount{
		ID:       otelexample.EmptyID,
		Username: decoded.Username,
		User: &otelexample.User{
			ID:        otelexample.EmptyID,
			FirstName: decoded.FirstName,
			LastName:  decoded.LastName,
			CreatedAt: time.Time{},
		},
		CreatedAt: time.Time{},
	}

	err = h.userAccountService.CreateUserAccount(ctx, account)
	if err != nil {
		encodeErrorResponse(writer, err)

		return
	}

	writer.Header().Set("Location", fmt.Sprintf("%s/%s", UserAccountHandlerPathPrefix, account.ID))
	encodeResponse(writer, http.StatusCreated, nil)
}

// FindUserAccountsRequest is the request parameters for
// retrieve user accounts list.
type FindUserAccountsRequest struct {
	// Start is the count of records that should be skipped.
	Start uint64

	// Limit is the maximum records that should be returned.
	Limit uint64
}

func decodeFindUserAccountsRequest(request *http.Request) (*FindUserAccountsRequest, error) {
	var (
		decoded   = new(FindUserAccountsRequest)
		queryArgs = request.URL.Query()

		err error
	)

	const (
		decimal    = 10
		uint64Size = 64
	)

	if start := queryArgs.Get("start"); start != "" {
		if decoded.Start, err = strconv.ParseUint(start, decimal, uint64Size); err != nil {
			return nil, &otelexample.Error{
				Code:    otelexample.ErrorCodeInvalid,
				Message: "failed to parse start value",
				Err:     err,
			}
		}
	}

	if limit := queryArgs.Get("limit"); limit != "" {
		if decoded.Limit, err = strconv.ParseUint(limit, decimal, uint64Size); err != nil {
			return nil, &otelexample.Error{
				Code:    otelexample.ErrorCodeInvalid,
				Message: "failed to parse limit value",
				Err:     err,
			}
		}
	}

	return decoded, nil
}

// FindUserAccountsResponse represents the result of user accounts search.
type FindUserAccountsResponse struct {
	// Links is the set of links for dynamic navigation.
	Links *Links `json:"_links"` // nolint:tagliatelle

	// Start is the count of records that should be skipped.
	Start uint64 `json:"start"`

	// Limit is the maximum records that should be returned.
	Limit uint64 `json:"limit"`

	// Total is the count of records that matches of request
	// without pagination parameters.
	Total uint64 `json:"total"`

	// Data is the list of user accounts that was found.
	Data []*UserAccount `json:"data"`
}

func newFindUserAccountsResponse(
	baseURL *url.URL,
	result *otelexample.FindUserAccountsResult,
) (*FindUserAccountsResponse, error) {
	var (
		limit = result.Options.Limit()
		start = result.Options.Offset()

		response = &FindUserAccountsResponse{
			Links: nil,
			Start: start,
			Limit: limit,
			Total: result.Total,
			Data:  make([]*UserAccount, len(result.Data)),
		}

		err error
	)

	response.Links, err = newLinks(baseURL, UserAccountHandlerPathPrefix, result.Options, result.HasNext)
	if err != nil {
		return nil, fmt.Errorf("create FindUserAccountsResponse: %w", err)
	}

	for i, ua := range result.Data {
		if response.Data[i], err = newUserAccount(baseURL, ua); err != nil {
			return nil, fmt.Errorf("create FindUserAccountsResponse: %w", err)
		}
	}

	return response, nil
}

func (h *UserAccountHandler) handleFindUserAccounts(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	decoded, err := decodeFindUserAccountsRequest(request)
	if err != nil {
		encodeErrorResponse(writer, err)

		return
	}

	opts := otelexample.NewFindOptions(decoded.Limit, decoded.Start)

	result, err := h.userAccountService.FindUserAccounts(ctx, opts)
	if err != nil {
		encodeErrorResponse(writer, err)

		return
	}

	response, err := newFindUserAccountsResponse(h.baseURL, result)
	if err != nil {
		encodeErrorResponse(writer, err)

		return
	}

	encodeResponse(writer, http.StatusOK, response)
}

// FindUserAccountRequest is the request parameters for retrieve a single user account.
type FindUserAccountRequest struct {
	// ID is the user account unique identifier.
	ID otelexample.ID
}

func decodeFindUserAccountRequest(request *http.Request) *FindUserAccountRequest {
	return &FindUserAccountRequest{
		ID: otelexample.ID(chi.URLParam(request, "id")),
	}
}

// User describes the real person.
type User struct {
	// CreatedAt is the time when user was created.
	CreatedAt int64 `json:"createdAt"`

	// ID is the user unique identifier.
	ID string `json:"id"`

	// FirstName is the user first name.
	FirstName string `json:"firstName"`

	// LastName is the user last name.
	LastName string `json:"lastName"`
}

// UserAccount is the user account in the system.
type UserAccount struct {
	// Link is the link to the object themselves.
	Link *SelfLink `json:"_links"` // nolint:tagliatelle

	// User is the person who owned the account.
	User *User `json:"user"`

	// CreatedAt is the time when user account was created.
	CreatedAt int64 `json:"createdAt"`

	// ID is the user account unique identifier.
	ID string `json:"id"`

	// Username is the user account name.
	Username string `json:"username"`
}

// FindUserAccountResponse represents the result of user account search.
type FindUserAccountResponse struct {
	*UserAccount
}

func newUserAccount(baseURL *url.URL, ua *otelexample.UserAccount) (*UserAccount, error) {
	out := &UserAccount{
		Link: &SelfLink{
			Self: "",
		},
		User: &User{
			ID:        ua.User.ID.String(),
			FirstName: ua.User.FirstName,
			LastName:  ua.User.LastName,
			CreatedAt: ua.User.CreatedAt.UnixMilli(),
		},
		CreatedAt: ua.CreatedAt.UnixMilli(),
		ID:        ua.ID.String(),
		Username:  ua.Username,
	}

	selfLink, err := baseURL.Parse(fmt.Sprintf("%s/%s", UserAccountHandlerPathPrefix, ua.ID))
	if err != nil {
		return nil, fmt.Errorf("create UserAccount: %w", err)
	}

	out.Link.Self = selfLink.String()

	return out, nil
}

func newFindUserAccountResponse(baseURL *url.URL, ua *otelexample.UserAccount) (*FindUserAccountResponse, error) {
	var (
		response = new(FindUserAccountResponse)

		err error
	)

	if response.UserAccount, err = newUserAccount(baseURL, ua); err != nil {
		return nil, err
	}

	return response, nil
}

func (h *UserAccountHandler) handleFindUserAccount(writer http.ResponseWriter, request *http.Request) {
	var (
		ctx     = request.Context()
		decoded = decodeFindUserAccountRequest(request)
	)

	ua, err := h.userAccountService.FindUserAccountByID(ctx, decoded.ID)
	if err != nil {
		encodeErrorResponse(writer, err)

		return
	}

	response, err := newFindUserAccountResponse(h.baseURL, ua)
	if err != nil {
		encodeErrorResponse(writer, err)

		return
	}

	encodeResponse(writer, http.StatusOK, response)
}
