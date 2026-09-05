package routers

import "github.com/gin-gonic/gin"

type Router struct {
	Router *gin.Engine
}

func NewRouter() *Router{
	r:=&Router{
		Router: gin.Default(),
	}
	r.RegisterRoutes()
	return r
}
