package service

import (
	"context"
	"fmt"
	"log"
	"projects_for_goland/internal/models"
	"projects_for_goland/internal/repository"
	"time"
)

func GetSubId(ctx context.Context, userID string) (*models.SubUser, error) {
	if userID == "" {
		return nil, fmt.Errorf("ID пользователя не указан")
	}
	return repository.GetUserId(ctx, userID)
}

func GetSum(ctx context.Context, startstr, endstr, userID, serviceName string) (*models.SubSum, error) {

	const layout = "01-2006"
	startdate, err := time.Parse(layout, startstr)
	if err != nil {
		log.Fatalf("Ошибка парсинга начальной даты: %v", err)
	}
	enddate, err := time.Parse(layout, endstr)
	if err != nil {
		log.Fatalf("Ошибка парсинга конечной даты: %v", err)
	}
	if enddate.Before(startdate) {
		log.Fatalf("end_date должен быть позже start_date")
	}
	enddate = time.Date(enddate.Year(), enddate.Month()+1, 0, 0, 0, 0, 0, time.UTC)

	lastDay := time.Date(enddate.Year(), enddate.Month()+1, 0, 0, 0, 0, 0, time.UTC)
	enddateWithTime := time.Date(enddate.Year(), enddate.Month(), lastDay.Day(), 23, 59, 59, 0, time.UTC)

	startDateStr := startdate.Format("2006-01-02")
	endDateStr := enddateWithTime.Format("2006-01-02")

	var sumtotal float64
	subscriptions, err := repository.GetSubByDateRange(ctx, startDateStr, endDateStr, userID, serviceName)
	if err != nil {
		log.Fatalf("Ошибка получения информации о подписках: %v", err)
	}
	for _, sub := range subscriptions {
		subStart, err := time.Parse("01-2006", sub.StartDate)
		if err != nil {
			log.Fatalf("Ошибка парсинга start_date (%s): %v", sub.StartDate, err)
			continue
		}

		subEnd, err := time.Parse("01-2006", sub.EndDate)
		if err != nil {
			log.Fatalf("Ошибка парсинга end_date (%s): %v", sub.EndDate, err)
			continue
		}

		subEnd = time.Date(subEnd.Year(), subEnd.Month()+1, 0, 23, 59, 59, 0, time.UTC)

		if subEnd.Before(startdate) || subStart.After(enddate) {
			continue
		}

		start := subStart
		if startdate.After(start) {
			start = startdate
		}

		end := subEnd
		if enddate.Before(end) {
			end = enddate
		}

		days := int(end.Sub(start).Hours()/24) + 1

		daysInMonth := time.Date(subStart.Year(), subStart.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()

		sumtotal += float64(sub.Price) * float64(days) / float64(daysInMonth)

	}
	return repository.GetUserSum(startDateStr, endDateStr, userID, serviceName, sumtotal)
}

func GetSubs(ctx context.Context, page, limit int) (*models.PaginatedResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit
	subs, total, err := repository.GetSubs(ctx, limit, offset)
	if err != nil {
		log.Fatalf("Ошибка получения подписок: %v", err)
	}
	totalPages := (total + int64(limit) - 1) / int64(limit)
	return &models.PaginatedResponse{
		Subs: subs,
		Pagination: models.Pagination{
			Page:    page,
			Limit:   limit,
			Total:   total,
			Pages:   totalPages,
			HasNext: page < int(totalPages),
			HasPrev: page > 1,
		},
	}, nil
}

func PostSub(ctx context.Context, sub *models.SubUser) (string, error) {
	if sub.ServiceName == "" || sub.Price <= 0 {
		return "", fmt.Errorf("некорректные данные")
	}
	return repository.PostSub(ctx, sub)
}

func PutSub(ctx context.Context, sub *models.SubUser) error {
	if sub.UserID == "" {
		return fmt.Errorf("ID пользователя не указан")
	}
	if sub.ServiceName == "" {
		return fmt.Errorf("название сервиса обязательно")
	}
	if sub.Price <= 0 {
		return fmt.Errorf("цена должна быть положительной")
	}
	return repository.PutSub(ctx, sub)
}

func DelSub(ctx context.Context, id string) error {
	if id == "" {
		log.Fatalf("Id не указан")
	}
	return repository.DelSub(ctx, id)
}
