# ReNet Catalog Service

A high-performance Go microservice built with the **Gin Gonic** framework and **GORM** that powers the movie catalog, handles user watch/rating interactions, executes lazy OMDb metadata hydration, and orchestrates real-time recommendation cache invalidation for the ReNet ecosystem.

---

## Architecture Overview

![ReNet Catalog Service Architecture](architecture.png)

The Catalog Service sits at the core of the ReNet platform between frontend clients, persistent storage, external metadata providers, and the Python ML Recommendation Service:

```text
                                  +-----------------------+
                                  |    Client Web App     |
                                  +-----------------------+
                                       |              ^
                 REST API (JSON over HTTP/S)          | Responses
                 X-User-ID / CORS Headers             |
                                       v              |
     =============================[ Go Catalog Microservice ]=============================
     |                                                                                   |
     |  +-----------------------------------------------------------------------------+  |
     |  |                            Gin Router Layer                                 |  |
     |  |   - CORS Middleware (OPTIONS 204)    - Dynamic Auth Context (X-User-ID)     |  |
     |  |   - /health & /api/health            - Standardized /api/movies & /api/history |
     |  +-----------------------------------------------------------------------------+  |
     |                                         |                                         |
     |  +-----------------------------------------------------------------------------+  |
     |  |                           Catalog Controller                                |  |
     |  |   - ListMovies     - GetMovieByID     - BatchGetMovies     - SearchMovies   |  |
     |  |   - RecordInteraction                 - GetUserHistory                      |  |
     |  +-----------------------------------------------------------------------------+  |
     |                                         |                                         |
     |  +-----------------------------------------------------------------------------+  |
     |  |                             Service Layer                                   |  |
     |  |   - Catalog & Search Logic            - Batch Hydration Engine              |  |
     |  |   - Interaction Logger                - Cache Invalidation Orchestrator     |  |
     |  |   - Title Normalizer (Regex)          - Async OMDb Enrichment Worker        |  |
     |  +-----------------------------------------------------------------------------+  |
     |          |                               |                             |          |
     ===========|===============================|=============================|===========
                | (Read/Write)                  | (Invalidate Cache)          | (Enrich)
                v                               v                             v
     +--------------------+           +-------------------+         +-------------------+
     | PostgreSQL (renet) |           |    Redis Cache    |         |  OMDb External    |
     |--------------------|           |-------------------|         |     REST API      |
     | - items            |           | recs:user:<id>:*  |         | - Posters & Plots |
     | - interactions     |           +-------------------+         +-------------------+
     +--------------------+                     ^
                                                | (Pulls fresh recs)
                                      +--------------------+
                                      | ReNet Python ML    |
                                      | Recommendation     |
                                      | Service (FastAPI)  |
                                      +--------------------+
```

---

## System Role & Responsibilities

### 1. Movie Catalog & Batch Hydration

- **Paginated Listings & Case-Insensitive Search**: Fast indexed queries over 9,700+ movies via PostgreSQL and GORM.
- **Batch Movie Lookup (`POST /api/movies/batch`)**: Hydrates multiple recommendation candidate IDs in a single roundtrip, eliminating $N+1$ HTTP queries from clients.

### 2. Lazy Metadata Enrichment with Article Normalization

- **Lazy Hydration**: When a movie is first accessed without poster or plot information, the service triggers an asynchronous background goroutine to query the OMDb API and persist the metadata to PostgreSQL. Subsequent reads are served directly from PostgreSQL at sub-millisecond speeds.
- **MovieLens Title Normalization**: Converts inverted MovieLens titles (e.g., `"Shawshank Redemption, The"` $\to$ `"The Shawshank Redemption"`, `"Lion King, The"` $\to$ `"The Lion King"`, `"Postman, The (Postino, Il)"` $\to$ `"The Postman"`) to ensure accurate OMDb match rates.
- **Infinite Loop Protection**: Identifies negative responses (`Response: "False"`) or missing posters and persists a sentinel (`"N/A"`) to prevent re-query loops on every read.
- **Dedicated HTTP Timeout**: Utilizes a pooled HTTP client with an explicit 5-second timeout to protect upstream thread availability.

### 3. Interaction Ingestion & Cache Invalidation Orchestrator

- **Interaction Storage**: Appends ratings and watch history to the PostgreSQL `interactions` table.
- **Real-Time Cache Synchronization**: Automatically detects and deletes the user's cached recommendation keys in Redis (`recs:user:<user_id>:*`).
- **Resilient Empty-Key Guard**: Safely inspects key length before issuing Redis `DEL` commands, eliminating empty argument errors.
- **Connection Pooling**: Utilizes a persistent singleton Redis client with connection pooling (`PoolSize: 10`, `MinIdleConns: 2`) across the application lifecycle.

### 4. Cross-Cutting Utilities

- **CORS Middleware**: Pre-configured for web/mobile frontends with automatic preflight `OPTIONS` handling (HTTP 204 No Content).
- **Flexible Auth Context**: Non-blocking `X-User-ID` header extraction with fallback to user `1` for test setups.
- **Health Probes**: `/health` and `/api/health` endpoints returning service status and UTC timestamp.

---

## Project Structure

```text
Renet_CataLog_Service/
├── api/                               # PowerShell end-to-end test suites
│   ├── health/
│   │   └── health_check.ps1           # Verifies /health and /api/health probes
│   ├── history/
│   │   ├── getHistory_history.ps1     # Tests user history retrieval
│   │   └── record-interaction_history.ps1 # Tests interaction logging & cache invalidation
│   └── movies/
│       ├── batch_movies.ps1           # Tests batch movie hydration endpoint
│       ├── get_movie.ps1              # Tests single movie retrieval & enrichment
│       ├── list_movies.ps1            # Tests paginated movie listing
│       └── search_movie.ps1           # Tests case-insensitive movie search
├── app/
│   └── application.go                 # Application bootstrap, client pooling, and HTTP server
├── config/
│   ├── db_config.go                   # PostgreSQL GORM connection management
│   ├── env_config.go                  # Environment variable configuration parser (.env)
│   └── redis_config.go                # Redis connection pooling and health ping
├── controller/
│   └── catalog-controller.go          # HTTP request handlers & validation logic
├── db/
│   └── repositories/
│       └── catalog_repositroies.go    # Database abstraction layer (PostgreSQL / GORM)
├── middleware/
│   ├── cors.go                        # Cross-Origin Resource Sharing handler
│   └── test_auth.go                   # Dynamic user context extraction (X-User-ID)
├── models/
│   └── schemas.go                     # Data transfer objects and database models
├── routers/
│   ├── history_router.go              # History and interaction route declarations
│   ├── movie_router.go                # Movie catalog, search, and batch routes
│   └── router.go                      # Engine setup, middleware mounting, and route wiring
├── services/
│   ├── catalog_services.go            # Business logic, OMDb client, and cache invalidator
│   └── catalog_services_test.go       # Unit tests for title normalization
├── .air.toml                          # Live reload configuration for local development
├── .env                               # Environment variables configuration file
├── architecture.png                   # System architecture diagram
├── go.mod                             # Go module definitions
├── go.sum                             # Go checksums
├── main.go                            # Service main entrypoint
└── README.md                          # Service documentation
```

---

## Environment Variables

Configure the following variables in your `.env` file or export them into the environment:

| Variable       | Type     | Default          | Description                                                                              |
| :------------- | :------- | :--------------- | :--------------------------------------------------------------------------------------- |
| `PORT`         | `string` | `3000`           | Port on which the HTTP server listens (e.g. `3000` or `:3000`)                           |
| `DB_URL`       | `string` | _Required_       | PostgreSQL connection string (`postgresql://user:pass@host:5432/dbname?sslmode=disable`) |
| `REDIS_ADDR`   | `string` | `localhost:6379` | Host and port of the Redis cache instance                                                |
| `OMDB_API_KEY` | `string` | _Required_       | OMDb API key (e.g. `YOUR_KEY`)                                                           |

---

## API Reference

All primary endpoints are prefixed with `/api`. Root paths (e.g., `/movies`, `/history`) are also supported for backwards compatibility.

### Health Probes

#### `GET /api/health` or `GET /health`

Returns service health status and current UTC server time.

**Response (`200 OK`)**:

```json
{
  "service": "renet-catalog",
  "status": "ok",
  "time": "2026-09-05T20:47:33Z"
}
```

---

### Movie Catalog

#### `GET /api/movies`

Returns a paginated list of movies. Supports both `/api/movies` and `/api/movies/`.

**Query Parameters**:

- `page` _(optional, int)_: Page number (default: `1`).
- `limit` _(optional, int)_: Items per page (default: `10`, max: `100`).

**Response (`200 OK`)**:

```json
{
  "data": [
    {
      "id": 1,
      "title": "Toy Story (1995)",
      "genres": "Adventure|Animation|Children|Comedy|Fantasy",
      "primary_genre": "Adventure",
      "poster_url": "https://m.media-amazon.com/images/M/...jpg",
      "plot": "A cowboy doll is profoundly jealous when a new spaceman action figure..."
    }
  ],
  "limit": 10,
  "page": 1,
  "total": 9742
}
```

---

#### `GET /api/movies/:id`

Retrieves details for a single movie. If `poster_url` is unpopulated, asynchronously triggers OMDb metadata enrichment.

**Parameters**:

- `id` _(path, int)_: Movie ID.

**Response (`200 OK`)**:

```json
{
  "id": 364,
  "title": "Lion King, The (1994)",
  "genres": "Adventure|Animation|Children|Drama|Musical|IMAX",
  "primary_genre": "Adventure",
  "poster_url": "https://m.media-amazon.com/images/M/...jpg",
  "plot": "Lion prince Simba and his father are targeted by his bitter uncle..."
}
```

---

#### `POST /api/movies/batch`

Batch-hydrates multiple movies in a single query. Ideal for recommendation list enrichment.

**Request Body**:

```json
{
  "ids": [1, 2, 364]
}
```

**Response (`200 OK`)**:

```json
{
  "count": 3,
  "data": [
    {
      "id": 1,
      "title": "Toy Story (1995)",
      "genres": "Adventure|Animation|Children|Comedy|Fantasy",
      "primary_genre": "Adventure",
      "poster_url": "https://m.media-amazon.com/images/M/...jpg",
      "plot": "A cowboy doll is profoundly jealous..."
    },
    {
      "id": 2,
      "title": "Jumanji (1995)",
      "genres": "Adventure|Children|Fantasy",
      "primary_genre": "Adventure",
      "poster_url": null,
      "plot": null
    },
    {
      "id": 364,
      "title": "Lion King, The (1994)",
      "genres": "Adventure|Animation|Children|Drama|Musical|IMAX",
      "primary_genre": "Adventure",
      "poster_url": "https://m.media-amazon.com/images/M/...jpg",
      "plot": "Lion prince Simba and his father are targeted by his bitter uncle..."
    }
  ]
}
```

---

#### `GET /api/movies/search`

Searches movies by title (case-insensitive substring match).

**Query Parameters**:

- `q` _(required, string)_: Search keyword.
- `limit` _(optional, int)_: Maximum results (default: `20`, max: `50`).

**Example Request**:

```http
GET /api/movies/search?q=toy&limit=5
```

---

### User History & Interactions

#### `POST /api/history`

Records a user rating or watch interaction, and invalidates the user's recommendation cache in Redis (`recs:user:<user_id>:*`).

**Headers**:

- `X-User-ID` _(optional, int)_: ID of the interacting user (defaults to `1` in test environments).

**Request Body**:

```json
{
  "item_id": 364,
  "rating": 5.0,
  "event_type": "rating"
}
```

**Response (`201 Created`)**:

```json
{
  "message": "Interaction recorded successfully"
}
```

---

#### `GET /api/history`

Retrieves interaction history for the authenticated user, ordered from most recent to oldest.

**Headers**:

- `X-User-ID` _(optional, int)_: User ID (defaults to `1`).

**Query Parameters**:

- `limit` _(optional, int)_: Maximum interactions to retrieve (default: `50`, max: `100`).

**Response (`200 OK`)**:

```json
{
  "limit": 50,
  "user_id": 1,
  "data": [
    {
      "id": 100845,
      "user_id": 1,
      "item_id": 364,
      "rating": 5.0,
      "event_type": "rating"
    }
  ]
}
```

---

## Local Development & Setup

### Prerequisites

- **Go 1.21+** installed
- **PostgreSQL 14+** running with the MovieLens database seeded (`renet`)
- **Redis 6+** running on `localhost:6379`
- Active **OMDb API Key**

### 1. Installation

Clone the repository and install dependencies:

```bash
cd Renet_CataLog_Service
go mod download
```

### 2. Configure Environment

Ensure `.env` contains your database and cache credentials:

```env
PORT=3000
DB_URL=postgresql://user:pass@localhost:5432/renet?sslmode=disable
REDIS_ADDR=localhost:6379
OMDB_API_KEY=YOUR_OMDB_KEY
```

### 3. Run Service

Run directly using the Go toolchain:

```bash
go run main.go
```

Or build the optimized binary:

```bash
go build -o renet-catalog.exe .
./renet-catalog.exe
```

For live reloading during development:

```bash
air
```

---

## Testing & Verification

### Running Unit Tests

Execute unit tests for title parsing and article inversion normalization:

```bash
go test -v ./services
```

### Running PowerShell Integration Suites

Interactive test scripts are located under the `api/` directory:

```powershell
# Health check probe
powershell -File api/health/health_check.ps1

# List movies test
powershell -File api/movies/list_movies.ps1

# Batch movie hydration test
powershell -File api/movies/batch_movies.ps1

# Record interaction & cache invalidation test
powershell -File api/history/record-interaction_history.ps1

# Retrieve user interaction history test
powershell -File api/history/getHistory_history.ps1
```
