package controller

import (
	"net/http"
	"renet-catalog/models"
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
func(ctrl *CatalogController) SearchMovies(c *gin.Context) {
	query := c.Query("q")

	if query == ""{
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}
	limit := c.DefaultQuery("limit", "20")
	limitint ,err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit number"})
		return
	}

	items, err := ctrl.CatalogService.SearchMovies(query, limitint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search movies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query": query,
		"limit": limitint,
		"data": items,
	})
}
func(ctrl *CatalogController) RecordInteraction(c *gin.Context) {
	userID,exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDInt,ok := userID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}
	var req models.CreateInteractionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	
	err := ctrl.CatalogService.RecordUserInteraction(userIDInt,&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record interaction"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Interaction recorded successfully"})
}
func(ctrl *CatalogController) GetUserHistory(c *gin.Context) {
	userID,exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDInt,ok := userID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}
	limit := c.DefaultQuery("limit", "50")
	limitint ,err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit number"})
		return
	}
	interactions, err := ctrl.CatalogService.GetUserHistory(userIDInt, limitint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user history"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user_id": userIDInt,
		"limit": limitint,
		"data": interactions,
	})
}