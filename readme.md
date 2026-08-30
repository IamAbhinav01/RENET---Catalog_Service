## catalog service

#### 1. **Movie Catalog & Hydration**:

- Delivers paginated movie listings and title search results from PostgreSQL.
- Performs **Lazy Metadata Enrichment** via the OMDb API: When a movie is first requested without poster or plot information, the service queries OMDb, updates PostgreSQL, and returns the enriched data. Subsequent queries serve directly from PostgreSQL at sub-millisecond speeds.

#### 2. **Watch & Interaction Logging**:

- Records user watch events and ratings in the `interactions` table.

#### 3. **Recommendation Cache Invalidation**:

- Automatically invalidates the user's cached recommendation keys in Redis (`recs:user:<user_id>:*`) upon recording any new interaction, ensuring the Python ML Recommendation Service immediately generates fresh suggestions on the user's next visit.
