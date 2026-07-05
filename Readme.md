# Screenshot Go Service

Сервис для создания скриншотов веб-страниц с авторизацией по API-ключам и оплатой через Bitcoin.

## Быстрый старт

```bash
cp .env.example .env
# Отредактируйте .env — заполните токены и пароли

docker compose up -d --build
```

Всё. API будет доступно на `http://localhost:8082`.

## API

### Регистрация
```bash
curl -X POST http://localhost:8082/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret"}'
```
Ответ содержит `api_key.key` — сохраните его.

### Авторизация (получить ключ повторно)
```bash
curl -X POST http://localhost:8082/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret"}'
```

### Скриншот
```bash
curl "http://localhost:8082/screenshot?url=https://example.com" \
  -H "X-API-Key: <ваш_ключ>" \
  --output screenshot.png
```

### Создать инвойс для пополнения баланса
```bash
curl -X POST http://localhost:8082/payments/create \
  -H "X-API-Key: <ваш_ключ>" \
  -H "Content-Type: application/json" \
  -d '{"amount_usd": 5.00, "payment_method": "bitcoin", "promo_code": "WELCOME10"}'
```

### Отменить инвойс
```bash
curl -X POST http://localhost:8082/payments/cancel \
  -H "X-API-Key: <ваш_ключ>" \
  -H "Content-Type: application/json" \
  -d '{"address": "<btc_address>"}'
```

### Создать промо-код
```bash
curl -X POST http://localhost:8082/payments/promos/create \
  -H "X-API-Key: <ваш_ключ>" \
  -H "Content-Type: application/json" \
  -d '{"code": "WELCOME10", "discount_percent": 10, "max_uses": 100, "expires_at": "2030-01-01T00:00:00Z"}'
```

### Подтвердить платёж (внутренний роут)
```bash
curl -X POST "http://localhost:8082/internal/confirm-payment?address=<btc_address>"
```
> ⚠️ В продакшене добавьте проверку `X-Internal-Secret` заголовка.

### Переменные окружения для Bitcoin RPC
```bash
BITCOIN_RPC_USER=your_user
BITCOIN_RPC_PASSWORD=your_password
BITCOIN_RPC_HOST=127.0.0.1
BITCOIN_RPC_PORT=8332
```

## Порты
| Сервис | Порт |
|---|---|
| API | 8082 |
| Browserless | 3002 |
| PostgreSQL | 5433 |

## Миграции
Запускаются автоматически при старте сервиса. Файлы находятся в `app/migrations/`.

## Тестирование
Перед деплоем обязательно запускайте полный набор проверок.

### Быстрый запуск тестов
```bash
cd app
go test ./...
```

### Тесты промо-кодов
```bash
cd app
go test ./... -run Promo
```

### Форматирование и сборка
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

## Структура проекта
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