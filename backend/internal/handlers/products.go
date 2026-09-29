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

type ProductHandler struct {
	gorseClient *gorse.Client
	itemStore   *store.ItemStore
	tenantID    string
}

func NewProductHandler(gc *gorse.Client, is *store.ItemStore, tenantID string) *ProductHandler {
	return &ProductHandler{gorseClient: gc, itemStore: is, tenantID: tenantID}
}

// GetProductRecommendations handles GET /api/demo/products/user/:userId/recommendations
// Implements Solution 1 (label-first), Solution 2 (stratified), Solution 5 (cross-scenario).
func (h *ProductHandler) GetProductRecommendations(c *gin.Context) {
	userId := c.Param("userId")
	limit := 10
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	refresh := c.Query("refresh")
	poolLimit := candidateLimit(limit)

	userLabels := h.resolveUserLabels(userId)

	gorseScores, _ := h.gorseClient.GetRecommendations(userId, "products", poolLimit)

	var scores []gorse.Score
	if len(userLabels) > 0 {
		scores = filterProductsByLabels(gorseScores, userLabels, h.itemStore, poolLimit)
	} else {
		scores = gorseScores
	}

	// Solution 5: cross-scenario fill using labels from any activity
	if len(scores) < limit && len(userLabels) > 0 {
		scores = h.itemStore.GetCrossScenarioRecs("products", userLabels, scores, poolLimit)
	}

	// Solution 2: stratified fallback
	if len(scores) < limit && len(userLabels) > 0 {
		seen := make(map[string]bool, len(scores))
		for _, s := range scores {
			seen[s.Id] = true
		}
		for _, s := range h.itemStore.GetStratifiedFallback("products", userLabels, poolLimit) {
			if !seen[s.Id] {
				scores = append(scores, s)
				seen[s.Id] = true
			}
			if len(scores) >= limit {
				break
			}
		}
	}

	// Last resort
	if len(scores) < limit {
		seen := make(map[string]bool, len(scores))
		for _, s := range scores {
			seen[s.Id] = true
		}
		for _, s := range h.itemStore.GetFallbackProductScores(poolLimit) {
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
		recs = append(recs, gin.H{
			"item_id":  s.Id,
			"title":    meta.Comment,
			"score":    s.Score,
			"reason":   buildProductReason(userLabels, meta.Labels),
			"labels":   meta.Labels,
			"category": "products",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":         userId,
		"scenario":        "product_recommendations",
		"tenant_id":       h.tenantID,
		"user_interests":  userLabels,
		"cold_start":      len(gorseScores) == 0,
		"count":           len(recs),
		"recommendations": recs,
	})
}

// GetSimilarProducts handles GET /api/demo/products/item/:itemId/similar
// Label-based item similarity (Solution 3) kicks in when Gorse has no neighbors.
func (h *ProductHandler) GetSimilarProducts(c *gin.Context) {
	itemId := c.Param("itemId")
	limit := 10
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	// Try Gorse item neighbors first
	refresh := c.Query("refresh")
	poolLimit := candidateLimit(limit)

	scores, err := h.gorseClient.GetSimilarItems(itemId, "products", poolLimit)

	// Label-based similarity fallback for new/cold items
	if err != nil || len(scores) == 0 {
		scores = h.itemStore.GetSimilarByLabels(itemId, "products", poolLimit)
	}

	// Pad any remaining slots with generic fallback
	if len(scores) < limit {
		seen := make(map[string]bool, len(scores))
		seen[itemId] = true
		for _, s := range scores {
			seen[s.Id] = true
		}
		for _, s := range h.itemStore.GetFallbackProductScores(poolLimit) {
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

	sourceMeta := h.itemStore.GetItem(itemId)
	var sourceTitle string
	if sourceMeta != nil {
		sourceTitle = sourceMeta.Comment
	}

	usedLabelSimilarity := err != nil || len(scores) == 0
	recs := make([]gin.H, 0, len(scores))
	for _, s := range scores {
		meta := h.itemStore.GetItem(s.Id)
		if meta == nil {
			continue
		}
		reason := "Similar to: " + sourceTitle
		if usedLabelSimilarity {
			reason = "Similar labels to: " + sourceTitle
		}
		recs = append(recs, gin.H{
			"item_id":          s.Id,
			"title":            meta.Comment,
			"score":            s.Score,
			"reason":           reason,
			"labels":           meta.Labels,
			"category":         "products",
			"label_similarity": usedLabelSimilarity,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"source_item_id":  itemId,
		"source_title":    sourceTitle,
		"scenario":        "similar_products",
		"tenant_id":       h.tenantID,
		"used_label_sim":  usedLabelSimilarity,
		"count":           len(recs),
		"recommendations": recs,
	})
}

func (h *ProductHandler) resolveUserLabels(userId string) []string {
	if u := h.itemStore.GetUser(userId); u != nil && len(u.Labels) > 0 {
		return u.Labels
	}
	if gu, _ := h.gorseClient.GetUser(userId); gu != nil {
		return gu.Labels
	}
	return nil
}

func buildProductReason(userLabels, itemLabels []string) string {
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
		return "Recommended based on your purchase history"
	}
	if len(matched) > 3 {
		matched = matched[:3]
	}
	return "Matches your interests: " + strings.Join(matched, ", ")
}

func filterProductsByLabels(scores []gorse.Score, userLabels []string, is *store.ItemStore, limit int) []gorse.Score {
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
		if len(out) >= limit {
			break
		}
	}
	return out
}
