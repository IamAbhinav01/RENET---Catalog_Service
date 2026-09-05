package routers

import (
	"renet-catalog/controller"
	"renet-catalog/middleware"
	"renet-catalog/services"

	"github.com/gin-gonic/gin"
)

func (r *Router) RegisterWithHistory(catalogService services.CatalogService) {
	catalogController := &controller.CatalogController{
		CatalogService: catalogService,
	}

	registerHistoryRoutes := func(group *gin.RouterGroup) {
		group.Use(middleware.TestAuth())
		group.POST("", catalogController.RecordInteraction)
		group.POST("/", catalogController.RecordInteraction)
		group.GET("", catalogController.GetUserHistory)
		group.GET("/", catalogController.GetUserHistory)
	}

	// Standard API route group: /api/history
	apiHistory := r.Router.Group("/api/history")
	registerHistoryRoutes(apiHistory)

	// Backward-compatible root route group: /history
	rootHistory := r.Router.Group("/history")
	registerHistoryRoutes(rootHistory)
}
