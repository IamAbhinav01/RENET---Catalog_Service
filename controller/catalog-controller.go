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
func (ctrl *CatalogController) ListMovies(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageint ,err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}
	limit := c.DefaultQuery("limit", "10")
	limitint ,err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit number"})
		return
	}

	items,total,err := ctrl.CatalogService.ListMovies(pageint,limitint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list movies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page": pageint,
		"limit": limitint,
		"data": items,
		"total": total,
	})
}