# GoShop Backend - Техническое задание

## Описание проекта
E-commerce платформа на Go с админской панелью, кэшированием и аналитикой.

## Архитектура
- **Language**: Go
- **Database**: PostgreSQL
- **Cache**: Redis
- **Auth**: JWT + 2FA
- **API**: REST API с Swagger документацией

## Базовый функционал

### Аутентификация и авторизация
- Регистрация пользователей (email, пароль, имя, телефон)
- Логин/логаут с JWT токенами
- Система ролей (админ, пользователь)
- Middleware для проверки прав доступа
- Восстановление пароля

### Управление пользователями
- Просмотр/редактирование профиля
- Добавление/редактирование адресов доставки
- Список всех адресов пользователя
- Удаление адресов

### Каталог товаров
- Создание/редактирование/удаление товаров (только админ)
- Получение списка товаров с пагинацией
- Фильтрация по категориям, цене
- Поиск товаров по названию
- Просмотр детальной информации о товаре
- Управление остатками на складе

### Категории
- Создание/редактирование/удаление категорий (только админ)
- Получение списка всех категорий
- Привязка товаров к категориям (many-to-many)

### Корзина
- Добавление товаров в корзину
- Изменение количества товаров
- Удаление товаров из корзины
- Просмотр содержимого корзины
- Очистка корзины
- Подсчет общей стоимости

### Заказы
- Создание заказа из корзины
- Просмотр истории заказов пользователя
- Изменение статуса заказа (только админ)
- Отмена заказа
- Просмотр деталей заказа
- Список всех заказов (только админ)

### Отзывы
- Добавление отзыва к товару (только после покупки)
- Редактирование своего отзыва
- Удаление отзыва (автор или админ)
- Просмотр всех отзывов к товару
- Средний рейтинг товара

## Кэширование (Redis)

### Стратегия кэширования
- Кэш сессий пользователей
- Кэш популярных товаров (топ-10, новинки)
- Кэш результатов поиска
- Кэш категорий и их товаров
- Кэш корзин пользователей
- TTL для разных типов данных
- Инвалидация кэша при изменениях

### TTL настройки
- Сессии: 24 часа
- Товары: 1 час
- Поиск: 15 минут
- Категории: 6 часов

## Файлы и медиа

### Изображения товаров
- Загрузка множественных изображений на товар
- Валидация файлов (размер, формат, безопасность)
- Генерация thumbnails разных размеров
- Сжатие изображений
- API для получения изображений с параметрами

### Аватары пользователей
- Загрузка и обновление аватаров
- Автоматическое сжатие
- Fallback на дефолтный аватар

### Хранение
- Файловая система или S3-совместимое хранилище
- Организация по папкам (users/, products/)

## Двухфакторная аутентификация (2FA)

### Функционал
- Включение/отключение 2FA в профиле
- Генерация QR-кода для authenticator apps
- Backup коды для восстановления доступа
- Проверка TOTP при логине
- Принудительная 2FA для админов

### Поддерживаемые приложения
- Google Authenticator
- Authy
- Microsoft Authenticator

## Админская панель с аналитикой

### Dashboard с графиками
- Продажи по дням/неделям/месяцам (линейный график)
- Топ товаров (bar chart)
- Распределение заказов по статусам (pie chart)
- Регистрации пользователей по времени
- Средний чек и количество заказов
- Конверсия из корзины в заказ

### Управление
- CRUD для всех сущностей
- Bulk операции (массовое изменение цен)
- Импорт/экспорт товаров (CSV/Excel)
- Модерация отзывов
- Управление пользователями (блокировка, смена ролей)

### Отчеты
- Экспорт отчетов в PDF/Excel
- Настраиваемые периоды
- Фильтрация по категориям, пользователям
- Отчет по остаткам на складе

## Уведомления (будущие фичи)

### Email уведомления
- При регистрации
- При смене статуса заказа
- Шаблоны писем
- Очередь для отправки (async)

## Безопасность

### Защита API
- Rate limiting (по IP, по пользователю)
- Логирование всех действий админов
- Валидация и санитизация всех inputs
- CORS настройки
- SQL injection защита

### Аудит
- Логирование критических операций
- История изменений для товаров
- Tracking изменений цен

## Мониторинг и логирование

### Health checks
- Database connectivity
- Redis connectivity
- Service health endpoint

### Metrics
- Metrics для Prometheus
- Custom business metrics
- Performance monitoring

### Logging
- Structured logging (JSON)
- Log levels (debug, info, warn, error)
- Request/response logging middleware

## API Structure

### Endpoints

#### 🟢 Публичные (без авторизации)
```
Authentication:
POST /api/v1/auth/register
POST /api/v1/auth/login

Products (readonly):
GET /api/v1/products
GET /api/v1/products/:id

Categories (readonly):
GET /api/v1/categories

Reviews (readonly):
GET /api/v1/products/:id/reviews

Health:
GET /api/v1/health
```

#### 🔒 Требуют авторизации (любой пользователь)
```
Authentication:
POST /api/v1/auth/logout
POST /api/v1/auth/2fa/enable
POST /api/v1/auth/2fa/verify

User Profile:
GET /api/v1/profile
PUT /api/v1/profile
POST /api/v1/profile/avatar
GET /api/v1/addresses
POST /api/v1/addresses
PUT /api/v1/addresses/:id
DELETE /api/v1/addresses/:id

Cart:
GET /api/v1/cart
POST /api/v1/cart/items
PUT /api/v1/cart/items/:productId
DELETE /api/v1/cart/items/:productId
DELETE /api/v1/cart

Orders:
POST /api/v1/orders
GET /api/v1/orders
GET /api/v1/orders/:id
DELETE /api/v1/orders/:id

Reviews:
POST /api/v1/reviews
PUT /api/v1/reviews/:id (только автор)
DELETE /api/v1/reviews/:id (автор или админ)
```

#### 🔐 Только для админов
```
Products (management):
POST /api/v1/products
PUT /api/v1/products/:id
DELETE /api/v1/products/:id
POST /api/v1/products/:id/images

Categories (management):
POST /api/v1/categories
PUT /api/v1/categories/:id
DELETE /api/v1/categories/:id

Orders (management):
PUT /api/v1/orders/:id/status
GET /api/v1/admin/orders (все заказы)

User Management:
GET /api/v1/admin/users
PUT /api/v1/admin/users/:id/role
PUT /api/v1/admin/users/:id/status

Analytics:
GET /api/v1/admin/dashboard
GET /api/v1/admin/reports/sales
GET /api/v1/admin/reports/users
GET /api/v1/admin/reports/products

Reviews (moderation):
DELETE /api/v1/reviews/:id (любой отзыв)
```

### Response format
```json
{
  "success": true,
  "data": {},
  "message": "Success",
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100
  }
}
```

### Error format
```json
{
  "success": false,
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {}
}
```

## Technical Requirements

### Performance
- Response time < 200ms for cached requests
- Response time < 500ms for database queries
- Support for 1000+ concurrent users
- Database connection pooling

### Scalability
- Horizontal scaling ready
- Stateless design
- Load balancer friendly
- Graceful shutdown

### Documentation
- OpenAPI/Swagger specification
- API documentation
- Code documentation
- Deployment guide

## Database Schema
См. init.sql для полной схемы базы данных.

## Configuration
- Environment-based configuration
- Support for .env files
- Docker-friendly setup
- Config validation