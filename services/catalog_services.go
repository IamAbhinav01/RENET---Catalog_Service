package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"renet-catalog/db/repositories"
	"renet-catalog/models"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type CatalogService interface {
	ConcatenateTitleAndYear(text string) (title string, year string)
	FetchMoviesMetaData(title string) (*models.OMDbResponse, error)
	GetMovieByID(id int) (*models.Item, error)
	GetMoviesByIDs(ids []int) ([]models.Item, error)
	ListMovies(page, limit int) ([]models.Item, int64, error)
	SearchMovies(query string, limit int) ([]models.Item, error)
	EmbedMovieMetadata(itemID int, rawTitle string)
	DiscoverAndIngestIndianMovies(years []int) (int, error)
	GetUserHistory(userId int, limit int) ([]models.Interaction, error)
	RecordUserInteraction(userId int, req *models.CreateInteractionRequest) error
	InvalidateRecommendations(userId int) error
}

type CatalogServiceImpl struct {
	repo        repositories.CatalogRepository
	client      *http.Client
	omdbAPI_URL string
	rdb         *redis.Client
}

func NewCatalogServiceImpl(repo repositories.CatalogRepository, client *http.Client, omdbAPI_URL string, rdb *redis.Client) CatalogService {
	return &CatalogServiceImpl{
		repo:        repo,
		client:      client,
		omdbAPI_URL: omdbAPI_URL,
		rdb:         rdb,
	}
}

// 1. Fetch movies metadata from external API (e.g., OMDb)
func (serv *CatalogServiceImpl) FetchMoviesMetaData(title string) (*models.OMDbResponse, error) {
	apikeyURL := serv.omdbAPI_URL
	cleanTitle, year := serv.ConcatenateTitleAndYear(title)
	reqURL := fmt.Sprintf("%s&t=%s", apikeyURL, url.QueryEscape(cleanTitle))
	if year != "" {
		reqURL += fmt.Sprintf("&y=%s", year)
	}

	res, err := serv.client.Get(reqURL)
	if err != nil {
		fmt.Printf("Error occurred while retrieving info from OMDB server: %v\n", err)
		return nil, err
	}
	defer res.Body.Close()

	var omdbResponse models.OMDbResponse
	if err := json.NewDecoder(res.Body).Decode(&omdbResponse); err != nil {
		fmt.Printf("Error occurred while decoding response from OMDB server: %v\n", err)
		return nil, err
	}

	if omdbResponse.Response == "False" {
		return nil, fmt.Errorf("omdb: %s", omdbResponse.Error)
	}

	return &omdbResponse, nil
}

// 2. Concatenate title and year, normalizing MovieLens inverted articles
func (serv *CatalogServiceImpl) ConcatenateTitleAndYear(text string) (title string, year string) {
	text = strings.TrimSpace(text)
	reYear := regexp.MustCompile(`^(.*?)\s*\((\d{4})\)$`)
	matches := reYear.FindStringSubmatch(text)
	if len(matches) == 3 {
		title = strings.TrimSpace(matches[1])
		year = matches[2]
	} else {
		title = text
		year = ""
	}

	// Remove secondary parenthetical info if any (e.g., "Postman, The (Postino, Il)" -> "Postman, The")
	reAlt := regexp.MustCompile(`\s*\([^)]*\)$`)
	cleaned := strings.TrimSpace(reAlt.ReplaceAllString(title, ""))
	if cleaned != "" {
		title = cleaned
	}

	// Normalize article inversion: ", The", ", A", ", An"
	articles := []string{", The", ", the", ", A", ", a", ", An", ", an"}
	for _, art := range articles {
		if strings.HasSuffix(title, art) {
			prefix := strings.TrimSpace(art[2:])
			base := strings.TrimSpace(title[:len(title)-len(art)])
			title = fmt.Sprintf("%s %s", prefix, base)
			break
		}
	}

	return title, year
}

// 3. Get movie by ID
func (serv *CatalogServiceImpl) GetMovieByID(id int) (*models.Item, error) {
	item, err := serv.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, fmt.Errorf("movie %d not found", id)
	}
	// Trigger metadata enrichment if not yet enriched (PosterURL is nil)
	if item.PosterURL == nil {
		go serv.EmbedMovieMetadata(id, item.Title)
	}
	return item, nil
}

// 3b. Get movies by slice of IDs
func (serv *CatalogServiceImpl) GetMoviesByIDs(ids []int) ([]models.Item, error) {
	if len(ids) == 0 {
		return []models.Item{}, nil
	}
	return serv.repo.GetByIDs(ids)
}

// 4. Get all movies
func (serv *CatalogServiceImpl) ListMovies(page, limit int) ([]models.Item, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	items, count, err := serv.repo.ListItems(page, limit)
	if err != nil {
		return nil, 0, err
	}
	return items, count, nil
}

// 5. Search movies
func (serv *CatalogServiceImpl) SearchMovies(query string, limit int) ([]models.Item, error) {
	if limit < 1 || limit > 50 {
		limit = 20
	}
	items, err := serv.repo.SearchItems(query, limit)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// 6. Embed movie metadata
func (serv *CatalogServiceImpl) EmbedMovieMetadata(itemID int, title string) {
	omdbData, err := serv.FetchMoviesMetaData(title)
	if err != nil {
		fmt.Printf("Notice: OMDb enrichment unavailable for item ID %d (%s): %v\n", itemID, title, err)
		na := "N/A"
		_ = serv.repo.UpdatePosterAndPlot(itemID, na, "")
		return
	}

	posterURL := omdbData.Poster
	if posterURL == "" {
		posterURL = "N/A"
	}
	plot := omdbData.Plot

	err = serv.repo.UpdatePosterAndPlot(itemID, posterURL, plot)
	if err != nil {
		fmt.Printf("Error updating metadata for item ID %d: %v\n", itemID, err)
	} else {
		fmt.Printf("Successfully updated metadata for item ID %d\n", itemID)
	}
}

// 7. Get User History
func (serv *CatalogServiceImpl) GetUserHistory(userId int, limit int) ([]models.Interaction, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	interaction, err := serv.repo.GetByUserIDInteraction(userId, limit)
	if err != nil {
		return nil, err
	}
	return interaction, nil
}

// 8. Record User Interaction
func (serv *CatalogServiceImpl) RecordUserInteraction(userId int, req *models.CreateInteractionRequest) error {
	eventType := req.EventType
	if eventType == "" {
		eventType = "rating"
	}

	interaction := &models.Interaction{
		UserID:    userId,
		ItemID:    req.ItemID,
		Rating:    req.Rating,
		EventType: eventType,
	}

	err := serv.repo.CreateInteraction(interaction)
	if err != nil {
		fmt.Printf("Error recording interaction for user ID %d: %v\n", userId, err)
		return err
	}
	if err := serv.InvalidateRecommendations(userId); err != nil {
		fmt.Printf("Warning: failed to invalidate recommendations for user ID %d: %v\n", userId, err)
	}
	return nil
}

// 9. Invalidate old recommendations based on new interactions from cache
func (serv *CatalogServiceImpl) InvalidateRecommendations(userId int) error {
	if serv.rdb == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pattern := fmt.Sprintf("recs:user:%d:*", userId)
	keys, err := serv.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		fmt.Printf("Error fetching keys for user ID %d: %v\n", userId, err)
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	err = serv.rdb.Del(ctx, keys...).Err()
	if err != nil {
		fmt.Printf("Error deleting keys for user ID %d: %v\n", userId, err)
		return err
	}

	fmt.Printf("Successfully invalidated %d recommendation cache keys for user ID %d\n", len(keys), userId)
	return nil
}
