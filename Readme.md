# Screenshot Go Service

Сервис для создания скриншотов веб-страниц с авторизацией по API-ключам и оплатой через Bitcoin*.
*В процессе разработки

## Quick start

```bash
cp .env.example .env
# Edit .env — fill the tokens and passwords

docker compose up -d --build
```

That's all. API is available in `http://localhost:8082`.

## API

### Registration
```bash
curl -X POST http://localhost:8082/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret"}'
```
Ответ содержит `api_key.key` — сохраните его.

### Auth (get the key again)
```bash
curl -X POST http://localhost:8082/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret"}'
```

### Screenshot
```bash
curl "http://localhost:8082/screenshot?url=https://example.com" \
  -H "X-API-Key: <ваш_ключ>" \
  --output screenshot.png
```

### Create invoice for top-up the balance
```bash
curl -X POST http://localhost:8082/payments/create \
  -H "X-API-Key: <ваш_ключ>" \
  -H "Content-Type: application/json" \
  -d '{"amount_usd": 5.00, "payment_method": "bitcoin", "promo_code": "WELCOME10"}'
```

### Cancel payment
```bash
curl -X POST http://localhost:8082/payments/cancel \
  -H "X-API-Key: <ваш_ключ>" \
  -H "Content-Type: application/json" \
  -d '{"address": "<btc_address>"}'
```

### Create PROMO-CODE
```bash
curl -X POST http://localhost:8082/payments/promos/create \
  -H "X-API-Key: <ваш_ключ>" \
  -H "Content-Type: application/json" \
  -d '{"code": "WELCOME10", "discount_percent": 10, "max_uses": 100, "expires_at": "2030-01-01T00:00:00Z"}'
```

### Confirm payment (внутренний роут)
```bash
curl -X POST "http://localhost:8082/internal/confirm-payment?address=<btc_address>"
```
> ⚠️ В продакшене добавьте проверку `X-Internal-Secret` заголовка.

### Bitcoin RPC environment variables (.env)
```bash
BITCOIN_RPC_USER=your_user
BITCOIN_RPC_PASSWORD=your_password
BITCOIN_RPC_HOST=127.0.0.1
BITCOIN_RPC_PORT=8332
```

## Ports
| Service | Port |
|---|---|
| API | 8082 |
| Browserless | 3002 |
| PostgreSQL | 5433 |

## Migrations
Запускаются автоматически при старте сервиса. Файлы находятся в `app/migrations/`.

## Тестирование
Перед деплоем обязательно запускайте полный набор проверок.

### Fast tests launch
```bash
cd app
go test ./...
```

### Promo codes tests
```bash
cd app
go test ./... -run Promo
```

### Build
```bash
cd app
go fmt ./...
go build ./...
```

### Полная pre-deploy проверка
```powershell
cd app
./run-predeploy-checks.ps1
```

Этот скрипт выполняет:
- форматирование Go-кода,
- запуск всех тестов,
- сборку проекта.

## Project's structure
```
.
├── docker-compose.yaml
├── .env.example
└── app/
    ├── Dockerfile
    ├── main.go
    ├── go.mod
    ├── go.sum
    ├── config/
    │   └── config.go
    ├── handler/
    │   ├── auth.go
    │   └── payment.go
    ├── middleware/
    │   └── auth.go
    ├── migrations/
    │   ├── 000001_create_users.{up,down}.sql
    │   ├── 000002_create_api_keys.{up,down}.sql
    │   ├── 000003_create_invoices.{up,down}.sql
    │   └── 000004_create_btcaddresses.{up,down}.sql
    ├── model/
    │   └── user.go
    └── storage/
        └── postgres.go
```
# Integration tests

The integration suite runs migrations against a dedicated PostgreSQL container. It does not use the application's `postgres` service or its data volume.

```powershell
docker compose -p screenshot-go-integration -f docker-compose.integration.yml up --build --abort-on-container-exit --exit-code-from integration-tests
docker compose -p screenshot-go-integration -f docker-compose.integration.yml down -v
```
