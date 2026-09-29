package store

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

// ItemStore is the in-memory metadata store for items and users.
type ItemStore struct {
	mu            sync.RWMutex
	items         map[string]*ItemMeta
	users         map[string]*UserMeta
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

	for _, f := range []string{"news_items.json", "product_items.json", "expo_booths.json", "sessions.json"} {
		if err := s.loadItems(filepath.Join(dataDir, f)); err != nil {
			fmt.Printf("Warning: could not load %s: %v\n", f, err)
		}
	}
	if err := s.loadUsers(filepath.Join(dataDir, "users.json")); err != nil {
		fmt.Printf("Warning: could not load users.json: %v\n", err)
	}
	if err := s.countFeedback(filepath.Join(dataDir, "feedback.json")); err != nil {
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
	var fb []interface{}
	if err := json.Unmarshal(data, &fb); err != nil {
		return err
	}
	s.feedbackCount = len(fb)
	return nil
}

// UpsertUser adds or updates a user in the local store.
// Called after onboarding so label-based fallbacks work immediately.
func (s *ItemStore) UpsertUser(u UserMeta) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := u
	s.users[u.UserId] = &cp
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

// sortedItemIDs returns item IDs in stable alphabetical order.
func (s *ItemStore) sortedItemIDs() []string {
	ids := make([]string, 0, len(s.items))
	for id := range s.items {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// sortedUserIDs returns user IDs in stable alphabetical order.
func (s *ItemStore) sortedUserIDs() []string {
	ids := make([]string, 0, len(s.users))
	for id := range s.users {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// ─────────────────────────────────────────────────────────────────────────────
// Solution 1 + 2: GetStratifiedFallback
// Groups items in a category by their subcategory (non-primary label or
// mind-subcategory), then picks top-N from each subcategory proportionally.
// This ensures a new user who selected ["sports","finance"] gets articles from
// both topics instead of all-sports or all-finance depending on alphabetic luck.
// ─────────────────────────────────────────────────────────────────────────────
func (s *ItemStore) GetStratifiedFallback(category string, userLabels []string, n int) []gorse.Score {
	s.mu.RLock()
	defer s.mu.RUnlock()

	labelSet := make(map[string]bool, len(userLabels))
	for _, l := range userLabels {
		labelSet[strings.ToLower(l)] = true
	}

	// Build bucket: subcategory → []scored item
	type scored struct {
		id     string
		score  float64
		subcat string
	}
	buckets := make(map[string][]scored)

	for _, id := range s.sortedItemIDs() {
		meta := s.items[id]
		if !hasCategory(meta.Categories, category) {
			continue
		}
		overlap := 0
		for _, l := range meta.Labels {
			if labelSet[strings.ToLower(l)] {
				overlap++
			}
		}
		// Also count matching subcategory (MIND subcategory stored in Categories)
		subcat := subCategory(meta.Categories, category)
		if subcat != "" && labelSet[subcat] {
			overlap++
		}
		if overlap == 0 {
			continue
		}
		sc := scored{
			id:     id,
			score:  0.70 + float64(overlap)*0.05,
			subcat: subcat,
		}
		if subcat == "" {
			subcat = "_other"
		}
		buckets[subcat] = append(buckets[subcat], sc)
	}

	if len(buckets) == 0 {
		// No label overlap at all — fall through to plain popularity fallback
		return s.getPlainFallback(category, n)
	}

	// Sort each bucket by score desc
	for k := range buckets {
		sort.Slice(buckets[k], func(i, j int) bool {
			return buckets[k][i].score > buckets[k][j].score
		})
	}

	// Round-robin across buckets so results are diversified
	bucketKeys := make([]string, 0, len(buckets))
	for k := range buckets {
		bucketKeys = append(bucketKeys, k)
	}
	sort.Strings(bucketKeys)

	// Only use buckets whose subcategory matches userLabels (or _other)
	var activeKeys []string
	for _, k := range bucketKeys {
		if k == "_other" || labelSet[k] {
			activeKeys = append(activeKeys, k)
		}
	}
	if len(activeKeys) == 0 {
		activeKeys = bucketKeys
	}

	ptrs := make(map[string]int, len(activeKeys))
	var out []gorse.Score
	seen := make(map[string]bool)

	for len(out) < n {
		added := false
		for _, k := range activeKeys {
			p := ptrs[k]
			if p >= len(buckets[k]) {
				continue
			}
			item := buckets[k][p]
			ptrs[k]++
			if seen[item.id] {
				continue
			}
			seen[item.id] = true
			out = append(out, gorse.Score{Id: item.id, Score: item.score})
			added = true
			if len(out) >= n {
				break
			}
		}
		if !added {
			break // all buckets exhausted
		}
	}
	return out
}

// ─────────────────────────────────────────────────────────────────────────────
// Solution 5: GetCrossScenarioRecs
// When a user has no feedback in a target category, use their labels (which may
// come from activity in OTHER scenarios) to filter target-category items.
// E.g. a user who clicked tech news automatically gets tech expo booths.
// Called by handlers after Gorse returns empty — no extra API call needed.
// ─────────────────────────────────────────────────────────────────────────────
func (s *ItemStore) GetCrossScenarioRecs(targetCategory string, userLabels []string, gorseScores []gorse.Score, n int) []gorse.Score {
	seen := make(map[string]bool, len(gorseScores))
	var out []gorse.Score
	for _, sc := range gorseScores {
		seen[sc.Id] = true
		out = append(out, sc)
	}
	if len(out) >= n {
		return out[:n]
	}
	// Fill remainder using stratified label fallback for the target category
	for _, f := range s.GetStratifiedFallback(targetCategory, userLabels, n*2) {
		if !seen[f.Id] {
			out = append(out, f)
			seen[f.Id] = true
		}
		if len(out) >= n {
			break
		}
	}
	return out
}

// ─────────────────────────────────────────────────────────────────────────────
// Solution Label-based Item Similarity (cold-start for new items)
// When Gorse returns empty neighbors (new item), compute Jaccard similarity
// between the target item's labels and all other items in the same category.
// ─────────────────────────────────────────────────────────────────────────────
func (s *ItemStore) GetSimilarByLabels(itemId string, category string, n int) []gorse.Score {
	s.mu.RLock()
	defer s.mu.RUnlock()

	source := s.items[itemId]
	if source == nil {
		return nil
	}

	sourceSet := make(map[string]bool, len(source.Labels))
	for _, l := range source.Labels {
		sourceSet[strings.ToLower(l)] = true
	}
	// Also include subcategory for similarity
	if sub := subCategory(source.Categories, category); sub != "" {
		sourceSet[sub] = true
	}

	type scored struct {
		id    string
		score float64
	}
	var candidates []scored

	for _, id := range s.sortedItemIDs() {
		if id == itemId {
			continue
		}
		meta := s.items[id]
		if !hasCategory(meta.Categories, category) {
			continue
		}
		// Jaccard similarity: |A ∩ B| / |A ∪ B|
		intersection := 0
		candidateSet := make(map[string]bool, len(meta.Labels))
		for _, l := range meta.Labels {
			ll := strings.ToLower(l)
			candidateSet[ll] = true
			if sourceSet[ll] {
				intersection++
			}
		}
		if sub := subCategory(meta.Categories, category); sub != "" {
			candidateSet[sub] = true
			if sourceSet[sub] {
				intersection++
			}
		}
		union := len(sourceSet) + len(candidateSet) - intersection
		if union == 0 || intersection == 0 {
			continue
		}
		jaccard := float64(intersection) / float64(union)
		candidates = append(candidates, scored{id: id, score: jaccard})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})
	if len(candidates) > n {
		candidates = candidates[:n]
	}
	out := make([]gorse.Score, len(candidates))
	for i, c := range candidates {
		out[i] = gorse.Score{Id: c.id, Score: c.score}
	}
	return out
}

// GetPersonalizedFallback returns items ranked by label overlap with userLabels.
func (s *ItemStore) GetPersonalizedFallback(category string, userLabels []string, n int) []gorse.Score {
	s.mu.RLock()
	defer s.mu.RUnlock()

	labelSet := make(map[string]bool, len(userLabels))
	for _, l := range userLabels {
		labelSet[strings.ToLower(l)] = true
	}

	type scored struct {
		id    string
		score float64
	}
	var candidates []scored
	for _, id := range s.sortedItemIDs() {
		meta := s.items[id]
		if !hasCategory(meta.Categories, category) {
			continue
		}
		overlap := 0
		for _, l := range meta.Labels {
			if labelSet[strings.ToLower(l)] {
				overlap++
			}
		}
		if sub := subCategory(meta.Categories, category); sub != "" && labelSet[sub] {
			overlap++
		}
		if overlap > 0 {
			candidates = append(candidates, scored{id: id, score: 0.75 + float64(overlap)*0.05})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})
	if len(candidates) > n {
		candidates = candidates[:n]
	}
	out := make([]gorse.Score, len(candidates))
	for i, c := range candidates {
		out[i] = gorse.Score{Id: c.id, Score: c.score}
	}
	return out
}

func (s *ItemStore) getPlainFallback(category string, n int) []gorse.Score {
	var out []gorse.Score
	score := 0.99
	for _, id := range s.sortedItemIDs() {
		meta := s.items[id]
		if hasCategory(meta.Categories, category) {
			out = append(out, gorse.Score{Id: id, Score: score})
			score -= 0.01
		}
		if len(out) >= n {
			break
		}
	}
	return out
}

func (s *ItemStore) GetFallbackNewsScores(n int) []gorse.Score {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.getPlainFallback("news", n)
}

func (s *ItemStore) GetFallbackProductScores(n int) []gorse.Score {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.getPlainFallback("products", n)
}

func (s *ItemStore) GetFallbackExpoScores(n int) []gorse.Score {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.getPlainFallback("expo_booth", n)
}

func (s *ItemStore) GetFallbackBuyerScores(n int) []gorse.Score {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []gorse.Score
	score := 0.95
	for _, id := range s.sortedUserIDs() {
		out = append(out, gorse.Score{Id: id, Score: score})
		score -= 0.01
		if len(out) >= n {
			break
		}
	}
	return out
}

func (s *ItemStore) GetLabelMatchedBuyers(boothLabels []string, n int) []gorse.Score {
	s.mu.RLock()
	defer s.mu.RUnlock()
	matchAll := len(boothLabels) == 0
	set := make(map[string]bool, len(boothLabels))
	for _, l := range boothLabels {
		set[strings.ToLower(l)] = true
	}
	type scored struct {
		id    string
		score float64
	}
	var candidates []scored
	for _, id := range s.sortedUserIDs() {
		u := s.users[id]
		if len(u.Labels) == 0 {
			continue
		}
		if matchAll {
			candidates = append(candidates, scored{id: id, score: 0.5})
			continue
		}
		overlap := 0
		for _, l := range u.Labels {
			if set[strings.ToLower(l)] {
				overlap++
			}
		}
		if overlap > 0 {
			candidates = append(candidates, scored{id: id, score: float64(overlap) * 0.25})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})
	if len(candidates) > n {
		candidates = candidates[:n]
	}
	out := make([]gorse.Score, len(candidates))
	for i, c := range candidates {
		out[i] = gorse.Score{Id: c.id, Score: c.score}
	}
	return out
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func hasCategory(categories []string, category string) bool {
	for _, c := range categories {
		if c == category {
			return true
		}
	}
	return false
}

// subCategory returns the non-primary category (the MIND/expo subcategory).
func subCategory(categories []string, primary string) string {
	skip := map[string]bool{"news": true, "products": true, "expo_booth": true, "session": true}
	for _, c := range categories {
		if c != primary && !skip[c] {
			return c
		}
	}
	return ""
}

// jaccardSimilarity computes |A ∩ B| / |A ∪ B| between two label slices.
func jaccardSimilarity(a, b []string) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 0
	}
	setA := make(map[string]bool, len(a))
	for _, l := range a {
		setA[strings.ToLower(l)] = true
	}
	intersection := 0
	setB := make(map[string]bool, len(b))
	for _, l := range b {
		ll := strings.ToLower(l)
		setB[ll] = true
		if setA[ll] {
			intersection++
		}
	}
	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0
	}
	return math.Round(float64(intersection)/float64(union)*10000) / 10000
}

