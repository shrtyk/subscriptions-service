FROM golang:1.25.0-alpine

RUN go install github.com/pressly/goose/v3/cmd/goose@v3.26.0

WORKDIR /migrations

ENTRYPOINT ["/go/bin/goose"]
