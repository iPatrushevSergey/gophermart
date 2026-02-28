# GopherMart

Система накопления баллов лояльности, реализованная на Go по принципам **Clean Architecture**.

## Архитектура

Проект построен по Clean Architecture со строгим правилом зависимостей.

```
app/internal/gophermart/domain/         Чистая бизнес-логика (сущности, value objects, доменные сервисы)
app/internal/gophermart/application/    Use cases, порты (интерфейсы), DTO
app/internal/gophermart/presentation/   HTTP-хендлеры, middleware, фоновые воркеры
app/internal/gophermart/adapters/       PostgreSQL-репозитории, JWT, bcrypt, accrual-клиент, clock
app/cmd/gophermart/bootstrap/           Composition root — сборка всех зависимостей
```

## Запуск

### Требования

- Go 1.24+
- PostgreSQL
- Docker (для интеграционных тестов)

### Запуск сервера

```bash
# Сборка
make build

# Запуск с YAML-конфигом по умолчанию (app/configs/gophermart.yaml)
make run

# Запуск с флагами
cd app && ./bin/gophermart -a 127.0.0.1:8080 -d "postgres://user:pass@localhost:5432/gophermart?sslmode=disable" -r http://127.0.0.1:8081

# Запуск с кастомным config файлом
cd app && ./bin/gophermart --config ./configs/gophermart.yaml

# Или напрямую без сборки
cd app && go run -tags=go_json ./cmd/gophermart
```

### Приоритет источников конфигурации

1. Флаги CLI
2. Переменные окружения
3. YAML-файл (`app/configs/gophermart.yaml` по умолчанию)
4. Значения по умолчанию в коде

### Переменные окружения

`app/configs/gophermart.yaml` хранит baseline non-secret конфиг, а ENV используется для секретов и env-specific override.
Обязательные секреты: `DATABASE_URI`, `JWT_SECRET` (приложение завершится с ошибкой при пустых значениях).

| Переменная | Флаг | Описание |
|---|---|---|
| `RUN_ADDRESS` | `-a` | Адрес сервера (по умолчанию `127.0.0.1:8080`) |
| `DATABASE_URI` | `-d` | DSN PostgreSQL |
| `ACCRUAL_SYSTEM_ADDRESS` | `-r` | URL сервиса начислений |
| `JWT_SECRET` | `-s` | Секрет подписи JWT |
| `JWT_TTL` | `-t` | Время жизни токена (по умолчанию `24h`) |
| `LOG_LEVEL` | `-l` | Уровень логирования (по умолчанию `info`) |
| `BCRYPT_COST` | `--bcrypt-cost` | Фактор bcrypt (4-31) |
| `DB_MAX_CONNS` | `-` | Максимум соединений пула |
| `DB_MIN_CONNS` | `-` | Минимум соединений пула |
| `DB_MAX_CONN_LIFE` | `-` | Максимальное время жизни соединения |
| `DB_MAX_CONN_IDLE` | `-` | Максимальный idle соединения |
| `DB_HEALTH_CHECK` | `-` | Период health-check пула |
| `DB_RETRY_MAX_RETRIES` | `-` | Количество retry для DB операций |
| `DB_RETRY_BASE_DELAY` | `-` | Базовая задержка retry |
| `DB_RETRY_MAX_DELAY` | `-` | Максимальная задержка retry |
| `ACCRUAL_POLL_INTERVAL` | `-` | Интервал фонового опроса accrual |
| `ACCRUAL_HTTP_TIMEOUT` | `-` | Таймаут HTTP клиента accrual |
| `ACCRUAL_BATCH_SIZE` | `-` | Размер батча обработки accrual |
| `ACCRUAL_MAX_WORKERS` | `-` | Количество воркеров accrual |
| `OPTIMISTIC_RETRIES` | `-` | Количество retry optimistic lock |

### Локальный `.env`

```bash
# Создать локальный файл (не коммитится)
cp app/.env.example app/.env

# Запустить (приложение подхватит app/.env автоматически, если файл существует)
make run
```

Файл `app/.env.example` содержит только секреты и точечные override-переменные. Реальные значения (`JWT_SECRET`, пароль в `DATABASE_URI`) храните в `app/.env`.

### Запуск сервиса начислений (accrual)

Бинарник accrual не входит в этот репозиторий. Запустите его отдельно и передайте адрес в GopherMart через:
- флаг `-r`
- или переменную `ACCRUAL_SYSTEM_ADDRESS`

> Адрес, на котором запущен accrual, передаётся серверу GopherMart через флаг `-r` или переменную `ACCRUAL_SYSTEM_ADDRESS`.

### Тесты

```bash
# Unit-тесты
make test

# Интеграционные тесты (требуется Docker)
make test-integration

# Покрытие
make cover
```
