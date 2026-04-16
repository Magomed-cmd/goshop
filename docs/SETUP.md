# Руководство по установке и настройке

## Требования

### Обязательные
- Go 1.21 или выше
- PostgreSQL 14+
- Redis 7+
- Git

### Опциональные (но рекомендуется)
- Docker и Docker Compose
- Make
- Postman или curl

## Шаги установки

### 1. Клонируйте репозиторий

```bash
git clone <repository-url>
cd goshop
```

### 2. Установите зависимости Go

```bash
go mod download
go mod tidy
```

### 3. Настройка переменных окружения

Создайте файл `.env` в корне проекта:

```bash
cp .env.example .env
```

Отредактируйте `.env` с вашей конфигурацией:

```env
# Сервер
SERVER_PORT=8080
SERVER_HOST=localhost
GIN_MODE=debug

# База данных
DB_HOST=localhost
DB_PORT=5432
DB_USER=goshop_user
DB_PASSWORD=ваш_защищенный_пароль
DB_NAME=goshop_db
DB_SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_DB=0

# JWT
JWT_SECRET=ваш_длинный_секретный_ключ_минимум_32_символа
JWT_EXPIRATION=24h

# S3 / MinIO
S3_ENDPOINT=http://localhost:9000
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_BUCKET=goshop
S3_USE_SSL=false

# OAuth Google
GOOGLE_CLIENT_ID=ваш_google_client_id
GOOGLE_CLIENT_SECRET=ваш_google_client_secret

# Логирование
LOG_LEVEL=info
```

### 4. Настройка БД

#### Вариант A: Ручная настройка

```bash
# Создать БД
createdb -U postgres goshop_db

# Создать пользователя (если нужно)
psql -U postgres -c "CREATE USER goshop_user WITH PASSWORD 'ваш_защищенный_пароль';"
psql -U postgres -c "ALTER USER goshop_user CREATEDB;"

# Предоставить привилегии
psql -U postgres -d goshop_db -c "GRANT ALL PRIVILEGES ON DATABASE goshop_db TO goshop_user;"

# Запустить миграции
make migrate-up
```

#### Вариант B: Docker Compose

```bash
docker-compose up -d postgres redis minio
```

Это запустит:
- PostgreSQL на порте 5432
- Redis на порте 6379
- MinIO на порте 9000

### 5. Миграции БД

```bash
# Создать миграцию
go run cmd/migrate/main.go create create_users_table

# Запустить все миграции
make migrate-up

# Откатить последнюю миграцию
make migrate-down

# Свежий старт (удалить все и пересоздать)
make migrate-fresh
```

### 6. Сгенерировать документацию Swagger

```bash
# Установить swag CLI
go install github.com/swaggo/swag/cmd/swag@latest

# Сгенерировать docs
swag init -g cmd/goshop/main.go
```

### 7. Запустите приложение

#### Режим разработки

```bash
make run
```

Или напрямую:

```bash
go run cmd/goshop/main.go
```

#### В Docker

```bash
# Собрать image
docker build -t goshop:latest .

# Запустить контейнер
docker run -p 8080:8080 --env-file .env goshop:latest
```

#### С Docker Compose

```bash
docker-compose up
```

## Проверка

### Проверьте, что сервер работает

```bash
curl http://localhost:8080/health
```

Ожидаемый ответ:
```json
{
  "status": "ok"
}
```

### Документация Swagger

Откройте в браузере: http://localhost:8080/swagger/index.html

## Рабочий процесс разработки

### Запустить тесты

```bash
# Все тесты
make test

# С покрытием
make coverage

# Конкретный тест
go test -run TestUserRegister ./...

# Watch mode
go test -v ./... --watch
```

### Качество кода

```bash
# Форматирование кода
make fmt

# Линтер
make lint

# Vet
make vet

# Все проверки
make check
```

### Команды БД

```bash
# Создать БД
make db-create

# Удалить БД
make db-drop

# Сбросить БД
make db-reset

# Показать статус миграций
make migrate-status
```

## Полезные команды

### Makefile targets

```bash
run              # Запустить приложение
test             # Запустить тесты
coverage         # Сгенерировать отчет покрытия
fmt              # Форматировать код
lint             # Запустить линтер
vet              # Запустить go vet
build            # Собрать бинарник
migrate-up       # Запустить миграции
migrate-down     # Откатить миграцию
migrate-fresh    # Сбросить БД
db-create        # Создать БД
db-drop          # Удалить БД
db-reset         # Сбросить БД
clean            # Очистить build артефакты
help             # Показать все targets
```

## Тестирование API

### Используя curl

#### Регистрация пользователя
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "name": "John Doe"
  }'
```

#### Вход
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

#### Получить товары
```bash
curl http://localhost:8080/products
```

#### Создать товар (админ)
```bash
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "name": "Product Name",
    "description": "Product description",
    "price": "99.99",
    "stock": 100,
    "category_ids": [1]
  }'
```

### Используя Postman

1. Импортируйте предоставленную Postman коллекцию (если доступна)
2. Установите переменные окружения:
   - `base_url`: http://localhost:8080
   - `token`: Ваш JWT токен с логина
3. Запустите запросы из коллекции

## Решение проблем

### Порт уже занят
```bash
# Найти процесс на порте 8080
lsof -i :8080

# Убить процесс
kill -9 <PID>
```

### Ошибка подключения к БД
```bash
# Проверьте, работает ли PostgreSQL
pg_isready -h localhost -p 5432

# Проверьте Redis
redis-cli ping
```

### Проблемы с миграциями
```bash
# Сбросить миграции
make migrate-fresh

# Показать статус миграций
make migrate-status

# Проверить логи миграций
tail -f /tmp/goshop-migrations.log
```

### Ошибка подключения к Redis
```bash
# Проверить статус Redis
redis-cli info server

# Проверить, что Redis принимает соединения
redis-cli ping
```

### Проблемы с S3/MinIO
```bash
# Проверить, работает ли MinIO
curl http://localhost:9000/minio/health/live

# Проверить учетные данные
aws s3 --endpoint-url http://localhost:9000 ls
```

## Оптимизация производительности

### БД
```bash
# Создать индексы для часто запрашиваемых столбцов
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_products_category ON product_categories(product_id);
CREATE INDEX idx_orders_user ON orders(user_id);
```

### Кэширование
- Redis настроен для кэширования товаров
- TTL: 1 час для товаров
- Ручная инвалидация при обновлении

### Мониторинг
- Включить структурированное логирование
- Мониторить время ответа
- Отслеживать коэффициент попадания кэша

## Развертывание

### Переменные окружения для production

```env
GIN_MODE=release
LOG_LEVEL=warn
DB_SSL_MODE=require
S3_USE_SSL=true
JWT_EXPIRATION=12h
```

### Health checks

```bash
# Проверить здоровье приложения
GET /health

# Проверить БД
GET /health/db

# Проверить Redis
GET /health/redis
```

### Graceful shutdown

Приложение обрабатывает SIGTERM для graceful shutdown:
- Перестает принимать новые запросы
- Ждет завершения текущих запросов
- Закрывает подключения БД
- Закрывает подключения Redis

## Дополнительные ресурсы

- [Go документация](https://golang.org/doc/)
- [Gin Framework](https://gin-gonic.com/)
- [PostgreSQL документация](https://www.postgresql.org/docs/)
- [Redis документация](https://redis.io/documentation)
- [Docker документация](https://docs.docker.com/)

## Следующие шаги

1. Настроить OAuth с Google (опционально)
2. Установить мониторинг и логирование
3. Настроить сервис отправки писем для уведомлений
4. Установить CI/CD pipeline
5. Развернуть в production

## Поддержка

При возникновении проблем или вопросов:
1. Проверьте эту документацию
2. Прочитайте комментарии в коде
3. Проверьте issues проекта
4. Посмотрите логи ошибок
