package v1

import (
	"bytes"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const (
	UserAccountHandlerPathPrefix = "/api/v1/user-accounts"

	CreateUserAccountPathPrefix = "/"
	FindUserAccountsPathPrefix  = "/"
)

var _ http.Handler = (*UserAccountHandler)(nil)

type UserAccountHandler struct {
	http.Handler
}

func NewUserAccountHandler() *UserAccountHandler {
	var (
		router  = chi.NewRouter()
		handler = &UserAccountHandler{
			Handler: router,
		}
	)

	router.Post(CreateUserAccountPathPrefix, handler.handleCreateUserAccount)
	router.Get(FindUserAccountsPathPrefix, handler.handleFindUserAccounts)

	return handler
}

type CreateUserAccountRequest struct{}

func decodeCreateUserAccount() (*CreateUserAccountRequest, error) {
	return nil, nil
}

type CreateUserAccountResponse struct{}

func newCreateUserAccountResponse() *CreateUserAccountResponse {
	return nil
}

func (h *UserAccountHandler) handleCreateUserAccount(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusCreated)
	_, _ = io.Copy(writer, bytes.NewBufferString("{}\n"))
}

func (h *UserAccountHandler) handleFindUserAccounts(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	_, _ = io.Copy(writer, bytes.NewBufferString("{}\n"))
}
