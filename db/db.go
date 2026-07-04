package db

import (
	"context"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"log"
	"os"
)

const scheme = `CREATE TABLE IF NOT EXISTS users (
         	service_name TEXT NOT NULL,
			price NUMERIC CHECK (price > 0),
			user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			start_date TEXT NOT NULL,
			end_date TEXT
        );`

func Init() error {
	if err := godotenv.Load(); err != nil {
		log.Println("env файл не найден, используем системные переменные")
	}
	log.Println("env файл найден")
	db, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Невозможно подключиться к базе данных: %v", err)
	}
	log.Println("Подключение к базе данных успешно")
	defer db.Close(context.Background())

	m, err := migrate.New("file://migrations", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Ошибка создания миграций: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Ошибка применения миграций: %v", err)
	}
	log.Println("Миграции успешно применены")
	return nil
}
