package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
)

const (
	DefaultGammaBaseURL = "https://gamma-api.polymarket.com"
	defaultTimeout      = 30 * time.Second
)

// GammaClient interacts with the Polymarket Gamma API.
type GammaClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewGammaClient creates a new Gamma API client.
func NewGammaClient(baseURL string) *GammaClient {
	if baseURL == "" {
		baseURL = DefaultGammaBaseURL
	}
	return &GammaClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// get performs a GET request and returns the response body.
func (c *GammaClient) get(path string) ([]byte, error) {
	url := c.baseURL + path
	log.Debug("GET", "url", url)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetMarket retrieves a single market by ID.
func (c *GammaClient) GetMarket(id string) (*Market, error) {
	data, err := c.get("/markets/" + id)
	if err != nil {
		return nil, err
	}

	var market Market
	if err := json.Unmarshal(data, &market); err != nil {
		return nil, fmt.Errorf("parsing market: %w", err)
	}
	return &market, nil
}

// GetMarketBySlug retrieves a single market by slug.
func (c *GammaClient) GetMarketBySlug(slug string) (*Market, error) {
	params := QueryParams{"slug": slug}
	data, err := c.get("/markets" + BuildQueryString(params))
	if err != nil {
		return nil, err
	}

	var markets []Market
	if err := json.Unmarshal(data, &markets); err != nil {
		return nil, fmt.Errorf("parsing markets: %w", err)
	}
	if len(markets) == 0 {
		return nil, fmt.Errorf("market not found: %s", slug)
	}
	return &markets[0], nil
}

// ListMarkets retrieves a list of markets with optional filters.
func (c *GammaClient) ListMarkets(params MarketListParams) ([]Market, error) {
	data, err := c.get("/markets" + BuildQueryString(params.ToQueryParams()))
	if err != nil {
		return nil, err
	}

	var markets []Market
	if err := json.Unmarshal(data, &markets); err != nil {
		return nil, fmt.Errorf("parsing markets: %w", err)
	}
	return markets, nil
}

// SearchMarkets performs client-side filtering of markets.
func (c *GammaClient) SearchMarkets(query string, limit int) ([]Market, error) {
	// Fetch a large batch to filter from
	fetchLimit := limit * 10
	if fetchLimit < 100 {
		fetchLimit = 100
	}
	if fetchLimit > 500 {
		fetchLimit = 500
	}

	markets, err := c.ListMarkets(MarketListParams{
		Limit:  fetchLimit,
		Active: BoolPtr(true),
		Closed: BoolPtr(false),
	})
	if err != nil {
		return nil, err
	}

	filtered := FilterMarkets(markets, query)
	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}
	return filtered, nil
}

// GetEvent retrieves a single event by ID.
func (c *GammaClient) GetEvent(id string) (*Event, error) {
	data, err := c.get("/events/" + id)
	if err != nil {
		return nil, err
	}

	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("parsing event: %w", err)
	}
	return &event, nil
}

// GetEventBySlug retrieves a single event by slug.
func (c *GammaClient) GetEventBySlug(slug string) (*Event, error) {
	params := QueryParams{"slug": slug}
	data, err := c.get("/events" + BuildQueryString(params))
	if err != nil {
		return nil, err
	}

	var events []Event
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, fmt.Errorf("parsing events: %w", err)
	}
	if len(events) == 0 {
		return nil, fmt.Errorf("event not found: %s", slug)
	}
	return &events[0], nil
}

// ListEvents retrieves a list of events with optional filters.
func (c *GammaClient) ListEvents(params EventListParams) ([]Event, error) {
	data, err := c.get("/events" + BuildQueryString(params.ToQueryParams()))
	if err != nil {
		return nil, err
	}

	var events []Event
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, fmt.Errorf("parsing events: %w", err)
	}
	return events, nil
}

// ListTags retrieves all available tags.
func (c *GammaClient) ListTags() ([]Tag, error) {
	data, err := c.get("/tags")
	if err != nil {
		return nil, err
	}

	var tags []Tag
	if err := json.Unmarshal(data, &tags); err != nil {
		return nil, fmt.Errorf("parsing tags: %w", err)
	}
	return tags, nil
}

// GetTag retrieves a single tag by slug.
func (c *GammaClient) GetTag(slug string) (*Tag, error) {
	data, err := c.get("/tags/" + slug)
	if err != nil {
		return nil, err
	}

	var tag Tag
	if err := json.Unmarshal(data, &tag); err != nil {
		return nil, fmt.Errorf("parsing tag: %w", err)
	}
	return &tag, nil
}

// Ping checks if the Gamma API is reachable.
func (c *GammaClient) Ping() error {
	_, err := c.get("/markets?limit=1")
	return err
}

// ListSeries retrieves a list of series.
func (c *GammaClient) ListSeries(limit, offset int) ([]Series, error) {
	params := QueryParams{}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	if offset > 0 {
		params["offset"] = fmt.Sprintf("%d", offset)
	}
	data, err := c.get("/series" + BuildQueryString(params))
	if err != nil {
		return nil, err
	}

	var series []Series
	if err := json.Unmarshal(data, &series); err != nil {
		return nil, fmt.Errorf("parsing series: %w", err)
	}
	return series, nil
}

// GetSeries retrieves a single series by ID.
func (c *GammaClient) GetSeries(id string) (*Series, error) {
	data, err := c.get("/series/" + id)
	if err != nil {
		return nil, err
	}

	var series Series
	if err := json.Unmarshal(data, &series); err != nil {
		return nil, fmt.Errorf("parsing series: %w", err)
	}
	return &series, nil
}

// ListComments retrieves comments for an entity.
func (c *GammaClient) ListComments(entityType, entityID string) ([]Comment, error) {
	params := QueryParams{}
	if entityType != "" {
		params["entity_type"] = entityType
	}
	if entityID != "" {
		params["entity_id"] = entityID
	}
	data, err := c.get("/comments" + BuildQueryString(params))
	if err != nil {
		return nil, err
	}

	var comments []Comment
	if err := json.Unmarshal(data, &comments); err != nil {
		return nil, fmt.Errorf("parsing comments: %w", err)
	}
	return comments, nil
}

// GetComment retrieves a single comment by ID.
func (c *GammaClient) GetComment(id string) (*Comment, error) {
	data, err := c.get("/comments/" + id)
	if err != nil {
		return nil, err
	}

	var comment Comment
	if err := json.Unmarshal(data, &comment); err != nil {
		return nil, fmt.Errorf("parsing comment: %w", err)
	}
	return &comment, nil
}

// GetUserComments retrieves all comments by a user address.
func (c *GammaClient) GetUserComments(address string) ([]Comment, error) {
	data, err := c.get("/comments/user/" + address)
	if err != nil {
		return nil, err
	}

	var comments []Comment
	if err := json.Unmarshal(data, &comments); err != nil {
		return nil, fmt.Errorf("parsing comments: %w", err)
	}
	return comments, nil
}

// GetProfile retrieves a user profile by address.
func (c *GammaClient) GetProfile(address string) (*Profile, error) {
	data, err := c.get("/profiles/" + address)
	if err != nil {
		return nil, err
	}

	var profile Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("parsing profile: %w", err)
	}
	return &profile, nil
}

// ListSports retrieves all sports.
func (c *GammaClient) ListSports() ([]Sport, error) {
	data, err := c.get("/sports")
	if err != nil {
		return nil, err
	}

	var sports []Sport
	if err := json.Unmarshal(data, &sports); err != nil {
		return nil, fmt.Errorf("parsing sports: %w", err)
	}
	return sports, nil
}

// GetMarketTypes retrieves all sports market types.
func (c *GammaClient) GetMarketTypes() ([]SportMarketType, error) {
	data, err := c.get("/sports/market-types")
	if err != nil {
		return nil, err
	}

	var types []SportMarketType
	if err := json.Unmarshal(data, &types); err != nil {
		return nil, fmt.Errorf("parsing market types: %w", err)
	}
	return types, nil
}

// ListTeams retrieves teams with optional league filter.
func (c *GammaClient) ListTeams(league string, limit int) ([]Team, error) {
	params := QueryParams{}
	if league != "" {
		params["league"] = league
	}
	if limit > 0 {
		params["limit"] = fmt.Sprintf("%d", limit)
	}
	data, err := c.get("/teams" + BuildQueryString(params))
	if err != nil {
		return nil, err
	}

	var teams []Team
	if err := json.Unmarshal(data, &teams); err != nil {
		return nil, fmt.Errorf("parsing teams: %w", err)
	}
	return teams, nil
}
