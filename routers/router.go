package routers

import (
	"renet-catalog/controller"
	"renet-catalog/middleware"
	"renet-catalog/services"

	"github.com/gin-gonic/gin"
)

type Router struct {
	Router *gin.Engine
}

func NewRouter(catalogService services.CatalogService) *Router {
	engine := gin.Default()
	engine.Use(middleware.CORS())

	r := &Router{
		Router: engine,
	}

	r.RegisterRoutes(catalogService)
	return r
}

func (r *Router) RegisterRoutes(catalogService services.CatalogService) {
	ctrl := &controller.CatalogController{
		CatalogService: catalogService,
	}

	// Health check endpoints
	r.Router.GET("/health", ctrl.HealthCheck)
	r.Router.GET("/api/health", ctrl.HealthCheck)

	// Resource routes
	r.RegisterWithMovies(catalogService)
	r.RegisterWithHistory(catalogService)
}
