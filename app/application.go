package app

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Application struct {
	PORT string
}

func NewApplication() *Application {
	return &Application{
		PORT: "3000",
	}
}

func (app *Application) Run() error {
	addr := app.PORT
	if !strings.HasPrefix(addr,":"){
		addr = ":"+addr
	}
	server := http.Server{
		Addr: addr,
		Handler: http.DefaultServeMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 150 * time.Second,
	}
	fmt.Printf("Application is successfully running on port %v",addr)
	return  server.ListenAndServe()
}