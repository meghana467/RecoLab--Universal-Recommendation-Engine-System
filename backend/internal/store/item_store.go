package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/recolab/backend/internal/gorse"
)

// ItemMeta holds metadata for a Gorse item.
type ItemMeta struct {
	ItemId     string   `json:"ItemId"`
	Labels     []string `json:"Labels"`
	Categories []string `json:"Categories"`
	Timestamp  string   `json:"Timestamp"`
	Comment    string   `json:"Comment"`
	IsHidden   bool     `json:"IsHidden"`
}

// UserMeta holds metadata for a Gorse user.
type UserMeta struct {
	UserId  string   `json:"UserId"`
	Labels  []string `json:"Labels"`
	Comment string   `json:"Comment"`
}

// ItemStore is an in-memory store for item and user metadata.
type ItemStore struct {
	mu    sync.RWMutex
	items map[string]*ItemMeta
	users map[string]*UserMeta

	// feedback counts for stats
	feedbackCount int
}

func NewItemStore() *ItemStore {
	return &ItemStore{
		items: make(map[string]*ItemMeta),
		users: make(map[string]*UserMeta),
	}
}

func (s *ItemStore) LoadFromDataDir(dataDir string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	itemFiles := []string{"news_items.json", "product_items.json", "expo_booths.json"}
	for _, f := range itemFiles {
		path := filepath.Join(dataDir, f)
		if err := s.loadItems(path); err != nil {
			fmt.Printf("Warning: could not load %s: %v\n", path, err)
		}
	}

	userPath := filepath.Join(dataDir, "users.json")
	if err := s.loadUsers(userPath); err != nil {
		fmt.Printf("Warning: could not load users.json: %v\n", err)
	}

	feedbackPath := filepath.Join(dataDir, "feedback.json")
	if err := s.countFeedback(feedbackPath); err != nil {
		fmt.Printf("Warning: could not count feedback.json: %v\n", err)
	}

	return nil
}

func (s *ItemStore) loadItems(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var items []ItemMeta
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}
	for i := range items {
		s.items[items[i].ItemId] = &items[i]
	}
	return nil
}

func (s *ItemStore) loadUsers(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var users []UserMeta
	if err := json.Unmarshal(data, &users); err != nil {
		return err
	}
	for i := range users {
		s.users[users[i].UserId] = &users[i]
	}
	return nil
}

func (s *ItemStore) countFeedback(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var feedback []interface{}
	if err := json.Unmarshal(data, &feedback); err != nil {
		return err
	}
	s.feedbackCount = len(feedback)
	return nil
}

func (s *ItemStore) GetItem(id string) *ItemMeta {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.items[id]
}

func (s *ItemStore) GetUser(id string) *UserMeta {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.users[id]
}

func (s *ItemStore) TotalItems() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

func (s *ItemStore) TotalUsers() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.users)
}

func (s *ItemStore) TotalFeedback() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.feedbackCount
}

func (s *ItemStore) AllItems() []*ItemMeta {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*ItemMeta, 0, len(s.items))
	for _, v := range s.items {
		out = append(out, v)
	}
	return out
}

// GetPersonalizedFallback returns items from a category ranked by label overlap with userLabels.
func (s *ItemStore) GetPersonalizedFallback(category string, userLabels []string, n int) []gorse.Score {
	s.mu.RLock()
	defer s.mu.RUnlock()

	labelSet := make(map[string]bool, len(userLabels))
	for _, l := range userLabels {
		labelSet[l] = true
	}

	type scored struct {
		id    string
		score float64
	}
	var candidates []scored
	for id, meta := range s.items {
		inCat := false
		for _, c := range meta.Categories {
			if c == category {
				inCat = true
				break
			}
		}
		if !inCat {
			continue
		}
		overlap := 0
		for _, l := range meta.Labels {
			if labelSet[l] {
				overlap++
			}
		}
		if overlap > 0 {
			candidates = append(candidates, scored{id: id, score: float64(overlap) * 0.2})
		}
	}
	// sort descending
	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].score > candidates[i].score {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}
	if len(candidates) > n {
		candidates = candidates[:n]
	}
	out := make([]gorse.Score, 0, len(candidates))
	for _, c := range candidates {
		out = append(out, gorse.Score{Id: c.id, Score: c.score})
	}
	return out
}

func (s *ItemStore) GetFallbackNewsScores(n int) []gorse.Score {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []gorse.Score
	score := 0.99
	for id, meta := range s.items {
		for _, cat := range meta.Categories {
			if cat == "news" {
				out = append(out, gorse.Score{Id: id, Score: score})
				score -= 0.01
				break
			}
		}
		if len(out) >= n {
			break
		}
	}
	return out
}

func (s *ItemStore) GetFallbackProductScores(n int) []gorse.Score {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []gorse.Score
	score := 0.99
	for id, meta := range s.items {
		for _, cat := range meta.Categories {
			if cat == "products" {
				out = append(out, gorse.Score{Id: id, Score: score})
				score -= 0.01
				break
			}
		}
		if len(out) >= n {
			break
		}
	}
	return out
}

func (s *ItemStore) GetFallbackExpoScores(n int) []gorse.Score {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []gorse.Score
	score := 0.99
	for id, meta := range s.items {
		for _, cat := range meta.Categories {
			if cat == "expo_booth" {
				out = append(out, gorse.Score{Id: id, Score: score})
				score -= 0.01
				break
			}
		}
		if len(out) >= n {
			break
		}
	}
	return out
}

func (s *ItemStore) GetFallbackBuyerScores(n int) []gorse.Score {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []gorse.Score
	score := 0.95
	for id := range s.users {
		out = append(out, gorse.Score{Id: id, Score: score})
		score -= 0.01
		if len(out) >= n {
			break
		}
	}
	return out
}
