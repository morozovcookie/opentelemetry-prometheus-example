package v1

import (
	"bytes"
	"encoding/json"
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

	encodeResponse(writer, http.StatusCreated, nil)
}

func (h *UserAccountHandler) handleFindUserAccounts(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	_, _ = io.Copy(writer, bytes.NewBufferString("{}\n"))
}
