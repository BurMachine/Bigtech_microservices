# Этап сборки
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git make protobuf-dev build-base

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .



# Ручная сборка бинарника (обход ошибки в make build)
RUN go build -o /app/bin/auth_service ./cmd/auth_service/main.go

# Финальный этап (минимальный образ)
FROM alpine:3.20

COPY --from=builder /app/bin /usr/local/bin

# Порт из вашего compose (8081)
EXPOSE 8081

CMD ["/usr/local/bin/auth_service"]