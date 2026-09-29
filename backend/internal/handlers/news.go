package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/recolab/backend/internal/gorse"
	"github.com/recolab/backend/internal/rules"
	"github.com/recolab/backend/internal/store"
)

type NewsHandler struct {
	gorseClient *gorse.Client
	itemStore   *store.ItemStore
	tenantID    string
}

func NewNewsHandler(gc *gorse.Client, is *store.ItemStore, tenantID string) *NewsHandler {
	return &NewsHandler{gorseClient: gc, itemStore: is, tenantID: tenantID}
}

func mindCategory(cats []string) string {
	for _, c := range cats {
		if c != "news" && c != "products" && c != "expo_booth" && c != "session" {
			return c
		}
	}
	return ""
}

func newsRec(s gorse.Score, meta *store.ItemMeta, reason string) gin.H {
	return gin.H{
		"item_id":       s.Id,
		"title":         meta.Comment,
		"score":         s.Score,
		"reason":        reason,
		"labels":        meta.Labels,
		"category":      "news",
		"mind_category": mindCategory(meta.Categories),
	}
}

// GetTrendingNews handles GET /api/demo/trending-news
func (h *NewsHandler) GetTrendingNews(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	refresh := c.Query("refresh")
	poolLimit := candidateLimit(limit)

	scores, err := h.gorseClient.GetPopular("news", poolLimit)
	if err != nil || len(scores) == 0 {
		scores = h.itemStore.GetFallbackNewsScores(poolLimit)
	}
	scores = rules.DeduplicateItems(scores)
	scores = selectRefreshWindow(scores, refresh, limit)

	recs := make([]gin.H, 0, limit)
	existing := make(map[string]bool)
	for _, s := range scores {
		meta := h.itemStore.GetItem(s.Id)
		if meta == nil || existing[s.Id] {
			continue
		}
		existing[s.Id] = true
		recs = append(recs, newsRec(s, meta, "Trending based on recent reads and likes across all users"))
	}
	if len(recs) < limit {
		for _, s := range h.itemStore.GetFallbackNewsScores(poolLimit) {
			if existing[s.Id] {
				continue
			}
			meta := h.itemStore.GetItem(s.Id)
			if meta == nil {
				continue
			}
			existing[s.Id] = true
			recs = append(recs, newsRec(s, meta, "Trending based on recent reads and likes across all users"))
			if len(recs) >= limit {
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"scenario":        "trending_news",
		"tenant_id":       h.tenantID,
		"count":           len(recs),
		"recommendations": recs,
	})
}

// GetPersonalizedNews handles GET /api/demo/personalized-news/:userId
// Implements Solution 1 (label-first), Solution 2 (stratified), Solution 5 (cross-scenario).
func (h *NewsHandler) GetPersonalizedNews(c *gin.Context) {
	userId := c.Param("userId")
	limit := 10
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	refresh := c.Query("refresh")
	poolLimit := candidateLimit(limit)

	// Resolve user labels (local store is updated by onboarding instantly)
	userLabels := h.resolveUserLabels(userId)

	// Try Gorse personalised recommendations
	gorseScores, _ := h.gorseClient.GetRecommendations(userId, "news", poolLimit)

	var scores []gorse.Score
	if len(userLabels) > 0 {
		// Solution 2: filter Gorse results to user's label space
		scores = filterByLabels(gorseScores, userLabels, h.itemStore, poolLimit)
	} else {
		scores = gorseScores
	}

	// Solution 5: cross-scenario — fill from labels (which may come from expo/product activity)
	if len(scores) < limit && len(userLabels) > 0 {
		scores = h.itemStore.GetCrossScenarioRecs("news", userLabels, scores, poolLimit)
	}

	// Solution 2: stratified fallback — top items per interest subcategory
	if len(scores) < limit && len(userLabels) > 0 {
		seen := make(map[string]bool, len(scores))
		for _, s := range scores {
			seen[s.Id] = true
		}
		for _, s := range h.itemStore.GetStratifiedFallback("news", userLabels, poolLimit) {
			if !seen[s.Id] {
				scores = append(scores, s)
				seen[s.Id] = true
			}
			if len(scores) >= limit {
				break
			}
		}
	}

	// Last resort: any news
	if len(scores) < limit {
		seen := make(map[string]bool, len(scores))
		for _, s := range scores {
			seen[s.Id] = true
		}
		for _, s := range h.itemStore.GetFallbackNewsScores(poolLimit) {
			if !seen[s.Id] {
				scores = append(scores, s)
				seen[s.Id] = true
			}
			if len(scores) >= limit {
				break
			}
		}
	}

	scores = rules.DeduplicateItems(scores)
	scores = selectRefreshWindow(scores, refresh, limit)

	recs := make([]gin.H, 0, len(scores))
	for _, s := range scores {
		meta := h.itemStore.GetItem(s.Id)
		if meta == nil {
			continue
		}
		recs = append(recs, newsRec(s, meta, buildNewsReason(userLabels, meta.Labels)))
	}

	interestSummary := "general reader"
	if len(userLabels) > 0 {
		tops := userLabels
		if len(tops) > 3 {
			tops = tops[:3]
		}
		interestSummary = strings.Join(tops, ", ")
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":          userId,
		"scenario":         "personalized_news",
		"tenant_id":        h.tenantID,
		"user_interests":   userLabels,
		"interest_summary": interestSummary,
		"cold_start":       len(gorseScores) == 0,
		"count":            len(recs),
		"recommendations":  recs,
	})
}

// resolveUserLabels checks local store first (updated by onboarding), then Gorse.
func (h *NewsHandler) resolveUserLabels(userId string) []string {
	if u := h.itemStore.GetUser(userId); u != nil && len(u.Labels) > 0 {
		return u.Labels
	}
	if gu, _ := h.gorseClient.GetUser(userId); gu != nil {
		return gu.Labels
	}
	return nil
}

// filterByLabels keeps only items whose labels overlap with userLabels.
func filterByLabels(scores []gorse.Score, userLabels []string, is *store.ItemStore, limit int) []gorse.Score {
	labelSet := make(map[string]bool, len(userLabels))
	for _, l := range userLabels {
		labelSet[strings.ToLower(l)] = true
	}
	var out []gorse.Score
	for _, s := range scores {
		meta := is.GetItem(s.Id)
		if meta == nil {
			continue
		}
		for _, l := range meta.Labels {
			if labelSet[strings.ToLower(l)] {
				out = append(out, s)
				break
			}
		}
		if len(out) == 0 || out[len(out)-1].Id != s.Id {
			mc := mindCategory(meta.Categories)
			if mc != "" && labelSet[mc] {
				out = append(out, s)
			}
		}
		if len(out) >= limit {
			break
		}
	}
	return out
}

func buildNewsReason(userLabels, itemLabels []string) string {
	labelSet := make(map[string]bool, len(userLabels))
	for _, l := range userLabels {
		labelSet[strings.ToLower(l)] = true
	}
	var matched []string
	for _, l := range itemLabels {
		if labelSet[strings.ToLower(l)] {
			matched = append(matched, l)
		}
	}
	if len(matched) == 0 {
		return "Recommended based on your reading history"
	}
	if len(matched) > 3 {
		matched = matched[:3]
	}
	return "Matches your interests: " + strings.Join(matched, ", ")
}
