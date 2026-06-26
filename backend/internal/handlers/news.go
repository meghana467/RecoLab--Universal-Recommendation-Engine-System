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

// GetTrendingNews handles GET /api/demo/trending-news
func (h *NewsHandler) GetTrendingNews(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	scores, err := h.gorseClient.GetPopular("news", limit)
	if err != nil || len(scores) == 0 {
		scores = h.itemStore.GetFallbackNewsScores(limit)
	}

	scores = rules.DeduplicateItems(scores)

	recs := make([]gin.H, 0, len(scores))
	for _, s := range scores {
		meta := h.itemStore.GetItem(s.Id)
		if meta == nil {
			continue
		}
		recs = append(recs, gin.H{
			"item_id":  s.Id,
			"title":    meta.Comment,
			"score":    s.Score,
			"reason":   "Trending based on recent reads and likes across all users",
			"labels":   meta.Labels,
			"category": "news",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"scenario":        "trending_news",
		"tenant_id":       h.tenantID,
		"count":           len(recs),
		"recommendations": recs,
	})
}

// GetPersonalizedNews handles GET /api/demo/personalized-news/:userId
func (h *NewsHandler) GetPersonalizedNews(c *gin.Context) {
	userId := c.Param("userId")
	limit := 10
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	// Resolve user labels from local store first (always available), then Gorse
	var userLabels []string
	if localUser := h.itemStore.GetUser(userId); localUser != nil {
		userLabels = localUser.Labels
	}
	if len(userLabels) == 0 {
		if gorseUser, _ := h.gorseClient.GetUser(userId); gorseUser != nil {
			userLabels = gorseUser.Labels
		}
	}

	// Try Gorse personalized recommendations, then strictly filter by user labels
	gorseScores, _ := h.gorseClient.GetRecommendations(userId, "news", limit*3)
	var scores []gorse.Score
	if len(userLabels) > 0 {
		scores = filterByLabels(gorseScores, userLabels, h.itemStore, limit)
	} else {
		scores = gorseScores
	}

	// Always pad with label-matched fallback items to reach the requested limit
	if len(scores) < limit && len(userLabels) > 0 {
		existing := make(map[string]bool, len(scores))
		for _, s := range scores {
			existing[s.Id] = true
		}
		fallback := h.itemStore.GetPersonalizedFallback("news", userLabels, limit)
		for _, s := range fallback {
			if !existing[s.Id] {
				scores = append(scores, s)
				existing[s.Id] = true
			}
			if len(scores) >= limit {
				break
			}
		}
	}

	scores = rules.DeduplicateItems(scores)

	recs := make([]gin.H, 0, len(scores))
	for _, s := range scores {
		meta := h.itemStore.GetItem(s.Id)
		if meta == nil {
			continue
		}
		reason := buildNewsReason(userLabels, meta.Labels)
		recs = append(recs, gin.H{
			"item_id":  s.Id,
			"title":    meta.Comment,
			"score":    s.Score,
			"reason":   reason,
			"labels":   meta.Labels,
			"category": "news",
		})
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
		"count":            len(recs),
		"recommendations":  recs,
	})
}

// filterByLabels keeps only items whose labels overlap with userLabels, up to limit.
func filterByLabels(scores []gorse.Score, userLabels []string, is *store.ItemStore, limit int) []gorse.Score {
	labelSet := make(map[string]bool, len(userLabels))
	for _, l := range userLabels {
		labelSet[l] = true
	}
	var out []gorse.Score
	for _, s := range scores {
		meta := is.GetItem(s.Id)
		if meta == nil {
			continue
		}
		for _, l := range meta.Labels {
			if labelSet[l] {
				out = append(out, s)
				break
			}
		}
		if len(out) >= limit {
			break
		}
	}
	return out
}

// buildNewsReason generates a human-readable explanation based on label overlap.
func buildNewsReason(userLabels, itemLabels []string) string {
	labelSet := make(map[string]bool, len(userLabels))
	for _, l := range userLabels {
		labelSet[l] = true
	}
	var matched []string
	for _, l := range itemLabels {
		if labelSet[l] {
			matched = append(matched, l)
		}
	}
	if len(matched) == 0 {
		return "Recommended based on your reading history"
	}
	shown := matched
	if len(shown) > 3 {
		shown = shown[:3]
	}
	return "Matches your interests: " + strings.Join(shown, ", ")
}
