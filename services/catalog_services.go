package services

import (
	"fmt"
	"net/url"
	"renet-catalog/config"
	"renet-catalog/db/repositories"
	"renet-catalog/models"
)

type CatalogService interface {
	FetchMoviesMetaData(title string) (*models.OMDbResponse, error)
}

type CatalogServiceImpl struct {
	repo repositories.CatalogRepository
}

func NewCatalogServiceImpl(repo repositories.CatalogRepository) CatalogService {
	return &CatalogServiceImpl{
		repo: repo,
	}
}


// 1. Fetch movies metadata from external API (e.g., OMDb)
func (serv *CatalogServiceImpl) FetchMoviesMetaData(title string) (*models.OMDbResponse, error) {

	apikey := config.GetString("OMDB_API_KEY")
	url := fmt.Sprintf("https://www.omdbapi.com/?apikey=%s&t=%s",apikey,url.QueryEscape(title)) //url.QueryEscape("hello world") // Returns "hello+world"   
	
	return nil, nil
}