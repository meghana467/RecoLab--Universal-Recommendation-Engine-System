package gorse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/recolab/backend/internal/config"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type User struct {
	UserId  string   `json:"UserId"`
	Labels  []string `json:"Labels"`
	Comment string   `json:"Comment"`
}

type Item struct {
	ItemId     string   `json:"ItemId"`
	Labels     []string `json:"Labels"`
	Categories []string `json:"Categories"`
	Timestamp  string   `json:"Timestamp,omitempty"`
	Comment    string   `json:"Comment"`
	IsHidden   bool     `json:"IsHidden"`
}

type Feedback struct {
	FeedbackType string `json:"FeedbackType"`
	UserId       string `json:"UserId"`
	ItemId       string `json:"ItemId"`
	Timestamp    string `json:"Timestamp"`
}

type Score struct {
	Id    string  `json:"Id"`
	Score float64 `json:"Score"`
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		baseURL: cfg.GorseAPIURL,
		apiKey:  cfg.GorseAPIKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) do(method, path string, body interface{}) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshal body: %w", err)
		}
		reqBody = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}
	return data, resp.StatusCode, nil
}

func (c *Client) InsertUsers(users []User) error {
	_, status, err := c.do("POST", "/api/users", users)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("gorse InsertUsers returned status %d", status)
	}
	return nil
}

func (c *Client) InsertItems(items []Item) error {
	_, status, err := c.do("POST", "/api/items", items)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("gorse InsertItems returned status %d", status)
	}
	return nil
}

func (c *Client) InsertFeedback(feedback []Feedback) error {
	_, status, err := c.do("POST", "/api/feedback", feedback)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("gorse InsertFeedback returned status %d", status)
	}
	return nil
}

func (c *Client) GetRecommendations(userId string, category string, n int) ([]Score, error) {
	path := fmt.Sprintf("/api/recommend/%s?n=%d", userId, n)
	if category != "" {
		path += "&category=" + category
	}
	data, status, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("gorse GetRecommendations returned status %d", status)
	}
	var scores []Score
	if err := json.Unmarshal(data, &scores); err != nil {
		return nil, fmt.Errorf("unmarshal recommendations: %w (body: %s)", err, string(data))
	}
	return scores, nil
}

func (c *Client) GetPopular(category string, n int) ([]Score, error) {
	path := fmt.Sprintf("/api/popular?n=%d", n)
	if category != "" {
		path += "&category=" + category
	}
	data, status, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("gorse GetPopular returned status %d", status)
	}
	var scores []Score
	if err := json.Unmarshal(data, &scores); err != nil {
		return nil, fmt.Errorf("unmarshal popular: %w (body: %s)", err, string(data))
	}
	return scores, nil
}

func (c *Client) GetLatest(category string, n int) ([]Score, error) {
	path := fmt.Sprintf("/api/latest?n=%d", n)
	if category != "" {
		path += "&category=" + category
	}
	data, status, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("gorse GetLatest returned status %d", status)
	}
	var scores []Score
	if err := json.Unmarshal(data, &scores); err != nil {
		return nil, fmt.Errorf("unmarshal latest: %w (body: %s)", err, string(data))
	}
	return scores, nil
}

func (c *Client) GetSimilarItems(itemId string, category string, n int) ([]Score, error) {
	path := fmt.Sprintf("/api/item/%s/neighbors?n=%d", itemId, n)
	if category != "" {
		path += "&category=" + category
	}
	data, status, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("gorse GetSimilarItems returned status %d", status)
	}
	var scores []Score
	if err := json.Unmarshal(data, &scores); err != nil {
		return nil, fmt.Errorf("unmarshal similar items: %w (body: %s)", err, string(data))
	}
	return scores, nil
}

func (c *Client) GetNeighbors(userId string, n int) ([]Score, error) {
	path := fmt.Sprintf("/api/user/%s/neighbors?n=%d", userId, n)
	data, status, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("gorse GetNeighbors returned status %d", status)
	}
	var scores []Score
	if err := json.Unmarshal(data, &scores); err != nil {
		return nil, fmt.Errorf("unmarshal neighbors: %w (body: %s)", err, string(data))
	}
	return scores, nil
}

func (c *Client) GetUser(userId string) (*User, error) {
	data, status, err := c.do("GET", "/api/user/"+userId, nil)
	if err != nil {
		return nil, err
	}
	if status == 404 {
		return nil, nil
	}
	if status >= 400 {
		return nil, fmt.Errorf("gorse GetUser returned status %d", status)
	}
	var user User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("unmarshal user: %w", err)
	}
	return &user, nil
}

func (c *Client) GetStats() (map[string]interface{}, error) {
	data, status, err := c.do("GET", "/api/dashboard/cluster", nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("gorse GetStats returned status %d", status)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal stats: %w", err)
	}
	return result, nil
}
