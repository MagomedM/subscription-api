package api

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"projects_for_goland/internal/models"
	"projects_for_goland/internal/service"
	"strconv"
	"strings"
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
// @Success      200  {object}  models.SubUser  "Успешный ответ"
// @Failure      400  {string}  string  "ID не указан"
// @Failure      404  {string}  string  "Подписка не найдена"
// @Failure      500  {string}  string  "Ошибка сервера"
// @Router       /api/subscription_id [get]
func getidHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	id := r.URL.Query().Get("user_id")
	if id == "" {
		log.Fatalf("Id не указан")
	}
	sub, err := service.GetSubId(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "не найдена") {
			http.Error(w, "Подписка не найдена", http.StatusNotFound)
		} else {
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			log.Printf("Ошибка получения подписки: %v", err)
		}
		return
	}
	writeJson(w, sub)
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
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	startstr := r.URL.Query().Get("start_date")
	endstr := r.URL.Query().Get("end_date")
	userID := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")

	subsum, err := service.GetSum(ctx, startstr, endstr, userID, serviceName)
	if err != nil {
		log.Fatalf("ошибка получения суммы подписок:%v", err)
	}
	writeJson(w, subsum)
	log.Println("Получение суммы всех подписок за выбранный период успешно")
}

// getSubHandler возвращает все подписки
// @Summary      Получить все подписки
// @Description  Возвращает список всех подписок
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Success      200  {object}  models.SubsFull  "Список всех подписок"
// @Failure      500  {string}  string  "Ошибка сервера"
// @Router       /api/subscription [get]
func getSubHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	subs, err := service.GetSubs(ctx, page, limit)
	if err != nil {
		log.Fatalf("Ошибка получения подписок: %v", err)
	}
	writeJson(w, subs)
	log.Printf("Получены записи (страница %d, лимит %d)", page, limit)
}

// postSubHandler создаёт новую подписку
// @Summary      Создать новую подписку
// @Description  Добавляет новую подписку в базу данных
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        subscription  body      models.SubUser  true  "Данные новой подписки"
// @Success      200  {object}  map[string]interface{}  "ID созданной подписки"
// @Failure      400  {string}  string  "Пустое тело запроса"
// @Failure      500  {string}  string  "Ошибка сервера"
// @Router       /api/subscription [post]
func postSubHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var sub models.SubUser
	body, err := io.ReadAll(r.Body)
	if err := json.Unmarshal(body, &sub); err != nil {
		log.Fatalf("Ошибка десериализации: %v", err)
	}

	defer r.Body.Close()
	userID, err := service.PostSub(ctx, &sub)
	if err != nil {
		log.Fatalf("Ошибка создания записи: %v", err)
	}
	log.Println(userID)
	response := map[string]interface{}{
		"user_id": userID,
	}
	w.WriteHeader(http.StatusCreated)
	writeJson(w, response)
	log.Printf("Создана новая запись по id: %s", userID)
}

// putSubHandler обновляет существующую подписку
// @Summary      Обновить подписку
// @Description  Обновляет данные существующей подписки
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        subscription  body      models.SubUser  true  "Обновлённые данные подписки"
// @Success      200  {object}  models.SubUser  "Обновлённая подписка"
// @Failure      400  {string}  string  "Пустое тело запроса"
// @Failure      500  {string}  string  "Ошибка сервера"
// @Router       /api/subscription [put]
func putSubHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("Ошибка чтения тела запроса: %v", err)
	}
	defer r.Body.Close()
	if len(body) == 0 {
		log.Fatalf("Пустое тело запроса")
	}

	var sub models.SubUser
	err = json.Unmarshal(body, &sub)
	if err != nil {
		log.Fatalf("Ошибка десериализации")
	}
	err = service.PutSub(ctx, &sub)
	writeJson(w, sub)
	log.Printf("Обновлена запись по id: %s", sub.UserID)
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
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id := r.URL.Query().Get("user_id")
	err := service.DelSub(ctx, id)
	if err != nil {
		log.Fatalf("Ошибка удаления записи: %v", err)
	}
	writeJson(w, nil)
	log.Printf("Удалена запись по id: %s", id)
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
