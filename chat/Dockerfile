# Этап сборки
FROM golang:1.25-alpine AS builder

# Настройка Go proxy для надежности
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=off

RUN apk add --no-cache git make protobuf-dev build-base

WORKDIR /app

COPY go.mod go.sum ./

# Повторяем загрузку модулей при ошибках сети
RUN go mod download || go mod download || go mod download

COPY . .

# Синхронизируем зависимости
RUN go mod tidy

# Ручная сборка бинарника
RUN go build -o /app/bin/chat_service ./cmd/chat_service/main.go

# Финальный этап (минимальный образ)
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

# Копируем КОНКРЕТНЫЙ файл
COPY --from=builder /app/bin/chat_service /usr/local/bin/

# Порт
EXPOSE 8082

CMD ["/usr/local/bin/chat_service"]