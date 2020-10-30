FROM golang:alpine AS builder
ARG VERSION
RUN apk add --no-cache --virtual .build-deps git libc6-compat build-base
WORKDIR /login_service

COPY . .
RUN go mod download
WORKDIR /login_service/main
RUN go build -gcflags "all=-N -l" -ldflags "-X main.version=${VERSION}" -o app;

FROM alpine:latest
WORKDIR /root
COPY --from=builder /login_service/main/app .

CMD ["./app"]
