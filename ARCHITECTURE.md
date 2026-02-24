# Gophermart — Architecture Overview

## Project Structure

```
cmd/gophermart/
├── main.go                          # Entry point
└── bootstrap/
    ├── bootstrap.go                 # Run — loads config, logger, DB, migrations, starts app
    ├── app.go                       # NewApp — Composition Root (wires all dependencies)
    ├── factory.go                   # UseCaseFactory implementation
    ├── router.go                    # NewRouter — Gin engine, routes, middleware
    ├── server.go                    # StartServer, WaitForShutdown (graceful)
    ├── postgres.go                  # NewPool — pgxpool.Pool from config
    └── migrate.go                   # RunMigrations — golang-migrate

internal/gophermart/
├── config/
│   ├── config.go                    # Config, LoadConfig (flags + env)
│   └── types.go                     # Address, Duration, BCryptCost custom types
│
├── domain/
│   ├── entity/
│   │   ├── user.go                  # User
│   │   ├── order.go                 # Order, OrderStatus (NEW/PROCESSING/INVALID/PROCESSED)
│   │   ├── balance_account.go       # BalanceAccount (AddAccrual, Withdraw, Version)
│   │   └── withdrawal.go           # Withdrawal
│   ├── vo/
│   │   ├── user_id.go              # UserID
│   │   ├── order_number.go         # OrderNumber + NewOrderNumber
│   │   ├── order_number_validator.go # OrderNumberValidator interface
│   │   ├── points.go               # Points
│   │   └── errors.go               # ErrInvalidOrderNumber
│   └── service/
│       ├── user.go                  # UserService.CreateUser
│       ├── order.go                 # OrderService.CreateOrder
│       ├── balance.go               # BalanceService.CreateAccount
│       └── withdrawal.go           # WithdrawalService.CreateWithdrawal
│
├── application/
│   ├── errors.go                    # Application errors (ErrAlreadyExists, ErrNotFound, etc.)
│   ├── retry.go                     # WithOptimisticRetry helper
│   ├── port/
│   │   ├── usecase.go              # UseCase[In,Out], BackgroundRunner interfaces
│   │   ├── user_repository.go      # UserReader, UserWriter, UserRepository
│   │   ├── order_repository.go     # OrderReader, OrderWriter, OrderRepository
│   │   ├── balance_repository.go   # BalanceAccountReader, Writer, Repository
│   │   ├── withdrawal_repository.go # WithdrawalReader, Writer, Repository
│   │   ├── transactor.go           # Transactor interface
│   │   ├── accrual_client.go       # AccrualClient interface
│   │   ├── token_provider.go       # TokenProvider interface
│   │   ├── password_hasher.go      # PasswordHasher interface
│   │   └── logger.go               # Logger interface
│   ├── dto/
│   │   ├── user.go                 # RegisterInput, LoginInput
│   │   ├── order.go                # UploadOrderInput, OrderOutput
│   │   ├── balance.go              # BalanceOutput
│   │   ├── withdrawal.go           # WithdrawInput, WithdrawalOutput
│   │   └── accrual.go              # AccrualOrderInfo
│   └── usecase/
│       ├── register.go             # RegisterUser (user + balance in TX)
│       ├── login.go                # LoginUser
│       ├── upload_order.go         # UploadOrder (Luhn + conflict check)
│       ├── list_orders.go          # ListOrders
│       ├── get_balance.go          # GetBalance
│       ├── withdraw.go             # Withdraw (optimistic lock + retry)
│       ├── list_withdrawals.go     # ListWithdrawals
│       └── process_accrual.go      # ProcessAccrual (background polling)
│
├── adapters/
│   ├── repository/postgres/
│   │   ├── user.go                 # UserRepository impl
│   │   ├── order.go                # OrderRepository impl (status int<->enum mapping)
│   │   ├── balance.go              # BalanceAccountRepository impl (optimistic lock)
│   │   ├── withdrawal.go           # WithdrawalRepository impl
│   │   ├── transactor.go           # Transactor impl (pgx TX management)
│   │   ├── querier.go              # Querier interface (pool or TX)
│   │   ├── retry.go                # DoWithRetry, RetryConfig, IsRetriable
│   │   └── errors.go               # mapPgError helper
│   ├── accrual/
│   │   └── client.go               # HTTP AccrualClient impl (200/204/429 handling)
│   ├── auth/
│   │   ├── jwt_provider.go         # JWTProvider impl
│   │   └── bcrypt_hasher.go        # BCryptHasher impl
│   ├── logger/
│   │   ├── zap.go                  # ZapLogger impl
│   │   └── nop.go                  # NopLogger (tests)
│   └── validation/
│       └── luhn.go                 # LuhnValidator impl
│
└── presentation/
    ├── http/
    │   ├── handler/
    │   │   ├── user.go             # UserHandler (Register, Login)
    │   │   ├── order.go            # OrderHandler (Upload, List)
    │   │   └── balance.go          # BalanceHandler (Get, Withdraw, ListWithdrawals)
    │   ├── dto/
    │   │   ├── user.go             # RegisterRequest, LoginRequest (+ easyjson)
    │   │   ├── order.go            # OrderResponse (+ easyjson)
    │   │   ├── balance.go          # BalanceResponse (+ easyjson)
    │   │   └── withdrawal.go       # WithdrawRequest, WithdrawalResponse (+ easyjson)
    │   ├── middleware/
    │   │   ├── auth.go             # Auth middleware + BearerTokenExtractor
    │   │   ├── compression.go      # Gzip middleware (sync.Pool)
    │   │   └── logger.go           # Request logging middleware
    │   └── httpcontext/
    │       └── user.go             # UserID() context helper
    ├── worker/
    │   └── accrual.go              # AccrualWorker (background polling, rate limit backoff)
    └── factory/
        └── usecase_factory.go      # UseCaseFactory interface

internal/shared/luhn/
└── luhn.go                          # Luhn algorithm (shared utility)

migrations/gophermart/
├── 000000_create_users              # users table
├── 000001_create_balance_accounts   # balance_accounts table (version column)
├── 000002_create_orders             # orders table (composite index)
├── 000003_create_withdrawals        # withdrawals table
└── 000004_add_updated_at_triggers   # set_updated_at() function + triggers
```

## Dependency Rule (Clean Architecture)

```mermaid
graph TD
    subgraph infrastructure ["Infrastructure — cmd/gophermart/bootstrap"]
        Bootstrap["Run()"]
        NewApp["NewApp() — Composition Root"]
        Router["NewRouter()"]
        FactoryImpl["useCaseFactory"]
        PgPool["pgxpool.Pool"]
        Migrator["golang-migrate"]
        Server["HTTP Server"]
    end

    subgraph presentation ["Presentation — presentation/ (driving adapters)"]
        subgraph handlers [HTTP Handlers]
            UserHandler["UserHandler"]
            OrderHandler["OrderHandler"]
            BalanceHandler["BalanceHandler"]
        end
        subgraph middleware [Middleware]
            AuthMW["Auth"]
            CompressMW["Gzip"]
            LoggerMW["Logger"]
        end
        Worker["AccrualWorker"]
        FactoryIface["UseCaseFactory interface"]
        HTTPDto["HTTP DTOs + easyjson"]
    end

    subgraph application ["Application — application/"]
        subgraph usecases [Use Cases]
            Register["RegisterUser"]
            Login["LoginUser"]
            UploadOrder["UploadOrder"]
            ListOrders["ListOrders"]
            GetBalance["GetBalance"]
            WithdrawUC["Withdraw"]
            ListWithdrawals["ListWithdrawals"]
            ProcessAccrual["ProcessAccrual"]
        end
        subgraph appDto [Application DTOs]
            AppDTOs["RegisterInput, OrderOutput, ..."]
        end
        AppErrors["errors.go + retry.go"]
    end

    subgraph ports ["Application Ports — application/port"]
        UserRepo["UserReader / UserWriter"]
        OrderRepo["OrderReader / OrderWriter"]
        BalanceRepo["BalanceAccountReader / Writer"]
        WithdrawalRepo["WithdrawalReader / Writer"]
        AccrualClient["AccrualClient"]
        Transactor["Transactor"]
        BackgroundRunner["BackgroundRunner"]
        TokenProvider["TokenProvider"]
        PasswordHasher["PasswordHasher"]
        Logger["Logger"]
    end

    subgraph domain ["Domain — domain/"]
        subgraph entities [Entities]
            User["User"]
            Order["Order + OrderStatus"]
            BalanceAccount["BalanceAccount + Version"]
            Withdrawal["Withdrawal"]
        end
        subgraph valueObjects [Value Objects]
            OrderNumber["OrderNumber"]
            Points["Points"]
            UserID["UserID"]
            Validator["OrderNumberValidator"]
        end
        subgraph services [Domain Services]
            UserSvc["UserService"]
            OrderSvc["OrderService"]
            BalanceSvc["BalanceService"]
            WithdrawalSvc["WithdrawalService"]
        end
    end

    subgraph adapters ["Adapters — adapters/ (driven adapters)"]
        subgraph pgRepos [PostgreSQL]
            PgUser["UserRepository"]
            PgOrder["OrderRepository"]
            PgBalance["BalanceAccountRepository"]
            PgWithdrawal["WithdrawalRepository"]
            PgTransactor["Transactor + Querier"]
            PgRetry["DoWithRetry"]
            PgErrors["mapPgError"]
        end
        subgraph external [External]
            AccrualHTTP["accrual.Client"]
        end
        subgraph authAdapters [Auth]
            BCrypt["BCryptHasher"]
            JWT["JWTProvider"]
        end
        subgraph infra [Infra]
            ZapLogger["ZapLogger"]
            Luhn["LuhnValidator"]
        end
    end

    %% Infrastructure wiring
    Bootstrap --> NewApp
    NewApp --> FactoryImpl
    NewApp --> Router
    NewApp --> Server
    Router --> handlers
    Router --> middleware

    %% Presentation -> Application
    handlers -->|"calls via factory"| usecases
    Worker -->|"polls"| ProcessAccrual

    %% Application -> Ports (Dependency Inversion)
    usecases -->|"depends on"| ports
    usecases -->|"uses"| domain

    %% Adapters implement Ports
    PgUser -.->|implements| UserRepo
    PgOrder -.->|implements| OrderRepo
    PgBalance -.->|implements| BalanceRepo
    PgWithdrawal -.->|implements| WithdrawalRepo
    PgTransactor -.->|implements| Transactor
    AccrualHTTP -.->|implements| AccrualClient
    BCrypt -.->|implements| PasswordHasher
    JWT -.->|implements| TokenProvider
    ZapLogger -.->|implements| Logger
    Luhn -.->|implements| Validator

    %% Infrastructure -> Adapters
    PgTransactor --> PgPool
    PgRetry --> PgTransactor
```

## Driving vs Driven Adapters

```mermaid
graph LR
    subgraph driving ["Driving (Primary) — presentation/"]
        direction TB
        H["HTTP Handlers"]
        W["AccrualWorker"]
        MW["Middleware"]
    end

    subgraph core ["Core"]
        direction TB
        UC["Use Cases"]
        P["Ports"]
        D["Domain"]
    end

    subgraph driven ["Driven (Secondary) — adapters/"]
        direction TB
        PG["PostgreSQL Repos"]
        AC["Accrual HTTP Client"]
        AU["JWT + BCrypt"]
        LG["ZapLogger"]
        LH["LuhnValidator"]
    end

    driving -->|"initiates"| core
    core -->|"calls via ports"| driven
```

## HTTP Routes

```mermaid
graph LR
    subgraph global ["Global Middleware"]
        Gzip["Gzip"]
        Log["Logger"]
    end

    subgraph public ["Public"]
        POST_Register["POST /api/user/register"]
        POST_Login["POST /api/user/login"]
    end

    subgraph protected ["Protected (Auth MW)"]
        POST_Orders["POST /api/user/orders"]
        GET_Orders["GET /api/user/orders"]
        GET_Balance["GET /api/user/balance"]
        POST_Withdraw["POST /api/user/balance/withdraw"]
        GET_Withdrawals["GET /api/user/withdrawals"]
    end

    Gzip --> Log
    Log --> public
    Log --> protected

    POST_Register -->|"200 / 400 / 409 / 500"| UserHandler
    POST_Login -->|"200 / 400 / 401 / 500"| UserHandler
    POST_Orders -->|"202 / 200 / 400 / 409 / 422 / 500"| OrderHandler
    GET_Orders -->|"200 / 204 / 500"| OrderHandler
    GET_Balance -->|"200 / 500"| BalanceHandler
    POST_Withdraw -->|"200 / 402 / 422 / 500"| BalanceHandler
    GET_Withdrawals -->|"200 / 204 / 500"| BalanceHandler
```

## Transaction & Concurrency Flows

```mermaid
sequenceDiagram
    participant H as Handler
    participant UC as UseCase
    participant TX as Transactor
    participant R as Repository
    participant DB as PostgreSQL

    Note over H,DB: Register (atomic: user + balance account)
    H->>UC: Execute(RegisterInput)
    UC->>TX: RunInTransaction
    TX->>DB: BEGIN
    UC->>R: userRepo.Create
    R->>DB: INSERT users RETURNING id
    UC->>R: balanceRepo.Create
    R->>DB: INSERT balance_accounts
    TX->>DB: COMMIT
    UC-->>H: userID

    Note over H,DB: Withdraw (optimistic locking + retry)
    H->>UC: Execute(WithdrawInput)
    UC->>UC: WithOptimisticRetry(N)
    loop up to N attempts
        UC->>TX: RunInTransaction
        TX->>DB: BEGIN
        UC->>R: balanceRepo.FindByUserID
        R->>DB: SELECT ... version
        UC->>UC: entity.Withdraw(amount)
        UC->>R: withdrawalRepo.Create
        R->>DB: INSERT withdrawals
        UC->>R: balanceRepo.Update
        R->>DB: UPDATE ... WHERE version=$old
        alt version mismatch
            R-->>UC: ErrOptimisticLock
            UC->>UC: retry
        else success
            TX->>DB: COMMIT
        end
    end
```

## Accrual Worker Flow

```mermaid
sequenceDiagram
    participant W as AccrualWorker
    participant UC as ProcessAccrual
    participant R as OrderRepository
    participant AC as AccrualClient
    participant BR as BalanceRepository
    participant DB as PostgreSQL

    loop every 2s (pollInterval)
        W->>UC: Run(ctx)
        UC->>R: ListByStatuses(NEW, PROCESSING)
        R->>DB: SELECT ... WHERE status IN (0,1) LIMIT batchSize

        loop each order
            UC->>AC: GetOrderAccrual(number)
            AC->>AC: GET /api/orders/{number}

            alt 204 — not registered
                AC-->>UC: nil, nil
                UC->>UC: skip
            else 429 — rate limited
                AC-->>UC: ErrRateLimit{RetryAfter}
                UC-->>W: propagate error
                W->>W: time.After(RetryAfter)
            else 200 — PROCESSING/REGISTERED
                UC->>R: Update(status=PROCESSING)
            else 200 — INVALID
                UC->>R: Update(status=INVALID)
            else 200 — PROCESSED
                UC->>UC: WithOptimisticRetry
                UC->>R: Update(status=PROCESSED, accrual)
                opt accrual > 0
                    UC->>BR: FindByUserID
                    UC->>UC: entity.AddAccrual
                    UC->>BR: Update(balance)
                end
            end
        end
    end
```

## Error Mapping Strategy

```mermaid
graph LR
    subgraph infra ["Infrastructure Errors"]
        PgErr["pgerrcode.UniqueViolation"]
        PgNoRows["pgx.ErrNoRows"]
        Http429["HTTP 429"]
        HttpOther["HTTP 5xx"]
    end

    subgraph app ["Application Errors"]
        AlreadyExists["ErrAlreadyExists"]
        NotFound["ErrNotFound"]
        RateLimit["ErrRateLimit"]
        Conflict["ErrConflict"]
        InsufficientBalance["ErrInsufficientBalance"]
        InvalidOrder["ErrInvalidOrderNumber"]
        InvalidCreds["ErrInvalidCredentials"]
        OptLock["ErrOptimisticLock"]
    end

    subgraph http ["HTTP Status Codes"]
        S200["200 OK"]
        S202["202 Accepted"]
        S400["400 Bad Request"]
        S401["401 Unauthorized"]
        S402["402 Payment Required"]
        S409["409 Conflict"]
        S422["422 Unprocessable"]
        S500["500 Internal"]
    end

    PgErr -->|"adapter maps"| AlreadyExists
    PgNoRows -->|"adapter maps"| NotFound
    Http429 -->|"adapter maps"| RateLimit
    HttpOther -->|"passthrough"| S500

    AlreadyExists -->|"handler maps"| S200
    Conflict -->|"handler maps"| S409
    InvalidOrder -->|"handler maps"| S422
    InvalidCreds -->|"handler maps"| S401
    InsufficientBalance -->|"handler maps"| S402
    NotFound -->|"handler maps"| S500
    OptLock -->|"retry or"| S500
```

## Database Schema

```mermaid
erDiagram
    users {
        bigserial id PK
        text login UK
        text password_hash
        timestamptz created_at
        timestamptz updated_at
    }

    balance_accounts {
        bigserial id PK
        bigint user_id UK,FK
        float8 current
        float8 withdrawn_total
        timestamptz created_at
        timestamptz updated_at
        bigint version
    }

    orders {
        bigserial id PK
        text number UK
        bigint user_id FK
        smallint status
        float8 accrual
        timestamptz uploaded_at
        timestamptz updated_at
        timestamptz processed_at
    }

    withdrawals {
        bigserial id PK
        bigint user_id FK
        text order_number
        float8 amount
        timestamptz processed_at
    }

    users ||--|| balance_accounts : "1:1"
    users ||--o{ orders : "1:N"
    users ||--o{ withdrawals : "1:N"
```
