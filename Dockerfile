FROM golang:1.25.7-alpine

ARG GIT_VERSION

RUN apk update && apk upgrade && apk add --no-cache git

COPY main.go main.go
COPY internal internal
COPY go.mod go.mod
COPY go.sum go.sum

RUN CGO_ENABLED=0 go build -ldflags "-X main.Version=${GIT_VERSION}" -a -installsuffix cgo -o ai-bot
