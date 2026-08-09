# Subscription Service

REST API для управления подписками с расчетом стоимости.

## Технологии

- Go 1.25
- PostgreSQL 16
- pgx v5
- Docker & Docker Compose

## Запуск
\`\`\`bash
make run  # или docker-compose up -d
\`\`\`

## API Endpoints
| Method | Endpoint                                              | Description                 |
|--------|-------------------------------------------------------|-----------------------------|
| GET    | /api/subscription?page=1&limit=20                     | Список подписок (пагинация) |
| POST   | /api/subscription                                     | Создать подписку            |
| GET    | /api/subscription/{id}                                | Получить подписку           |
| GET    | /api/subscription?end_date=06-2026&start_date=01-2026 | Расчет суммы за период      |
| PUT    | /api/subscription       | Обновление подписки         |
| DELETE | /api/subscription       | Удаление подписки           |

## Примеры запросов
\`\`\`bash
curl -X GET "http://localhost:8080/api/subscription?page=1&limit=10"
\`\`\`

## Структура проекта
\`\`\`
├── cmd/           # Точки входа
├── internal/      # Приватные пакеты
│   ├── api/       # HTTP handlers
│   ├── service/   # Бизнес-логика
│   ├── repository/# Работа с БД
│   └── models/    # DTO/Entity
├── migrations/    # SQL-миграции
└── docker-compose.yml
\`\`\`

## Переменные окружения
Создайте файл `.env`:
\`\`\`
DATABASE_URL=postgres://postgres:121334579@localhost:5432/subdata?sslmode=disable
DB_PASSWORD=postgres
\`\`\`