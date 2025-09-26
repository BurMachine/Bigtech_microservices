# Этап сборки
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git make protobuf-dev build-base

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Ручная сборка бинарника
RUN go build -o /app/bin/user_service ./cmd/user_service/main.go

# Финальный этап (минимальный образ)
FROM alpine:3.20

COPY --from=builder /app/bin /usr/local/bin

# Порт (8084)
EXPOSE 8084

CMD ["/usr/local/bin/user_service"]