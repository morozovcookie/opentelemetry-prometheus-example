# golang:1.18.3-alpine3.16
FROM golang@sha256:725f8fd50191209a4c4a00def1d93c4193c4d0a1c2900139daf8f742480f3367 as builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# add a non-privileged user
RUN addgroup -S nonroot && adduser -u 10001 -S nonroot -G nonroot

WORKDIR /src

COPY ./src/go.mod ./src/go.sum ./
COPY ./src/vendor ./vendor

COPY ./src/*.go ./src

COPY ./src/opentelemetry ./opentelemetry
COPY ./src/zap ./zap

COPY ./src/percona ./percona
COPY ./src/http ./http

COPY ./src/cmd/server ./cmd/server

RUN go build \
    -mod=vendor \
    -ldflags "-s -w" \
    -o /usr/local/server ./cmd/server/...

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

COPY --from=builder /usr/local/server /usr/local/server

USER nonroot

ENTRYPOINT ["/usr/local/server"]
