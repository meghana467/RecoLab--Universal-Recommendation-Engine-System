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

// GetProductRecommendations handles GET /api/demo/products/:userId/recommendations
func (h *ProductHandler) GetProductRecommendations(c *gin.Context) {
	userId := c.Param("userId")
	limit := 10
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	var userLabels []string
	if localUser := h.itemStore.GetUser(userId); localUser != nil {
		userLabels = localUser.Labels
	}
	if len(userLabels) == 0 {
		if gorseUser, _ := h.gorseClient.GetUser(userId); gorseUser != nil {
			userLabels = gorseUser.Labels
		}
	}

	gorseScores, _ := h.gorseClient.GetRecommendations(userId, "products", limit*3)
	var scores []gorse.Score
	if len(userLabels) > 0 {
		scores = filterProductsByLabels(gorseScores, userLabels, h.itemStore, limit)
	} else {
		scores = gorseScores
	}

	// Pad with label-matched fallback if needed
	if len(scores) < limit && len(userLabels) > 0 {
		existing := make(map[string]bool, len(scores))
		for _, s := range scores {
			existing[s.Id] = true
		}
		fallback := h.itemStore.GetPersonalizedFallback("products", userLabels, limit)
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
		reason := buildProductReason(userLabels, meta.Labels)
		recs = append(recs, gin.H{
			"item_id":  s.Id,
			"title":    meta.Comment,
			"score":    s.Score,
			"reason":   reason,
			"labels":   meta.Labels,
			"category": "products",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":          userId,
		"scenario":         "product_recommendations",
		"tenant_id":        h.tenantID,
		"user_interests":   userLabels,
		"count":            len(recs),
		"recommendations":  recs,
	})
}

func buildProductReason(userLabels, itemLabels []string) string {
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
		return "Recommended based on your purchase history"
	}
	shown := matched
	if len(shown) > 3 {
		shown = shown[:3]
	}
	return "Matches your interests: " + strings.Join(shown, ", ")
}

// filterProductsByLabels keeps products whose labels overlap with userLabels.
func filterProductsByLabels(scores []gorse.Score, userLabels []string, is *store.ItemStore, limit int) []gorse.Score {
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

// GetSimilarProducts handles GET /api/demo/products/:itemId/similar
func (h *ProductHandler) GetSimilarProducts(c *gin.Context) {
	itemId := c.Param("itemId")
	limit := 10
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	scores, err := h.gorseClient.GetSimilarItems(itemId, "products", limit)
	if err != nil || len(scores) == 0 {
		scores = h.itemStore.GetFallbackProductScores(limit)
	}

	scores = rules.DeduplicateItems(scores)

	sourceMeta := h.itemStore.GetItem(itemId)
	var sourceTitle string
	if sourceMeta != nil {
		sourceTitle = sourceMeta.Comment
	}

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
			"reason":   "Similar to: " + sourceTitle,
			"labels":   meta.Labels,
			"category": "products",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"source_item_id":  itemId,
		"source_title":    sourceTitle,
		"scenario":        "similar_products",
		"tenant_id":       h.tenantID,
		"count":           len(recs),
		"recommendations": recs,
	})
}
