package api

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"os"
	"projects_for_goland/db"
	"strconv"
	"time"
)

func Init() {
	http.HandleFunc("/api/subscription", subHandler)
	http.HandleFunc("/api/subscription_id", getidHandler)
	http.HandleFunc("/api/subscription_summary", getSumHandler)
}

// subHandler обрабатывает все запросы к /api/subscription
// @Summary      Обработчик подписок
// @Description  Маршрутизирует запросы к подпискам в зависимости от метода
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Router       /api/subscription [get]
// @Router       /api/subscription [post]
// @Router       /api/subscription [put]
// @Router       /api/subscription [delete]

func subHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getSubHandler(w, r)
		log.Println("Получение всех записей успешно")
	case http.MethodPost:
		postSubHandler(w, r)
		log.Println("Добавление записи успешно")
	case http.MethodPut:
		putSubHandler(w, r)
		log.Println("Обновление записи успешно")
	case http.MethodDelete:
		deleteSubHandler(w, r)
		log.Println("Удаление записи успешно")
	}
}

// getidHandler возвращает подписку по ID пользователя
// @Summary      Получить подписку по ID
// @Description  Возвращает информацию о подписке по user_id
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        user_id   query     string  true  "ID пользователя"  example(123)
// @Success      200  {object}  db.SubUser  "Успешный ответ"
// @Failure      400  {string}  string  "ID не указан"
// @Failure      404  {string}  string  "Подписка не найдена"
// @Failure      500  {string}  string  "Ошибка сервера"
// @Router       /api/subscription_id [get]
func getidHandler(w http.ResponseWriter, r *http.Request) {
	if err := godotenv.Load(); err != nil {
		log.Println("env файл не найден")
	}
	dba, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Невозможно подключиться к базе данных: %v", err)
	}
	defer dba.Close(context.Background())
	id := r.URL.Query().Get("user_id")
	if id == "" {
		log.Fatalf("Id не указан")
	}
	row := dba.QueryRow(context.Background(), "SELECT service_name, price, user_id, start_date, end_date FROM users WHERE user_id = $1", id)
	var u db.SubUser
	err = row.Scan(&u.ServiceName, &u.Price, &u.UserID, &u.StartDate, &u.EndDate)
	if err != nil {
		log.Fatalf("Ошибка получения записи по id: %v", err)
	}
	writeJson(w, &u)
	log.Println("Получение записи по id успешно")
}

// getSumHandler возвращает сумму подписок за период
// @Summary      Получить сумму подписок за период
// @Description  Возвращает общую сумму подписок за указанный период с возможностью фильтрации
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        startdate   query     string  true  "Дата начала в формате MM-YYYY"  example(01-2026)
// @Param        enddate     query     string  true  "Дата окончания в формате MM-YYYY" example(07-2026)
// @Param        user_id     query     string  false "ID пользователя для фильтрации"  example(123)
// @Param        service     query     string  false "Название сервиса для фильтрации" example(Netflix)
// @Success      200  {object}  map[string]interface{}  "Успешный ответ с суммой"
// @Failure      400  {object}  map[string]string  "Неверный запрос"
// @Failure      500  {object}  map[string]string  "Внутренняя ошибка сервера"
// @Router       /api/subscription_summary [get]
func getSumHandler(w http.ResponseWriter, r *http.Request) {
	if err := godotenv.Load(); err != nil {
		log.Println("env файл не найден")
	}
	dba, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Невозможно подключиться к базе данных: %v", err)
	}
	defer dba.Close(context.Background())
	startstr := r.URL.Query().Get("start_date")
	endstr := r.URL.Query().Get("end_date")
	userID := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")

	const layout = "01-2006"
	stardate, err := time.Parse(layout, startstr)
	if err != nil {
		log.Fatalf("Ошибка парсинга начальной даты: %v", err)
	}
	enddate, err := time.Parse(layout, endstr)
	if err != nil {
		log.Fatalf("Ошибка парсинга конечной даты: %v", err)
	}
	query := `
		SELECT COALESCE(SUM(price), 0)
		FROM users
		WHERE TO_DATE(start_date, 'MM-YYYY') >= $1::DATE AND TO_DATE(end_date, 'MM-YYYY') <= $2::DATE
	`

	argIdx := 3
	args := []interface{}{stardate, enddate}

	if userID != "" {
		query += ` AND user_id = $` + strconv.Itoa(argIdx)
		args = append(args, userID)
		argIdx++
	}

	if serviceName != "" {
		query += ` AND service_name ILIKE $` + strconv.Itoa(argIdx)
		args = append(args, "%"+serviceName+"%") // поиск по частичному совпадению
		argIdx++
	}

	var sumtotal int64
	err = dba.QueryRow(context.Background(), query, args...).Scan(&sumtotal)
	if err != nil {
		log.Fatalf("Ошибка подсчёта суммы: %v", err)
	}
	response := map[string]interface{}{
		"start_date":   startstr,
		"end_date":     endstr,
		"user_id":      userID,
		"service_name": serviceName,
		"total_sum":    sumtotal,
	}
	writeJson(w, response)
	log.Println("Получение суммы всех подписок за выбранный период успешно")
}

// getSubHandler возвращает все подписки
// @Summary      Получить все подписки
// @Description  Возвращает список всех подписок
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Success      200  {object}  db.SubsFull  "Список всех подписок"
// @Failure      500  {string}  string  "Ошибка сервера"
// @Router       /api/subscription [get]
func getSubHandler(w http.ResponseWriter, r *http.Request) {
	if err := godotenv.Load(); err != nil {
		log.Println("env файл не найден")
	}
	dba, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Невозможно подключиться к базе данных: %v", err)
	}
	defer dba.Close(context.Background())
	row, err := dba.Query(context.Background(), "SELECT * FROM users")
	var subs []*db.SubUser
	for row.Next() {
		var u db.SubUser
		err = row.Scan(&u.ServiceName, &u.Price, &u.UserID, &u.StartDate, &u.EndDate)
		if err != nil {
			log.Fatalf("Ошибка получения записей: %v", err)
		}
		subs = append(subs, &u)
	}
	writeJson(w, db.SubsFull{
		Subs: subs,
	})
}

// postSubHandler создаёт новую подписку
// @Summary      Создать новую подписку
// @Description  Добавляет новую подписку в базу данных
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        subscription  body      db.SubUser  true  "Данные новой подписки"
// @Success      200  {object}  map[string]interface{}  "ID созданной подписки"
// @Failure      400  {string}  string  "Пустое тело запроса"
// @Failure      500  {string}  string  "Ошибка сервера"
// @Router       /api/subscription [post]
func postSubHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("Ошибка чтения тела запроса: %v", err)
	}
	defer r.Body.Close()
	if len(body) == 0 {
		log.Fatalf("Пустое тело запроса")
	}
	if err := godotenv.Load(); err != nil {
		log.Println("env файл не найден")
	}
	dba, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Невозможно подключиться к базе данных: %v", err)
	}
	defer dba.Close(context.Background())
	var userID string
	var u db.SubUser
	err = json.Unmarshal(body, &u)
	if err != nil {
		log.Fatalf("Ошибка десериализации: %v", err)
	}
	err = dba.QueryRow(context.Background(), `
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
	log.Println(userID)
	response := map[string]interface{}{
		"user_id": userID,
	}
	writeJson(w, response)
}

// putSubHandler обновляет существующую подписку
// @Summary      Обновить подписку
// @Description  Обновляет данные существующей подписки
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        subscription  body      db.SubUser  true  "Обновлённые данные подписки"
// @Success      200  {object}  db.SubUser  "Обновлённая подписка"
// @Failure      400  {string}  string  "Пустое тело запроса"
// @Failure      500  {string}  string  "Ошибка сервера"
// @Router       /api/subscription [put]
func putSubHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("Ошибка чтения тела запроса: %v", err)
	}
	defer r.Body.Close()
	if len(body) == 0 {
		log.Fatalf("Пустое тело запроса")
	}
	if err := godotenv.Load(); err != nil {
		log.Println("env файл не найден")
	}
	dba, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Невозможно подключиться к базе данных: %v", err)
	}
	defer dba.Close(context.Background())
	var u db.SubUser
	err = json.Unmarshal(body, &u)
	if err != nil {
		log.Fatalf("Ошибка десериализации")
	}
	_, err = dba.Exec(context.Background(), `UPDATE users
    SET service_name = $1, price = $2, start_date = $3, end_date = $4 
    WHERE user_id = $5 
    RETURNING *`, u.ServiceName, u.Price, u.StartDate, u.EndDate, u.UserID)
	if err != nil {
		log.Fatalf("Ошибка обновления записи: %v", err)
	}
	writeJson(w, u)
}

// deleteSubHandler удаляет подписку
// @Summary      Удалить подписку
// @Description  Удаляет подписку по user_id
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        user_id   query     string  true  "ID пользователя для удаления"  example(123)
// @Success      200  {object}  nil  "Подписка удалена"
// @Failure      400  {string}  string  "ID не указан"
// @Failure      500  {string}  string  "Ошибка сервера"
// @Router       /api/subscription [delete]
func deleteSubHandler(w http.ResponseWriter, r *http.Request) {
	if err := godotenv.Load(); err != nil {
		log.Println("env файл не найден")
	}
	dba, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Невозможно подключиться к базе данных: %v", err)
	}
	defer dba.Close(context.Background())
	id := r.URL.Query().Get("user_id")
	if id == "" {
		log.Fatalf("Id не указан")
	}
	_, err = dba.Exec(context.Background(), `DELETE FROM users WHERE user_id = $1`, id)
	if err != nil {
		log.Fatalf("Ошибка удаления записи: %v", err)
	}
	writeJson(w, nil)
}
func writeJson(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Ошибка сериализации^ %v", err)
	}
	_, err = w.Write(jsonData)
	if err != nil {
		log.Fatalf("Ошибка отправления: %v", err)
	}
}
