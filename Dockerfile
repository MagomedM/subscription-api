# Используем официальный образ Go для сборки
FROM golang:1.25-alpine AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы зависимостей (это ускоряет сборку)
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальной код
COPY . .

# Устанавливаем утилиту swag и генерируем документацию
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g main.go -o ./docs

# Собираем приложение
RUN go build -o /app/main .

# Финальный, минимальный образ
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Копируем бинарный файл и папки с миграциями и документацией
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/docs ./docs

# Открываем порт, который использует ваше приложение
EXPOSE 8080

# Запускаем приложение
CMD ["./main"]
