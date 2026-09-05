package routers

import (
	"renet-catalog/controller"
	"renet-catalog/services"

	"github.com/gin-gonic/gin"
)

func (r *Router) RegisterWithMovies(catalogService services.CatalogService) {
	catalogController := &controller.CatalogController{
		CatalogService: catalogService,
	}

	registerMovieRoutes := func(group *gin.RouterGroup) {
		group.GET("", catalogController.ListMovies)
		group.GET("/", catalogController.ListMovies)
		group.GET("/search", catalogController.SearchMovies)
		group.POST("/batch", catalogController.BatchGetMovies)
		group.GET("/:id", catalogController.GetMovie)
	}

	// Standard API route group: /api/movies
	apiMovies := r.Router.Group("/api/movies")
	registerMovieRoutes(apiMovies)

	// Backward-compatible root route group: /movies
	rootMovies := r.Router.Group("/movies")
	registerMovieRoutes(rootMovies)
}
