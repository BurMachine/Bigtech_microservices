# Этап сборки
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git make protobuf-dev build-base 

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Ручная сборка бинарника
RUN go build -o /app/bin/gateway ./cmd/gateway/main.go

# Финальный этап (минимальный образ)
FROM alpine:3.20

COPY --from=builder /app/bin /usr/local/bin

# Порт (8080 для gateway)
EXPOSE 8080

CMD ["/usr/local/bin/gateway"]