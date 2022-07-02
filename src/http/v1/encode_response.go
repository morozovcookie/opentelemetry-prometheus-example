package v1

import (
	"encoding/json"
	"net/http"

	otelexample "github.com/morozovcookie/opentelemetry-prometheus-example"
)

func encodeResponse(writer http.ResponseWriter, status int, response any) {
	writer.WriteHeader(status)
	writer.Header().Set("Content-Type", "application/json")

	if response == nil {
		return
	}

	if err := json.NewEncoder(writer).Encode(response); err != nil {
		panic(err)
	}
}

// nolint:gochecknoglobals
var mapErrorCodeToStatusCode = map[otelexample.ErrorCode]int{
	otelexample.ErrorCodeOK: http.StatusOK,

	otelexample.ErrorCodeInvalid: http.StatusBadRequest,

	otelexample.ErrorCodeInternal: http.StatusInternalServerError,
}

func encodeErrorResponse(writer http.ResponseWriter, err error) {
	var (
		code     = otelexample.ErrorCodeFromError(err)
		status   = http.StatusInternalServerError
		response = &struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}{
			Code:    code.String(),
			Message: otelexample.ErrorMessageFromError(err),
		}
	)

	if statusCode, ok := mapErrorCodeToStatusCode[code]; ok {
		status = statusCode
	}

	encodeResponse(writer, status, response)
}
