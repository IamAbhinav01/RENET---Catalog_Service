package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"renet-catalog/config"
	"renet-catalog/db/repositories"
	"renet-catalog/models"
	"strings"
)

type CatalogService interface {
	ConcatenateTitleAndYear(text string) (title string, year string) 
	FetchMoviesMetaData(title string) (*models.OMDbResponse, error)
}

type CatalogServiceImpl struct {
	repo repositories.CatalogRepository
	client *http.Client
}

func NewCatalogServiceImpl(repo repositories.CatalogRepository, client *http.Client) CatalogService {
	return &CatalogServiceImpl{
		repo: repo,
		client: client,
	}
}


// 1. Fetch movies metadata from external API (e.g., OMDb)
func (serv *CatalogServiceImpl) FetchMoviesMetaData(title string) (*models.OMDbResponse, error) {

	apikey := config.GetString("OMDB_API_KEY")
	url := fmt.Sprintf("https://www.omdbapi.com/?apikey=%s&t=%s",apikey,url.QueryEscape(title)) //url.QueryEscape("hello world") // Returns "hello+world"   
	res,err:=serv.client.Get(url)
	if err!=nil{
		fmt.Printf("Error occured while retreiving the info from OMDB server %s",err)
		return nil,err
	}
	defer res.Body.Close()

	var omdbResponse models.OMDbResponse
	if err := json.NewDecoder(res.Body).Decode(&omdbResponse); err != nil {
		fmt.Printf("Error occured while decoding the response from OMDB server %s", err)
		return nil, err
	}

	return &omdbResponse, nil
}
// 2. Concatenate title and year
func (serv *CatalogServiceImpl) ConcatenateTitleAndYear(text string) (title string, year string)  {
	re := regexp.MustCompile(`^(.*)\s*\((\d{4})\)$`)
	matches := re.FindStringSubmatch(strings.TrimSpace(text))
	if len(matches) == 3 {
		return strings.TrimSpace(matches[1]), matches[2]
	}
	return strings.TrimSpace(text), ""
}