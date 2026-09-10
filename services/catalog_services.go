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
	omdbAPIKey  string
	rdb         *redis.Client
}

func NewCatalogServiceImpl(repo repositories.CatalogRepository, client *http.Client, omdbAPIKey string, rdb *redis.Client) CatalogService {
	return &CatalogServiceImpl{
		repo:        repo,
		client:      client,
		omdbAPIKey:  omdbAPIKey,
		rdb:         rdb,
	}
}

// 1. Fetch movies metadata from external API (e.g., OMDb)
func (serv *CatalogServiceImpl) FetchMoviesMetaData(title string) (*models.OMDbResponse, error) {
	apikeyURL := fmt.Sprintf("https://www.omdbapi.com/?apikey=%s", url.QueryEscape(serv.omdbAPIKey))
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

// 7. DiscoverAndIngestIndianMovies dynamically discovers and ingests new Indian cinema
// releases across all major regional industries (Bollywood, Tollywood, Kollywood,
// Mollywood, Sandalwood) from OMDb without hardcoding any movie titles.
//
// Flow per (language, year) pair:
//  1. OMDb search endpoint &s=<language>&y=<year>&type=movie  →  list of candidates
//  2. OMDb detail endpoint &i=<imdbID>                        →  Country, Language, Genre, Plot, Poster
//  3. Filter: Country must include "India" (rejects foreign keyword matches)
//  4. Dedup: skip movies already present in PostgreSQL by title
//  5. Insert into `items` with pipe-separated genres and a bootstrap interaction (rating=3.0)
//
// All OMDb communication uses serv.omdbAPIKey from the OMDB_API_KEY environment
// variable — no API key is ever hardcoded here.
func (serv *CatalogServiceImpl) DiscoverAndIngestIndianMovies(years []int) (int, error) {
	// Regional language keywords used as search terms — no hardcoded movie names.
	// OMDb's &s= search is keyword-based, so these surface language-tagged entries.
	languages := []string{
		"malayalam", // Mollywood
		"tamil",     // Kollywood
		"telugu",    // Tollywood
		"hindi",     // Bollywood
		"kannada",   // Sandalwood
	}

	ingested := 0

	for _, lang := range languages {
		for _, year := range years {
			fmt.Printf("[ingest] Searching OMDb: language=%s year=%d\n", lang, year)

			// ────── Step 1: OMDb search (returns up to 10 results per page) ──────
			searchURL := fmt.Sprintf(
				"https://www.omdbapi.com/?apikey=%s&s=%s&y=%d&type=movie&page=1",
				url.QueryEscape(serv.omdbAPIKey),
				url.QueryEscape(lang),
				year,
			)

			res, err := serv.client.Get(searchURL)
			if err != nil {
				fmt.Printf("[ingest] Warning: OMDb search failed for %s/%d: %v\n", lang, year, err)
				continue
			}

			var searchResp models.OMDbSearchResponse
			if decErr := json.NewDecoder(res.Body).Decode(&searchResp); decErr != nil {
				res.Body.Close()
				fmt.Printf("[ingest] Warning: failed to decode OMDb search response for %s/%d: %v\n", lang, year, decErr)
				continue
			}
			res.Body.Close()

			if searchResp.Response != "True" || len(searchResp.Search) == 0 {
				fmt.Printf("[ingest] No results for %s/%d\n", lang, year)
				continue
			}

			// ────── Step 2 + 3: Fetch detail per candidate, filter to Indian films ──────
			for _, candidate := range searchResp.Search {
				if candidate.ImdbID == "" {
					continue
				}

				detailURL := fmt.Sprintf("https://www.omdbapi.com/?apikey=%s&i=%s&plot=full", url.QueryEscape(serv.omdbAPIKey), candidate.ImdbID)
				detRes, detErr := serv.client.Get(detailURL)
				if detErr != nil {
					fmt.Printf("[ingest] Warning: detail fetch failed for %s: %v\n", candidate.ImdbID, detErr)
					continue
				}

				var detail models.OMDbDetailResponse
				if decErr := json.NewDecoder(detRes.Body).Decode(&detail); decErr != nil {
					detRes.Body.Close()
					continue
				}
				detRes.Body.Close()

				if detail.Response != "True" {
					continue
				}

				// Country filter — only ingest movies produced in India
				if !strings.Contains(detail.Country, "India") {
					continue
				}

				// ────── Step 4: Dedup by title ──────
				exists, chkErr := serv.repo.ItemExistsByTitle(detail.Title)
				if chkErr != nil {
					fmt.Printf("[ingest] Warning: existence check failed for %q: %v\n", detail.Title, chkErr)
					continue
				}
				if exists {
					fmt.Printf("[ingest] Skipping duplicate: %q\n", detail.Title)
					continue
				}

				// ────── Step 5: Build item and insert ──────
				// Convert "Action, Drama, Thriller" (OMDb comma-sep) to "Action|Drama|Thriller" (our format)
				rawGenres := strings.TrimSpace(detail.Genre)
				genres := strings.ReplaceAll(rawGenres, ", ", "|")
				if genres == "" || genres == "N/A" {
					genres = "Unknown"
				}

				primaryGenre := strings.SplitN(genres, "|", 2)[0]

				var posterPtr *string
				if detail.Poster != "" && detail.Poster != "N/A" {
					p := detail.Poster
					posterPtr = &p
				}

				var plotPtr *string
				if detail.Plot != "" && detail.Plot != "N/A" {
					p := detail.Plot
					plotPtr = &p
				}

				// Allocate next sequential ID
				maxID, idErr := serv.repo.GetMaxItemID()
				if idErr != nil {
					fmt.Printf("[ingest] Warning: failed to get max ID: %v\n", idErr)
					continue
				}
				newID := maxID + 1

				item := &models.Item{
					ID:           newID,
					Title:        detail.Title,
					Genres:       genres,
					PrimaryGenre: primaryGenre,
					PosterURL:    posterPtr,
					Plot:         plotPtr,
				}

				if createErr := serv.repo.CreateItem(item); createErr != nil {
					fmt.Printf("[ingest] Warning: failed to insert %q: %v\n", detail.Title, createErr)
					continue
				}

				// Bootstrap interaction so the item participates in future ALS training.
				// We use a synthetic user_id=1 with a neutral rating=3.0.
				// This ensures the item is included in FAISS content indexing immediately.
				bootstrapInteraction := &models.Interaction{
					UserID:    1,
					ItemID:    newID,
					Rating:    3.0,
					EventType: "system_ingest",
				}
				if intErr := serv.repo.CreateInteraction(bootstrapInteraction); intErr != nil {
					fmt.Printf("[ingest] Warning: bootstrap interaction failed for item %d: %v\n", newID, intErr)
					// Non-fatal — item is still ingested
				}

				fmt.Printf("[ingest] ✅ Ingested: [%s/%d] %q (id=%d, genres=%s)\n",
					lang, year, detail.Title, newID, genres)
				ingested++

				// Courteous pause to avoid rate-limiting on the free OMDb tier (1000 req/day)
				time.Sleep(250 * time.Millisecond)
			}
		}
	}

	fmt.Printf("[ingest] Complete. Total new Indian movies ingested: %d\n", ingested)
	return ingested, nil
}

// 8. Get User History
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

// 9. Record User Interaction
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

// 10. Invalidate old recommendations based on new interactions from cache
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

