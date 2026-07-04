package server

import "net/http"

func Run() error {
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
	return nil
}
