package main

import (
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"projects_for_goland/api"
	"projects_for_goland/db"
	_ "projects_for_goland/docs"
	"projects_for_goland/server"
)

func main() {
	log.Println("Запуск сервера")
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)
	db.Init()
	log.Println("Иницилизация базы данных прошла успешно")
	api.Init()
	log.Println("Загрузка API прошла успешно")
	server.Run()
	log.Println("Отключение сервера")
}
