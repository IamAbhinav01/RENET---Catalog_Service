package routers

import (
	"renet-catalog/controller"
	"renet-catalog/services"
)

func (r *Router) RegiserWithMovies(catalogService services.CatalogService) {
    movies := r.Router.Group("/movies")

    catalogController := &controller.CatalogController{
        CatalogService: catalogService,
    }

    movies.GET("/:id", catalogController.GetMovie)
    movies.GET("/", catalogController.ListMovies)
}