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
	GetMovieByID(id int) (*models.Item, error)
	ListMovies(page, limit int) ([]models.Item, int64, error)
	SearchMovies(query string, limit int) ([]models.Item, error)
	EmbedMovieMetadata(itemID int, rawTitle string)
	GetUserHistory(userId int, limit int)([]models.Interaction, error)
	RecordUserInteraction(userId int,req *models.CreateInteractionRequest) error
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

	apikey_url := config.GetString("OMDB_API_KEY")
	title,year := serv.ConcatenateTitleAndYear(title)
	url := fmt.Sprintf("%s&t=%s",apikey_url,url.QueryEscape(title)) //url.QueryEscape("hello world") // Returns "hello+world"   
	if year != "" {
		url += fmt.Sprintf("&y=%s", year)
	}
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
//3. Get movie by ID
func (serv *CatalogServiceImpl) GetMovieByID(id int) (*models.Item, error) {
	item,err:=serv.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if item == nil || *item.PosterURL == "" || *item.PosterURL == "N/A" {
		go serv.EmbedMovieMetadata(id, item.Title)
	}
	return item, nil
}
//4. Get all movies
func (serv *CatalogServiceImpl) ListMovies(page, limit int) ([]models.Item, int64, error) {
	if page < 1{
		page = 1
	}
	if limit < 1 || limit > 100{
		limit = 20
	}
	items, count, err := serv.repo.ListItems(page, limit)
	if err != nil {
		return nil, 0, err
	}
	return items, count, nil
}
//5. Search movies 
func(serv *CatalogServiceImpl)SearchMovies(query string, limit int) ([]models.Item, error){
	if limit < 1 || limit > 50{
		limit = 20
	}
	items,err := serv.repo.SearchItems(query,limit)
	if err != nil {
		return nil, err
	}
	return items, nil
}
//6. Embed movie metadata
func (serv *CatalogServiceImpl) EmbedMovieMetadata(itemID int, title string){
	omdbData,err := serv.FetchMoviesMetaData(title)
	if err != nil {
		fmt.Printf("Error fetching metadata for item ID %d: %v\n", itemID, err)
		return
	}
	posterURL := omdbData.Poster
	if posterURL == "N/A" {
		posterURL = ""
	}
	plot := omdbData.Plot
	err = serv.repo.UpdatePosterAndPlot(itemID,posterURL,plot)
	if err != nil {
		fmt.Printf("Error updating metadata for item ID %d: %v\n", itemID, err)
	}else{
		fmt.Printf("Successfully updated metadata for item ID %d\n", itemID)
	}
}
//7. Get User History
func(serv *CatalogServiceImpl)GetUserHistory(userId int, limit int)([]models.Interaction, error){
	if limit <= 0 || limit > 100{
		limit = 50
	}
	interaction,err:=serv.repo.GetByUserIDInteraction(userId,limit)
	if err != nil {
		return nil, err
	}
	return interaction, nil
}
//8. Record User Interaction
func(serv *CatalogServiceImpl)RecordUserInteraction(userId int,req *models.CreateInteractionRequest) error{
	eventType :=req.EventType
	if eventType == ""{
		eventType = "rating"
	}

	interaction := &models.Interaction{
		UserID: userId,
		ItemID: req.ItemID,
		Rating: req.Rating,
		EventType: eventType,
	}

	err:=serv.repo.CreateInteraction(interaction)
	if err != nil {
		fmt.Printf("Error recording interaction for user ID %d: %v\n", userId, err)
		return err
	}
	return nil
}