package app

import (
	"fmt"
	"net/http"
	"renet-catalog/config"
	"renet-catalog/db/repositories"
	"renet-catalog/routers"
	"renet-catalog/services"
	"strings"
	"time"
)

type Application struct {
	PORT         string
	DB_URL       string
	OMDB_API_KEY string
	REDIS_ADDR   string
}

func NewApplication() *Application {
	return &Application{
		PORT:         config.GetString("PORT"),
		DB_URL:       config.GetString("DB_URL"),
		OMDB_API_KEY: config.GetString("OMDB_API_KEY"),
		REDIS_ADDR:   config.GetString("REDIS_ADDR"),
	}
}

func (app *Application) Run() error {
	db, err := config.ConnectDB(app.DB_URL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	rdb := config.ConnectRedis(app.REDIS_ADDR)
	if rdb != nil {
		defer rdb.Close()
	}

	omdbHTTPClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	repo := repositories.NewCatalogRepository(db)
	catalogService := services.NewCatalogServiceImpl(repo, omdbHTTPClient, app.OMDB_API_KEY, rdb)

	router := routers.NewRouter(catalogService)

	addr := app.PORT
	if addr == "" {
		addr = "3000"
	}
	if !strings.HasPrefix(addr, ":") {
		addr = ":" + addr
	}

	server := http.Server{
		Addr:         addr,
		Handler:      router.Router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	fmt.Printf("Application is successfully running on port %v\n", addr)
	return server.ListenAndServe()
}
