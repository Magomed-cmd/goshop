# Stage 1: Build
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Установить зависимости сборки
RUN apk add --no-cache git make

# Копировать go.mod и go.sum
COPY go.mod go.sum ./

# Загрузить зависимости
RUN go mod download

# Копировать исходный код
COPY . .

# Собрать приложение
RUN make build

# Stage 2: Runtime
FROM alpine:3.18

WORKDIR /app

# Установить зависимости runtime
RUN apk add --no-cache ca-certificates tzdata

# Копировать бинарник из builder stage
COPY --from=builder /app/bin/goshop /app/goshop

# Копировать миграции
COPY migrations/ ./migrations/

# Expose порт
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Запустить приложение
CMD ["./goshop"]
