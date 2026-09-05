package routers

import (
	"renet-catalog/services"

	"github.com/gin-gonic/gin"
)

type Router struct {
    Router *gin.Engine
}

func NewRouter(catalogService services.CatalogService) *Router {
    r := &Router{
        Router: gin.Default(),
    }

    r.RegisterRoutes(catalogService)
    return r
}

func (r *Router) RegisterRoutes(catalogService services.CatalogService) {
    r.RegiserWithMovies(catalogService)
    r.RegiserWithHistory(catalogService)
}