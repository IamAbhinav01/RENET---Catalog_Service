package controller

import (
	"net/http"
	"renet-catalog/services"
	"strconv"
)

type CatalogController struct {
	CatalogService services.CatalogService
}

func (ctrl *CatalogController) GetMovie(w http.ResponseWriter, r *http.Request) {
	
	id := r.URL.Query().Get("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}
	if id == "" {
		http.Error(w, "Missing movie ID", http.StatusBadRequest)
		return
	}
	movie,err := ctrl.CatalogService.GetMovieByID(idInt)
	if err != nil {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}
	
}