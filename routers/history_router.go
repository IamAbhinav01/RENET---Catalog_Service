package routers

import (
	"renet-catalog/controller"
	"renet-catalog/middleware"
	"renet-catalog/services"
)

func (r *Router) RegiserWithHistory(catalogService services.CatalogService) {
	history := r.Router.Group("/history")
	history.Use(middleware.TestAuth())

	catalogController := &controller.CatalogController{
		CatalogService: catalogService,
	}

	history.POST("", catalogController.RecordInteraction)
	history.GET("", catalogController.GetUserHistory)
}
