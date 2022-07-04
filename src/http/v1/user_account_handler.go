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

type UserAccountHandler struct {
	http.Handler

	baseURL *url.URL

	userAccountService otelexample.UserAccountService
}

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

type CreateUserAccountRequest struct {
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func decodeCreateUserAccount(reader io.Reader) (*CreateUserAccountRequest, error) {
	decoded := new(CreateUserAccountRequest)

	if err := json.NewDecoder(reader).Decode(decoded); err != nil {
		return nil, &otelexample.Error{
			Code:    otelexample.ErrorCodeInvalid,
			Message: "failed to decode request",
			Err:     err,
		}
	}

	return decoded, nil
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

type FindUserAccountsRequest struct {
	Start uint64
	Limit uint64
}

func decodeFindUserAccountsRequest(request *http.Request) (*FindUserAccountsRequest, error) {
	var (
		decoded   = new(FindUserAccountsRequest)
		queryArgs = request.URL.Query()

		err error
	)

	if start := queryArgs.Get("start"); start != "" {
		if decoded.Start, err = strconv.ParseUint(start, 10, 64); err != nil {
			return nil, &otelexample.Error{
				Code:    otelexample.ErrorCodeInvalid,
				Message: "failed to parse start value",
				Err:     err,
			}
		}
	}

	if limit := queryArgs.Get("limit"); limit != "" {
		if decoded.Limit, err = strconv.ParseUint(limit, 10, 64); err != nil {
			return nil, &otelexample.Error{
				Code:    otelexample.ErrorCodeInvalid,
				Message: "failed to parse limit value",
				Err:     err,
			}
		}
	}

	return decoded, nil
}

type FindUserAccountsResponse struct {
	Links *Links         `json:"_links"`
	Start uint64         `json:"start"`
	Limit uint64         `json:"limit"`
	Total uint64         `json:"total"`
	Data  []*UserAccount `json:"data"`
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

type FindUserAccountRequest struct {
	ID otelexample.ID
}

func decodeFindUserAccountRequest(request *http.Request) *FindUserAccountRequest {
	return &FindUserAccountRequest{
		ID: otelexample.ID(chi.URLParam(request, "id")),
	}
}

type User struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	CreatedAt int64  `json:"createdAt"`
}

type UserAccount struct {
	Link      *SelfLink `json:"_links"`
	User      *User     `json:"user"`
	CreatedAt int64     `json:"createdAt"`
	ID        string    `json:"id"`
	Username  string    `json:"username"`
}

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
