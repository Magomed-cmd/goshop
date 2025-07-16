# GoShop Development Roadmap

## 📋 Phase 1: Основа проекта (Week 1-3)

### 🔧 Setup & Infrastructure
- [ ] Создать структуру папок проекта
- [ ] Настроить `go.mod` и зависимости
- [ ] Создать `Makefile`
- [ ] Настроить `docker-compose.yaml`
- [ ] Создать базовый `Dockerfile`

### 🗄️ Database & Config
- [ ] Реализовать `internal/config/config.go` с Viper
- [ ] Создать `internal/db/postgres/connection.go`
- [ ] Реализовать `internal/db/db.go` (главный менеджер)
- [ ] Протестировать подключение к PostgreSQL
- [ ] Создать базовые модели в `internal/models/`

### 🔐 Authentication Core
- [ ] Создать `internal/service/auth.go`
- [ ] Реализовать JWT middleware в `internal/middleware/`
- [ ] Хэширование паролей (bcrypt)
- [ ] Создать `internal/handler/auth.go`
- [ ] Endpoints: `POST /auth/register`, `POST /auth/login`

### 🏗️ Basic Structure
- [ ] Настроить роутер в `internal/routes/`
- [ ] Создать базовую структуру ответов API
- [ ] Middleware для CORS
- [ ] Error handling middleware
- [ ] Логгирование запросов

---

## 🛍️ Phase 2: Core E-commerce (Week 4-7)

### 👥 User Management
- [ ] `internal/repository/user.go` - CRUD операции
- [ ] `internal/service/user.go` - бизнес логика
- [ ] `internal/handler/user.go` - HTTP handlers
- [ ] Endpoints: `GET/PUT /profile`, `GET/POST/PUT/DELETE /addresses`

### 🏷️ Categories
- [ ] `internal/models/category.go`
- [ ] `internal/repository/category.go`
- [ ] `internal/service/category.go`
- [ ] `internal/handler/category.go`
- [ ] Endpoints: `GET /categories`, admin CRUD

### 📦 Products
- [ ] `internal/models/product.go`
- [ ] `internal/repository/product.go` (с поиском и фильтрами)
- [ ] `internal/service/product.go`
- [ ] `internal/handler/product.go`
- [ ] Пагинация и поиск по названию
- [ ] Фильтрация по категориям и цене
- [ ] Endpoints: публичные GET + admin CRUD

### 🛒 Shopping Cart
- [ ] `internal/models/cart.go`
- [ ] `internal/repository/cart.go`
- [ ] `internal/service/cart.go`
- [ ] `internal/handler/cart.go`
- [ ] Логика добавления/изменения товаров
- [ ] Подсчет общей стоимости
- [ ] Endpoints: полный CRUD корзины
    
### 📋 Orders
- [ ] `internal/models/order.go`
- [ ] `internal/repository/order.go`
- [ ] `internal/service/order.go`
- [ ] `internal/handler/order.go`
- [ ] Создание заказа из корзины
- [ ] Управление статусами
- [ ] Endpoints: создание, просмотр, управление статусами

### ⭐ Reviews
- [ ] `internal/models/review.go`
- [ ] `internal/repository/review.go`
- [ ] `internal/service/review.go`
- [ ] `internal/handler/review.go`
- [ ] Проверка права на отзыв (только после покупки)
- [ ] Подсчет среднего рейтинга
- [ ] Endpoints: CRUD отзывов

---

## ⚡ Phase 3: Performance & Caching (Week 8-10)

### 🔴 Redis Setup
- [ ] Создать `internal/db/redis/connection.go`
- [ ] Интегрировать Redis в `internal/db/db.go`
- [ ] Создать `internal/cache/` пакет

### 📦 Product Caching
- [ ] Кэш популярных товаров
- [ ] Кэш результатов поиска
- [ ] Кэш категорий
- [ ] Инвалидация при изменениях

### 🛒 Cart & Session Caching
- [ ] Кэш корзин пользователей
- [ ] Кэш сессий (JWT в Redis)
- [ ] Настройка TTL для разных типов данных

### 🚀 Optimization
- [ ] Оптимизация SQL запросов
- [ ] Добавление индексов в БД
- [ ] Connection pooling настройка
- [ ] Middleware кэширования

### 📊 Monitoring Basic
- [ ] Health check endpoint
- [ ] Graceful shutdown
- [ ] Базовое логирование
- [ ] Metrics подготовка

---

## 🖼️ Phase 4: File Management (Week 11-12)

### 📸 Product Images
- [ ] Создать `internal/storage/` пакет
- [ ] Upload endpoint для изображений товаров
- [ ] Валидация файлов (размер, формат)
- [ ] Генерация thumbnails
- [ ] Сжатие изображений
- [ ] Endpoint: `POST /products/:id/images`

### 👤 User Avatars
- [ ] Upload аватаров пользователей
- [ ] Обработка и сжатие
- [ ] Endpoint: `POST /profile/avatar`

### 🗂️ File Organization
- [ ] Организация файлов по папкам
- [ ] Cleanup старых файлов
- [ ] API для получения изображений разных размеров

---

## 🔐 Phase 5: Advanced Security (Week 13-14)

### 🔒 Two-Factor Authentication
- [ ] Создать `internal/service/totp.go`
- [ ] Генерация секретов и QR-кодов
- [ ] Backup коды
- [ ] Middleware для проверки 2FA
- [ ] Endpoints: `/auth/2fa/enable`, `/auth/2fa/verify`

### 🛡️ Enhanced Security
- [ ] Rate limiting (Redis-based)
- [ ] Audit logging для админов
- [ ] Input sanitization
- [ ] CSRF protection
- [ ] Security headers middleware

---

## 📊 Phase 6: Admin Panel & Analytics (Week 15-18)

### 👨‍💼 Admin Management
- [ ] `internal/handler/admin.go`
- [ ] User management endpoints
- [ ] Bulk operations для товаров
- [ ] Модерация отзывов
- [ ] Role management

### 📈 Analytics Core
- [ ] `internal/service/analytics.go`
- [ ] `internal/repository/analytics.go`
- [ ] Метрики продаж
- [ ] Метрики пользователей
- [ ] Популярные товары

### 📊 Dashboard API
- [ ] `/admin/dashboard` endpoint
- [ ] Данные для графиков продаж
- [ ] Статистика по заказам
- [ ] Конверсия корзина → заказ

### 📄 Reports
- [ ] `internal/service/reports.go`
- [ ] Генерация отчетов
- [ ] Export в CSV/Excel
- [ ] Фильтрация по периодам
- [ ] Endpoints: `/admin/reports/*`

### 📋 Export Functionality
- [ ] CSV экспорт товаров
- [ ] Excel отчеты
- [ ] PDF генерация (опционально)

---

## 🔔 Phase 7: Notifications (Week 19-20)

### 📧 Email System
- [ ] Создать `internal/service/email.go`
- [ ] SMTP конфигурация
- [ ] Шаблоны писем
- [ ] Очередь для отправки

### 📨 Notification Types
- [ ] Email при регистрации
- [ ] Уведомления о смене статуса заказа
- [ ] Восстановление пароля
- [ ] Async отправка

---

## 📚 Phase 8: Documentation & Production (Week 21-22)

### 📖 API Documentation
- [ ] Swagger/OpenAPI интеграция
- [ ] Документация всех endpoints
- [ ] Postman collection
- [ ] README обновление

### 🚀 Production Ready
- [ ] Docker optimization
- [ ] Environment configs
- [ ] CI/CD базовая настройка
- [ ] Security checklist

### 🧪 Testing
- [ ] Unit tests для сервисов
- [ ] Integration tests
- [ ] E2E тесты критичных flow
- [ ] Test coverage отчеты

---

## 🎯 Критерии завершения каждой фазы

### ✅ Phase 1 Complete When:
- Можно зарегистрироваться и залогиниться
- База данных подключена и работает
- Конфиг загружается из файла

### ✅ Phase 2 Complete When:
- Можно создать товар, добавить в корзину, оформить заказ
- Все CRUD операции работают
- Базовая валидация есть

### ✅ Phase 3 Complete When:
- API отвечает быстро (< 200ms кэшированные запросы)
- Redis работает и кэширует данные
- Нет проблем с производительностью

### ✅ Phase 4 Complete When:
- Товары имеют изображения
- Пользователи могут загружать аватары
- Все файлы обрабатываются корректно

### ✅ Phase 5 Complete When:
- 2FA работает с Google Authenticator
- API защищен от основных атак
- Rate limiting активен

### ✅ Phase 6 Complete When:
- Админ может получить любую аналитику
- Отчеты генерируются и экспортируются
- Dashboard показывает реальные данные

### ✅ Phase 7 Complete When:
- Email уведомления приходят
- Очередь обрабатывает отправку
- Шаблоны настроены

### ✅ Phase 8 Complete When:
- API полностью задокументирован
- Готов к деплою в продакшн
- Все тесты проходят

---

## 🚨 Важные правила

1. **НЕ переходи к следующей фазе**, пока текущая не завершена на 100%
2. **Коммить код каждый день** - маленькими частями
3. **Тестировать каждую функцию** перед переходом к следующей
4. **Документировать по ходу** - не оставлять на потом
5. **Если застрял на задаче > 4 часов** - пропусти ее и вернись позже