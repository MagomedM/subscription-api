package repository

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"log"
	"math"
	"os"
	"projects_for_goland/internal/models"
	"strconv"
)

func GetUserId(ctx context.Context, userID string) (*models.SubUser, error) {
	dba, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Невозможно подключиться к базе данных: %v", err)
	}
	defer dba.Close(ctx)
	if err := godotenv.Load(); err != nil {
		log.Println("env файл не найден")
	}
	var u models.SubUser
	row := dba.QueryRow(ctx, "SELECT service_name, price, user_id, start_date, end_date FROM users WHERE user_id = $1", userID)
	err = row.Scan(&u.ServiceName, &u.Price, &u.UserID, &u.StartDate, &u.EndDate)
	if err != nil {
		log.Fatalf("Ошибка получения записи по id: %v", err)
	}
	return &u, nil
}

func GetUserSum(startDateStr, endDateStr, userID, serviceName string, sumtotal float64) (*models.SubSum, error) {
	return &models.SubSum{
		StartDate:   startDateStr,
		EndDate:     endDateStr,
		UserID:      userID,
		ServiceName: serviceName,
		TotalSum:    math.Ceil(sumtotal),
	}, nil
}

func GetSubByDateRange(ctx context.Context, startDateStr, endDateStr, userID, serviceName string) ([]models.SubUser, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("env файл не найден")
	}
	dba, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Невозможно подключиться к базе данных: %v", err)
	}
	defer dba.Close(ctx)
	query := `
        SELECT service_name, price, user_id, start_date, end_date
        FROM users
        WHERE TO_DATE(end_date, 'MM-YYYY') >= $1::DATE 
        AND TO_DATE(start_date, 'MM-YYYY') <= $2::DATE
    `

	args := []interface{}{startDateStr, endDateStr}
	argIdx := 3

	if userID != "" {
		query += ` AND user_id = $` + strconv.Itoa(argIdx) + `::UUID`
		args = append(args, userID)
		argIdx++
	}

	if serviceName != "" {
		query += ` AND service_name ILIKE $` + strconv.Itoa(argIdx)
		args = append(args, "%"+serviceName+"%")
		argIdx++
	}
	rows, err := dba.Query(ctx, query, args...)
	if err != nil {
		log.Fatalf("Ошибка получения подписок: %v", err)
	}
	defer rows.Close()
	var subscriptions []models.SubUser
	for rows.Next() {
		var sub models.SubUser
		err = rows.Scan(&sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate)
		if err != nil {
			log.Fatalf("Ошибка сканирования: %v", err)
		}
		subscriptions = append(subscriptions, sub)
	}
	return subscriptions, nil
}

func GetSubs(ctx context.Context, limit, offset int) ([]*models.SubUser, int64, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("env файл не найден")
	}
	dba, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Невозможно подключиться к базе данных: %v", err)
	}
	defer dba.Close(ctx)

	var total int64
	err = dba.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		log.Fatalf("Ошибка подсчета записей: %v", err)
	}

	row, err := dba.Query(ctx, "SELECT * FROM users ORDER BY user_id LIMIT $1 OFFSET $2", limit, offset)
	defer row.Close()

	var subs []*models.SubUser
	for row.Next() {
		var u models.SubUser
		err = row.Scan(&u.ServiceName, &u.Price, &u.UserID, &u.StartDate, &u.EndDate)
		if err != nil {
			log.Fatalf("Ошибка получения записей: %v", err)
		}
		subs = append(subs, &u)
	}
	return subs, total, nil
}

func PostSub(ctx context.Context, u *models.SubUser) (string, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("env файл не найден")
	}
	dba, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Невозможно подключиться к базе данных: %v", err)
	}
	defer dba.Close(ctx)

	var userID string
	if err != nil {
		log.Fatalf("Ошибка десериализации: %v", err)
	}
	err = dba.QueryRow(ctx, `
		INSERT INTO users(service_name, price, start_date, end_date)
		VALUES($1, $2, $3, $4)
		RETURNING user_id
	`,
		u.ServiceName,
		u.Price,
		u.StartDate,
		u.EndDate,
	).Scan(&userID)
	if err != nil {
		log.Printf("Ошибка создания записи: %v", err)
	}

	return userID, nil
}

func PutSub(ctx context.Context, u *models.SubUser) error {
	if err := godotenv.Load(); err != nil {
		log.Println("env файл не найден")
	}
	dba, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Невозможно подключиться к базе данных: %v", err)
	}
	defer dba.Close(ctx)

	_, err = dba.Exec(ctx, `UPDATE users
    SET service_name = $1, price = $2, start_date = $3, end_date = $4 
    WHERE user_id = $5 
    RETURNING *`, u.ServiceName, u.Price, u.StartDate, u.EndDate, u.UserID)
	if err != nil {
		log.Fatalf("Ошибка обновления записи: %v", err)
	}
	return nil
}

func DelSub(ctx context.Context, id string) error {
	if err := godotenv.Load(); err != nil {
		log.Println("env файл не найден")
	}
	dba, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Невозможно подключиться к базе данных: %v", err)
	}
	defer dba.Close(ctx)

	_, err = dba.Exec(ctx, `DELETE FROM users WHERE user_id = $1`, id)
	if err != nil {
		log.Fatalf("Ошибка удаления записи: %v", err)
	}
	return nil
}
