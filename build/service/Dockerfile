FROM golang:1.25.0-alpine AS builder
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o /app/pvz -ldflags="-w -s" ./cmd/app

EXPOSE 8888

ENTRYPOINT [ "/app/pvz" ]
