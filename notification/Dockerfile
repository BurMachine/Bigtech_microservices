# Этап сборки
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git make protobuf-dev build-base

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Синхронизируем зависимости
RUN go mod tidy

# Ручная сборка бинарника
RUN go build -o /app/bin/notification ./cmd/notification/main.go

# Финальный этап (минимальный образ)
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/bin /usr/local/bin

# Порт (8083)
EXPOSE 8083

CMD ["/usr/local/bin/notification"]