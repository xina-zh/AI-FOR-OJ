# syntax=docker/dockerfile:1.7

FROM golang:1.22-alpine AS builder

ARG GOPROXY=https://goproxy.cn,direct
ARG GOSUMDB=sum.golang.google.cn

ENV GOPROXY=${GOPROXY}
ENV GOSUMDB=${GOSUMDB}
ENV GOMODCACHE=/go/pkg/mod
ENV GOCACHE=/root/.cache/go-build

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/server ./cmd/server

FROM docker:cli

WORKDIR /app

COPY --from=builder /bin/server /app/server

EXPOSE 8080

CMD ["/app/server"]
