# GopherMart

Сервис лояльности на Go, реализованный по **Clean Architecture** с модульной (вертикальной) организацией.

## Архитектурный стиль

Проект использует:

- **Clean Architecture**: зависимости направлены внутрь (presentation/adapters -> application -> domain).
- **Modular Monolith (vertical modules)**: бизнес-разделение по модулям `identity`, `orders`, `balance`.
- **Intermodule contracts**: межмодульные вызовы идут только через публичные API-контракты модуля-провайдера.

Ключевые директории:

```text
app/cmd/gophermart/bootstrap/           Composition root: wiring, router, server lifecycle
app/internal/gophermart/modules/        Вертикальные модули: identity, orders, balance
app/internal/gophermart/application/    Shared kernel: errors/retry + общие infra-порты
app/internal/gophermart/adapters/       Общие инфраструктурные адаптеры
app/internal/gophermart/presentation/   Общий HTTP слой (middleware, httpcontext)
app/tests/contract/                     Contract tests межмодульных API-контрактов
app/tests/e2e/                          E2E/API flow тесты
```

Подробнее см. `ARCHITECTURE.md`.

## Запуск

### Требования

- Go 1.24+
- PostgreSQL
- Docker (для integration/e2e тестов)

### Команды

```bash
# Сборка
make build

# Запуск приложения
make run

# Запуск миграций
cd app && go run ./cmd/migrate -d "postgres://user:pass@localhost:5432/gophermart?sslmode=disable"

# Запуск с флагами
cd app && ./bin/gophermart -a 127.0.0.1:8080 -d "postgres://user:pass@localhost:5432/gophermart?sslmode=disable" -r http://127.0.0.1:8081

# Запуск с кастомным config
cd app && ./bin/gophermart --config ./configs/gophermart.yaml
```

## Конфигурация

Загрузка конфигурации выполняется в порядке:

1. CLI flags
2. ENV
3. YAML (`app/configs/gophermart.yaml`)
4. defaults

`app/configs/gophermart.yaml` содержит baseline параметры, секреты задаются через ENV.

Обязательные переменные:

- `DATABASE_URI`
- `JWT_SECRET`

### ENV / Flags

| Переменная | Флаг | Назначение |
|---|---|---|
| `RUN_ADDRESS` | `-a` | адрес HTTP сервера |
| `DATABASE_URI` | `-d` | DSN PostgreSQL |
| `ACCRUAL_SYSTEM_ADDRESS` | `-r` | адрес сервиса начислений |
| `JWT_SECRET` | `-s` | секрет подписи JWT |
| `JWT_TTL` | `-t` | TTL JWT |
| `LOG_LEVEL` | `-l` | уровень логирования |
| `BCRYPT_COST` | `--bcrypt-cost` | стоимость bcrypt |
| `DB_MAX_CONNS` | - | лимиты пула БД |
| `DB_MIN_CONNS` | - | лимиты пула БД |
| `DB_MAX_CONN_LIFE` | - | лимиты пула БД |
| `DB_MAX_CONN_IDLE` | - | лимиты пула БД |
| `DB_HEALTH_CHECK` | - | лимиты пула БД |
| `DB_RETRY_MAX_RETRIES` | - | retry БД |
| `DB_RETRY_BASE_DELAY` | - | retry БД |
| `DB_RETRY_MAX_DELAY` | - | retry БД |
| `ACCRUAL_POLL_INTERVAL` | - | интервал воркера accrual |
| `ACCRUAL_HTTP_TIMEOUT` | - | таймаут HTTP клиента accrual |
| `ACCRUAL_BATCH_SIZE` | - | размер батча accrual |
| `ACCRUAL_MAX_WORKERS` | - | число воркеров accrual |
| `OPTIMISTIC_RETRIES` | - | retry optimistic lock |

### Локальный `.env`

```bash
cp app/.env.example app/.env
make run
```

## Тесты

```bash
# Unit/package tests
make test-unit

# Contract tests
make test-contract

# Integration tests (repository-level, real PostgreSQL)
make test-integration

# E2E tests
make test-e2e

# Full suite
make test-all
```

## API эндпоинты

- `POST /api/user/register`
- `POST /api/user/login`
- `POST /api/user/orders` (auth)
- `GET /api/user/orders` (auth)
- `GET /api/user/balance` (auth)
- `POST /api/user/balance/withdraw` (auth)
- `GET /api/user/withdrawals` (auth)
