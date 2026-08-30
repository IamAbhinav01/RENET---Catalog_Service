package app

import (
	"fmt"
	"net/http"
	"renet-catalog/config"
	"strings"
	"time"
)

type Application struct {
	PORT string
	DB_URL string
	OMDB_API_KEY string
	REDIS_ADDR string
}

func NewApplication() *Application {
	port := config.GetString("PORT")
	

	return &Application{
		PORT:        port,
		DB_URL:      config.GetString("DB_URL"),
		OMDB_API_KEY: config.GetString("OMDB_API_KEY"),
		REDIS_ADDR:  config.GetString("REDIS_ADDR"),
	}
}

func (app *Application) Run() error {
	addr := app.PORT
	
	if !strings.HasPrefix(addr, ":") {
		addr = ":" + addr
	}
	server := http.Server{
		Addr:         addr,
		Handler:      http.DefaultServeMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 150 * time.Second,
	}
	
	fmt.Printf("Application is successfully running on port %v\n", addr)
	
	return server.ListenAndServe()
}