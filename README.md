# GopherMart

Система накопления баллов лояльности, реализованная на Go по принципам **Clean Architecture**.

## Архитектура

Проект построен по Clean Architecture со строгим правилом зависимостей.

```
domain/           Чистая бизнес-логика (сущности, value objects, доменные сервисы)
application/      Use cases, порты (интерфейсы), DTO
presentation/     HTTP-хендлеры, middleware, фоновые воркеры
adapters/         PostgreSQL-репозитории, JWT, bcrypt, accrual-клиент, clock
cmd/bootstrap/    Composition root — сборка всех зависимостей
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

# Запуск с флагами
./bin/gophermart -a 127.0.0.1:8080 -d "postgres://user:pass@localhost:5432/gophermart?sslmode=disable" -r http://127.0.0.1:8081

# Или напрямую
make run
```

### Переменные окружения (переопределяют флаги)

| Переменная | Флаг | Описание |
|---|---|---|
| `RUN_ADDRESS` | `-a` | Адрес сервера (по умолчанию `127.0.0.1:8080`) |
| `DATABASE_URI` | `-d` | DSN PostgreSQL |
| `ACCRUAL_SYSTEM_ADDRESS` | `-r` | URL сервиса начислений |
| `JWT_SECRET` | `-s` | Секрет подписи JWT |
| `JWT_TTL` | `-t` | Время жизни токена (по умолчанию `24h`) |
| `LOG_LEVEL` | `-l` | Уровень логирования (по умолчанию `info`) |

### Запуск сервиса начислений (accrual)

В каталоге `cmd/accrual/` находятся готовые бинарники для разных платформ:

```
accrual_darwin_amd64    macOS (Intel)
accrual_darwin_arm64    macOS (Apple Silicon)
accrual_linux_amd64     Linux
accrual_windows_amd64   Windows
```

```bash
# Linux / macOS
chmod +x cmd/accrual/accrual_linux_amd64
./cmd/accrual/accrual_linux_amd64 -a 127.0.0.1:8081

# Windows
cmd\accrual\accrual_windows_amd64 -a 127.0.0.1:8081
```

> Адрес, на котором запущен accrual, передаётся серверу GopherMart через флаг `-r` или переменную `ACCRUAL_SYSTEM_ADDRESS`.

### Тесты

```bash
# Unit-тесты
go test ./...

# Интеграционные тесты (требуется Docker)
go test -tags integration ./...

# Покрытие
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```
