# Go Microservices

Микросервисная архитектура для системы управления заказами и складом (LOMS - Logistics and Order Management System).

## 📋 Описание

Проект представляет собой набор микросервисов на Go, реализующих функционал интернет-магазина:
- Управление корзиной товаров
- Создание и отслеживание заказов
- Управление складскими остатками
- Система уведомлений

## 🏗️ Архитектура

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│    Cart     │────▶│    LOMS     │◀────│  Notifier   │
│  (HTTP/8082)│     │(gRPC/50051) │     │  (Kafka)    │
└─────────────┘     └──────┬──────┘     └─────────────┘
                           │
                    ┌──────▼──────┐
                    │  PostgreSQL │
                    │   (5432)    │
                    └─────────────┘
```

### Сервисы

| Сервис | Порт | Протокол | Описание |
|--------|------|----------|----------|
| **Cart** | 8082 | HTTP REST | Управление корзиной пользователя |
| **LOMS** | 50051, 8081 | gRPC + HTTP Gateway | Логистика и управление заказами |
| **Notifier** | - | Kafka Consumer | Отправка уведомлений о заказах |

## 🛠️ Технологический стек

- **Язык**: Go 1.22
- **Базы данных**: PostgreSQL 16
- **Message Broker**: Apache Kafka
- **API**: gRPC + REST (gRPC-Gateway)
- **Мониторинг**: Prometheus + Grafana
- **Трассировка**: Jaeger (OpenTelemetry)
- **Контейнеризация**: Docker + Docker Compose

## 📦 Структура проекта

```
.
├── api/                    # Proto-контракты
│   └── loms/v1/loms.proto
├── cart/                   # Сервис корзины
│   ├── cmd/server/
│   ├── internal/app/
│   └── internal/adapters/
├── loms/                   # Сервис заказов
│   ├── cmd/server/
│   ├── internal/app/
│   ├── migrations/
│   └── sqlc/
├── notifier/               # Сервис уведомлений
│   └── cmd/notification_server/
├── prometheus/             # Конфигурация Prometheus
├── grafana/                # Конфигурация Grafana
└── docker-compose.yml
```

## 🚀 Быстрый старт

### Требования

- Docker и Docker Compose
- Go 1.22+ (для локальной разработки)

### Запуск через Docker Compose

```bash
# Запуск всех сервисов
docker-compose up --build -d

# Просмотр логов
docker-compose logs -f

# Остановка
docker-compose down
```

### Локальная разработка

```bash
# Запуск LOMS
cd loms
export DB_CONN="postgres://user:password@localhost:5432/dbname"
go run ./cmd/server

# Запуск Cart
cd cart
go run ./cmd/server
```

## 📡 API

### Cart Service (HTTP, порт 8082)

| Метод | Endpoint | Описание |
|-------|----------|----------|
| POST | `/user/{user_id}/cart/{sku_id}` | Добавить товар в корзину |
| DELETE | `/user/{user_id}/cart/{sku_id}` | Удалить товар из корзины |
| DELETE | `/user/{user_id}/cart` | Очистить корзину |
| GET | `/user/{user_id}/cart/list` | Получить содержимое корзины |
| POST | `/user/{user_id}/cart/checkout` | Оформить заказ |
| GET | `/metrics` | Prometheus метрики |

### LOMS Service (gRPC порт 50051, HTTP порт 8081)

| Метод | Endpoint | Описание |
|-------|----------|----------|
| POST | `/v1/order/create` | Создать заказ |
| GET | `/v1/order/{order_id}/info` | Получить информацию о заказе |
| POST | `/v1/order/{order_id}/pay` | Подтвердить оплату заказа |
| DELETE | `/v1/order/{order_id}/cancel` | Отменить заказ |
| GET | `/v1/stocks/{sku_id}/info` | Получить остатки товара |

## 🔧 Конфигурация

### Переменные окружения

| Переменная | Сервис | Описание |
|------------|--------|----------|
| `DB_CONN` | LOMS | Connection string для PostgreSQL |
| `POSTGRES_USER` | PostgreSQL | Пользователь БД |
| `POSTGRES_PASSWORD` | PostgreSQL | Пароль БД |
| `POSTGRES_DB` | PostgreSQL | Имя БД |

## 📊 Мониторинг

| Сервис | URL | Описание |
|--------|-----|----------|
| Prometheus | http://localhost:9090 | Сбор и хранение метрик |
| Grafana | http://localhost:3000 | Визуализация метрик |
| Jaeger | http://localhost:16686 | Distributed tracing |
| Kafka UI | http://localhost:8080 | Управление Kafka |

## 🧪 Тестирование

```bash
# Запуск unit-тестов
make unit-test

# Запуск e2e тестов
make e2e-test

# Запуск линтера
make lint

# Автоисправление линтера
make lint-fix
```

## 📈 Профилирование

Сервисы поддерживают pprof для профилирования:

```bash
# Сбор профиля
make pprof-profile

# Доступные endpoints:
# - /debug/pprof/
# - /debug/pprof/profile
# - /debug/pprof/trace
# - /debug/pprof/heap
```

## 🔗 Kafka Topics

| Topic | Описание |
|-------|----------|
| `loms.order-events` | События заказов для нотификации |
