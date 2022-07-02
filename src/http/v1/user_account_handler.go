package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	userAccountService otelexample.UserAccountService
}

func NewUserAccountHandler(userAccountService otelexample.UserAccountService) *UserAccountHandler {
	var (
		router  = chi.NewRouter()
		handler = &UserAccountHandler{
			Handler: router,

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

func (h *UserAccountHandler) handleFindUserAccounts(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	_, _ = io.Copy(writer, bytes.NewBufferString("{}\n"))
}

type FindUserAccountRequest struct {
	ID otelexample.ID
}

func decodeFindUserAccount(request *http.Request) *FindUserAccountRequest {
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
	ID        string `json:"id"`
	Username  string `json:"username"`
	User      *User  `json:"user"`
	CreatedAt int64  `json:"createdAt"`
}

type FindUserAccountResponse UserAccount

func newFindUserAccountResponse(ua *otelexample.UserAccount) *FindUserAccountResponse {
	return &FindUserAccountResponse{
		ID:       ua.ID.String(),
		Username: ua.Username,
		User: &User{
			ID:        ua.User.ID.String(),
			FirstName: ua.User.FirstName,
			LastName:  ua.User.LastName,
			CreatedAt: ua.User.CreatedAt.UnixMilli(),
		},
		CreatedAt: ua.CreatedAt.UnixMilli(),
	}
}

func (h *UserAccountHandler) handleFindUserAccount(writer http.ResponseWriter, request *http.Request) {
	var (
		ctx     = request.Context()
		decoded = decodeFindUserAccount(request)
	)

	ua, err := h.userAccountService.FindUserAccountByID(ctx, decoded.ID)
	if err != nil {
		encodeErrorResponse(writer, err)

		return
	}

	encodeResponse(writer, http.StatusOK, newFindUserAccountResponse(ua))
}
