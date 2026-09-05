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

	db,err := config.ConnectDB(app.DB_URL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	repo := repositories.NewCatalogRepository(db)
	catalogService := services.NewCatalogServiceImpl(repo,http.DefaultClient,app.OMDB_API_KEY,app.REDIS_ADDR)

	router := routers.NewRouter(catalogService)

	addr := app.PORT
	
	if !strings.HasPrefix(addr, ":") {
		addr = ":" + addr
	}
	server := http.Server{
		Addr:         addr,
		Handler:      router.Router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 150 * time.Second,
	}
	
	fmt.Printf("Application is successfully running on port %v\n", addr)
	
	return server.ListenAndServe()
}
