package controller

import (
	"net/http"
	"renet-catalog/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CatalogController struct {
	CatalogService services.CatalogService
}

func (ctrl *CatalogController) GetMovie(c *gin.Context) {
	
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})
		return
	}
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing movie ID"})
		return
	}
	movie,err := ctrl.CatalogService.GetMovieByID(idInt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}
	c.JSON(http.StatusOK, movie)
}