# Jobify — IT Job Platform Backend

Go REST API для IT-джоб-платформы. Чистая архитектура, PostgreSQL + Redis.

**Стек:** Go 1.22, Chi v5, pgx/v5, go-redis, JWT, bcrypt, golang-migrate, Swagger.

---

## Что уже работает

- Регистрация и логин (JWT, bcrypt).
- Профиль разработчика: скиллы, опыт, зарплатная вилка, remote-only.
- Джобы: CRUD для админа, публичный список с фильтрами (skills, remote, salary), пагинация.
- Skill Match %: для каждого джоба считается процент совпадения с твоими скиллами + списки matched/missing.
- Apply Tracker: сохраняешь джоб → saved → applied → interview → offer/rejected. Своими заявками управляешь только ты (проверка владельца).
- Кэш листинга джобов в Redis (TTL 10 мин).
- Background worker: раз в 6 часов тянет джобы с Remotive API и складывает в БД.
- Rate limit: 60 rpm на IP (меняется через `RATE_LIMIT_PER_MIN`), Redis `INCR + EXPIRE`.
- Graceful shutdown (SIGINT/SIGTERM → 10s на дренаж).
- Swagger UI на `/swagger/index.html`.

---

## Как запустить

### Через Docker (быстро)

```bash
cp .env.example .env
# заполни JWT_SECRET в .env — без него compose не стартует
make docker-up-build
```

Поднимет Postgres, Redis и API на `:8080`. Миграции надо накатить отдельно:

```bash
make migrate-up
```

### Локально (без докера)

Нужны живые Postgres и Redis.

```bash
cp .env.example .env
# заполни DATABASE_URL, REDIS_ADDR, JWT_SECRET
make migrate-up
make run
```

### Проверить

- Health: `curl http://localhost:8080/health`
- Swagger: http://localhost:8080/swagger/index.html
- Базовый путь API: `/api/v1`

---

## Полезные команды

```bash
make run              # поднять API локально
make test             # go test ./... -v -cover
make migrate-up       # накатить миграции
make migrate-down     # откатить
make swag             # перегенерить docs/ по аннотациям
make docker-up        # поднять compose (без пересборки)
make docker-down      # остановить compose
```

---

## Переменные окружения

| Переменная | Обязательна | Дефолт | Назначение |
|---|---|---|---|
| `DATABASE_URL` | да | — | DSN Postgres |
| `JWT_SECRET` | да | — | секрет для подписи JWT |
| `REDIS_ADDR` | нет | `localhost:6379` | адрес Redis |
| `HTTP_PORT` | нет | `8080` | порт API |
| `RATE_LIMIT_PER_MIN` | нет | `60` | лимит запросов на IP в минуту |
| `REMOTIVE_API_URL` | нет | `https://remotive.com/api/remote-jobs` | источник для воркера |

---

## Структура

```
cmd/api/            — entrypoint
internal/
  config/           — env
  domain/           — модели и доменные ошибки
  handler/          — HTTP, роуты
  middleware/       — auth, logger, rate_limit
  repository/
    postgres/       — pgx реализации
    redis/          — кэш и rate-limit storage
  service/          — бизнес-логика
  worker/           — фоновый job aggregator
pkg/                — jwt, hasher, response, validator
migrations/         — SQL up/down
docker/             — Dockerfile, compose
docs/               — сгенерированный swagger
```
