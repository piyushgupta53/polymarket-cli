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

// getJSON performs a GET request and streams the JSON response into dest.
func (c *GammaClient) getJSON(path string, dest any) error {
	url := c.baseURL + path
	log.Debug("GET", "url", url)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}
	return nil
}

// GetMarket retrieves a single market by ID.
func (c *GammaClient) GetMarket(id string) (*Market, error) {
	var market Market
	if err := c.getJSON("/markets/"+id, &market); err != nil {
		return nil, err
	}
	return &market, nil
}

// GetMarketBySlug retrieves a single market by slug.
func (c *GammaClient) GetMarketBySlug(slug string) (*Market, error) {
	params := QueryParams{"slug": slug}
	var markets []Market
	if err := c.getJSON("/markets"+BuildQueryString(params), &markets); err != nil {
		return nil, err
	}
	if len(markets) == 0 {
		return nil, fmt.Errorf("market not found: %s", slug)
	}
	return &markets[0], nil
}

// ListMarkets retrieves a list of markets with optional filters.
func (c *GammaClient) ListMarkets(params MarketListParams) ([]Market, error) {
	var markets []Market
	if err := c.getJSON("/markets"+BuildQueryString(params.ToQueryParams()), &markets); err != nil {
		return nil, err
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
	var event Event
	if err := c.getJSON("/events/"+id, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

// GetEventBySlug retrieves a single event by slug.
func (c *GammaClient) GetEventBySlug(slug string) (*Event, error) {
	params := QueryParams{"slug": slug}
	var events []Event
	if err := c.getJSON("/events"+BuildQueryString(params), &events); err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, fmt.Errorf("event not found: %s", slug)
	}
	return &events[0], nil
}

// ListEvents retrieves a list of events with optional filters.
func (c *GammaClient) ListEvents(params EventListParams) ([]Event, error) {
	var events []Event
	if err := c.getJSON("/events"+BuildQueryString(params.ToQueryParams()), &events); err != nil {
		return nil, err
	}
	return events, nil
}

// ListTags retrieves all available tags.
func (c *GammaClient) ListTags() ([]Tag, error) {
	var tags []Tag
	if err := c.getJSON("/tags", &tags); err != nil {
		return nil, err
	}
	return tags, nil
}

// GetTag retrieves a single tag by slug.
func (c *GammaClient) GetTag(slug string) (*Tag, error) {
	var tag Tag
	if err := c.getJSON("/tags/"+slug, &tag); err != nil {
		return nil, err
	}
	return &tag, nil
}

// Ping checks if the Gamma API is reachable.
func (c *GammaClient) Ping() error {
	var discard json.RawMessage
	return c.getJSON("/markets?limit=1", &discard)
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
	var series []Series
	if err := c.getJSON("/series"+BuildQueryString(params), &series); err != nil {
		return nil, err
	}
	return series, nil
}

// GetSeries retrieves a single series by ID.
func (c *GammaClient) GetSeries(id string) (*Series, error) {
	var series Series
	if err := c.getJSON("/series/"+id, &series); err != nil {
		return nil, err
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
	var comments []Comment
	if err := c.getJSON("/comments"+BuildQueryString(params), &comments); err != nil {
		return nil, err
	}
	return comments, nil
}

// GetComment retrieves a single comment by ID.
func (c *GammaClient) GetComment(id string) (*Comment, error) {
	var comment Comment
	if err := c.getJSON("/comments/"+id, &comment); err != nil {
		return nil, err
	}
	return &comment, nil
}

// GetUserComments retrieves all comments by a user address.
func (c *GammaClient) GetUserComments(address string) ([]Comment, error) {
	var comments []Comment
	if err := c.getJSON("/comments/user/"+address, &comments); err != nil {
		return nil, err
	}
	return comments, nil
}

// GetProfile retrieves a user profile by address.
func (c *GammaClient) GetProfile(address string) (*Profile, error) {
	var profile Profile
	if err := c.getJSON("/profiles/"+address, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

// ListSports retrieves all sports.
func (c *GammaClient) ListSports() ([]Sport, error) {
	var sports []Sport
	if err := c.getJSON("/sports", &sports); err != nil {
		return nil, err
	}
	return sports, nil
}

// GetMarketTypes retrieves all sports market types.
func (c *GammaClient) GetMarketTypes() ([]SportMarketType, error) {
	var types []SportMarketType
	if err := c.getJSON("/sports/market-types", &types); err != nil {
		return nil, err
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
	var teams []Team
	if err := c.getJSON("/teams"+BuildQueryString(params), &teams); err != nil {
		return nil, err
	}
	return teams, nil
}
