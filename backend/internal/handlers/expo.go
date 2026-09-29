package handlers

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/recolab/backend/internal/gorse"
	"github.com/recolab/backend/internal/rules"
	"github.com/recolab/backend/internal/store"
)

type ExpoHandler struct {
	gorseClient *gorse.Client
	itemStore   *store.ItemStore
	tenantID    string
}

func NewExpoHandler(gc *gorse.Client, is *store.ItemStore, tenantID string) *ExpoHandler {
	return &ExpoHandler{gorseClient: gc, itemStore: is, tenantID: tenantID}
}

// GetBuyerBoothRecommendations handles GET /api/demo/expo/buyer/:buyerId/booths
func (h *ExpoHandler) GetBuyerBoothRecommendations(c *gin.Context) {
	buyerId := c.Param("buyerId")
	limit := parseLimit(c, 8)
	refresh := c.Query("refresh")
	poolLimit := candidateLimit(limit)

	userLabels := h.getUserLabels(buyerId)

	gorseScores, _ := h.gorseClient.GetRecommendations(buyerId, "expo_booth", poolLimit)
	scores := filterExpoByLabels(gorseScores, userLabels, h.itemStore, poolLimit)

	if len(scores) < limit && len(userLabels) > 0 {
		scores = h.padWithFallback(scores, "expo_booth", userLabels, poolLimit)
	}
	if len(scores) == 0 {
		scores, _ = h.gorseClient.GetPopular("expo_booth", poolLimit)
		if len(scores) == 0 {
			scores = h.itemStore.GetFallbackExpoScores(poolLimit)
		}
	}

	scores = rules.DeduplicateItems(scores)
	scores = selectRefreshWindow(scores, refresh, limit)
	recs := h.buildBoothRecs(scores, userLabels, "expo_buyer_booths")

	c.JSON(http.StatusOK, gin.H{
		"user_id":         buyerId,
		"scenario":        "buyer_to_booth",
		"scenario_label":  "Buyer → Booth",
		"tenant_id":       h.tenantID,
		"user_interests":  userLabels,
		"count":           len(recs),
		"recommendations": recs,
	})
}

// GetBuyerSupplierRecommendations handles GET /api/demo/expo/buyer/:buyerId/suppliers
func (h *ExpoHandler) GetBuyerSupplierRecommendations(c *gin.Context) {
	buyerId := c.Param("buyerId")
	limit := parseLimit(c, 8)
	refresh := c.Query("refresh")
	poolLimit := candidateLimit(limit)

	userLabels := h.getUserLabels(buyerId)

	// Suppliers are booths with manufacturing/b2b/export labels
	supplierLabels := append(userLabels, "manufacturing", "b2b", "export", "supplier")
	gorseScores, _ := h.gorseClient.GetRecommendations(buyerId, "expo_booth", poolLimit)
	scores := filterExpoByLabels(gorseScores, supplierLabels, h.itemStore, poolLimit)

	if len(scores) < limit {
		scores = h.padWithFallback(scores, "expo_booth", supplierLabels, poolLimit)
	}
	if len(scores) == 0 {
		scores = h.itemStore.GetFallbackExpoScores(poolLimit)
	}

	scores = rules.DeduplicateItems(scores)
	scores = selectRefreshWindow(scores, refresh, limit)
	recs := h.buildBoothRecs(scores, userLabels, "expo_buyer_suppliers")

	c.JSON(http.StatusOK, gin.H{
		"user_id":         buyerId,
		"scenario":        "buyer_to_supplier",
		"scenario_label":  "Buyer → Supplier",
		"tenant_id":       h.tenantID,
		"user_interests":  userLabels,
		"count":           len(recs),
		"recommendations": recs,
	})
}

// GetSellerBuyers handles GET /api/demo/expo/seller/:sellerId/buyers
func (h *ExpoHandler) GetSellerBuyers(c *gin.Context) {
	sellerId := c.Param("sellerId")
	limit := parseLimit(c, 8)
	refresh := c.Query("refresh")
	poolLimit := candidateLimit(limit)

	// Get booth labels to match buyers
	boothMeta := h.itemStore.GetItem(sellerId)
	var boothLabels []string
	var boothName string
	if boothMeta != nil {
		boothLabels = boothMeta.Labels
		boothName = boothMeta.Comment
	}

	neighbors, err := h.gorseClient.GetSimilarUsers(sellerId, poolLimit)
	// Filter out neighbors with no labels (not useful for demo)
	var validNeighbors []gorse.Score
	for _, n := range neighbors {
		if u := h.itemStore.GetUser(n.Id); u != nil && len(u.Labels) > 0 {
			validNeighbors = append(validNeighbors, n)
		}
	}
	if err != nil || len(validNeighbors) == 0 {
		// Prioritise label-matched buyers (the 3 demo users will rank highest)
		validNeighbors = h.itemStore.GetLabelMatchedBuyers(boothLabels, poolLimit)
	}
	// Pad to limit with any remaining label-bearing users
	if len(validNeighbors) < limit {
		extra := h.itemStore.GetLabelMatchedBuyers(nil, poolLimit) // all labeled users
		seen := make(map[string]bool, len(validNeighbors))
		for _, n := range validNeighbors {
			seen[n.Id] = true
		}
		for _, n := range extra {
			if !seen[n.Id] {
				validNeighbors = append(validNeighbors, n)
				seen[n.Id] = true
			}
			if len(validNeighbors) >= limit {
				break
			}
		}
	}

	validNeighbors = selectRefreshWindow(validNeighbors, refresh, limit)

	buyers := make([]gin.H, 0, len(validNeighbors))
	for _, n := range validNeighbors {
		u := h.itemStore.GetUser(n.Id)
		var labels []string
		var name string
		if u != nil {
			labels = u.Labels
			name = u.Comment
			if name == "" {
				name = n.Id
			}
		} else {
			name = n.Id
		}
		matchReason := buildMatchReason(boothLabels, labels, boothName)
		buyers = append(buyers, gin.H{
			"user_id": n.Id,
			"name":    name,
			"score":   n.Score,
			"reason":  matchReason,
			"labels":  labels,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"seller_id":      sellerId,
		"seller_name":    boothName,
		"seller_labels":  boothLabels,
		"scenario":       "seller_to_buyer",
		"scenario_label": "Seller → Buyer",
		"tenant_id":      h.tenantID,
		"count":          len(buyers),
		"buyers":         buyers,
	})
}

// GetTrendingBooths handles GET /api/demo/expo/trending
func (h *ExpoHandler) GetTrendingBooths(c *gin.Context) {
	limit := parseLimit(c, 10)
	refresh := c.Query("refresh")
	poolLimit := candidateLimit(limit)

	scores, err := h.gorseClient.GetPopular("expo_booth", poolLimit)
	if err != nil || len(scores) == 0 {
		scores = h.itemStore.GetFallbackExpoScores(poolLimit)
	}
	scores = rules.DeduplicateItems(scores)
	scores = selectRefreshWindow(scores, refresh, limit)

	existing := make(map[string]bool)
	recs := make([]gin.H, 0, limit)
	for _, s := range scores {
		meta := h.itemStore.GetItem(s.Id)
		if meta == nil || existing[s.Id] {
			continue
		}
		existing[s.Id] = true
		recs = append(recs, gin.H{
			"item_id":  s.Id,
			"title":    meta.Comment,
			"score":    s.Score,
			"reason":   "High visitor interest this session",
			"labels":   meta.Labels,
			"category": "expo_booth",
		})
	}
	// Pad with fallback if Gorse returned stale IDs
	if len(recs) < limit {
		for _, s := range h.itemStore.GetFallbackExpoScores(poolLimit) {
			if existing[s.Id] {
				continue
			}
			meta := h.itemStore.GetItem(s.Id)
			if meta == nil {
				continue
			}
			existing[s.Id] = true
			recs = append(recs, gin.H{
				"item_id":  s.Id,
				"title":    meta.Comment,
				"score":    s.Score,
				"reason":   "High visitor interest this session",
				"labels":   meta.Labels,
				"category": "expo_booth",
			})
			if len(recs) >= limit {
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"scenario":        "trending_booths",
		"scenario_label":  "Trending Booths",
		"tenant_id":       h.tenantID,
		"count":           len(recs),
		"recommendations": recs,
	})
}

// GetRecommendedSessions handles GET /api/demo/expo/buyer/:buyerId/sessions
func (h *ExpoHandler) GetRecommendedSessions(c *gin.Context) {
	buyerId := c.Param("buyerId")
	limit := parseLimit(c, 8)
	refresh := c.Query("refresh")
	poolLimit := candidateLimit(limit)

	userLabels := h.getUserLabels(buyerId)

	gorseScores, _ := h.gorseClient.GetRecommendations(buyerId, "session", poolLimit)
	scores := filterExpoByLabels(gorseScores, userLabels, h.itemStore, poolLimit)

	if len(scores) < limit && len(userLabels) > 0 {
		scores = h.padWithFallback(scores, "session", userLabels, poolLimit)
	}
	if len(scores) == 0 {
		scores = h.itemStore.GetStratifiedFallback("session", userLabels, poolLimit)
	}

	scores = rules.DeduplicateItems(scores)
	scores = selectRefreshWindow(scores, refresh, limit)

	recs := make([]gin.H, 0, len(scores))
	for _, s := range scores {
		meta := h.itemStore.GetItem(s.Id)
		if meta == nil {
			continue
		}
		reason := buildSessionReason(userLabels, meta.Labels)
		recs = append(recs, gin.H{
			"item_id":  s.Id,
			"title":    meta.Comment,
			"score":    s.Score,
			"reason":   reason,
			"labels":   meta.Labels,
			"category": "session",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":         buyerId,
		"scenario":        "recommended_sessions",
		"scenario_label":  "Recommended Sessions",
		"tenant_id":       h.tenantID,
		"user_interests":  userLabels,
		"count":           len(recs),
		"recommendations": recs,
	})
}

// --- helpers ---

func parseLimit(c *gin.Context, def int) int {
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			return n
		}
	}
	return def
}

func (h *ExpoHandler) getUserLabels(userId string) []string {
	if u := h.itemStore.GetUser(userId); u != nil && len(u.Labels) > 0 {
		return u.Labels
	}
	if gu, _ := h.gorseClient.GetUser(userId); gu != nil && len(gu.Labels) > 0 {
		return gu.Labels
	}
	return nil
}

func (h *ExpoHandler) padWithFallback(existing []gorse.Score, category string, userLabels []string, limit int) []gorse.Score {
	existing = h.itemStore.GetCrossScenarioRecs(category, userLabels, existing, limit)
	if len(existing) >= limit {
		return existing[:limit]
	}
	seen := make(map[string]bool, len(existing))
	for _, s := range existing {
		seen[s.Id] = true
	}
	for _, s := range h.itemStore.GetStratifiedFallback(category, userLabels, limit) {
		if !seen[s.Id] {
			existing = append(existing, s)
			seen[s.Id] = true
		}
		if len(existing) >= limit {
			break
		}
	}
	return existing
}

func (h *ExpoHandler) buildBoothRecs(scores []gorse.Score, userLabels []string, scenario string) []gin.H {
	recs := make([]gin.H, 0, len(scores))
	for _, s := range scores {
		meta := h.itemStore.GetItem(s.Id)
		if meta == nil {
			continue
		}
		reason := buildBoothReason(userLabels, meta.Labels)
		recs = append(recs, gin.H{
			"item_id":  s.Id,
			"title":    meta.Comment,
			"score":    s.Score,
			"reason":   reason,
			"labels":   meta.Labels,
			"category": "expo_booth",
		})
	}
	return recs
}

func filterExpoByLabels(scores []gorse.Score, userLabels []string, is *store.ItemStore, limit int) []gorse.Score {
	if len(userLabels) == 0 {
		return scores
	}
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

func buildBoothReason(userLabels, itemLabels []string) string {
	matched := intersectLabels(userLabels, itemLabels)
	if len(matched) == 0 {
		return "Relevant to your industry"
	}
	shown := matched
	if len(shown) > 3 {
		shown = shown[:3]
	}
	return "Matches: " + strings.Join(capitalize(shown), ", ")
}

func buildSessionReason(userLabels, itemLabels []string) string {
	matched := intersectLabels(userLabels, itemLabels)
	if len(matched) == 0 {
		return "Recommended for expo attendees"
	}
	shown := matched
	if len(shown) > 3 {
		shown = shown[:3]
	}
	return "Relevant to: " + strings.Join(capitalize(shown), ", ")
}

func buildMatchReason(boothLabels, buyerLabels []string, boothName string) string {
	matched := intersectLabels(boothLabels, buyerLabels)
	if len(matched) == 0 {
		return "Potential buyer for " + boothName
	}
	shown := matched
	if len(shown) > 3 {
		shown = shown[:3]
	}
	return "Interest overlap: " + strings.Join(capitalize(shown), ", ")
}

func intersectLabels(a, b []string) []string {
	set := make(map[string]bool, len(a))
	for _, v := range a {
		set[strings.ToLower(v)] = true
	}
	var result []string
	for _, v := range b {
		if set[strings.ToLower(v)] {
			result = append(result, v)
		}
	}
	return result
}

func capitalize(ss []string) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		if len(s) > 0 {
			out[i] = strings.ToUpper(s[:1]) + s[1:]
		}
	}
	return out
}

// boothScoreByLabels ranks booths by label overlap count (used for test/debug).
func boothScoreByLabels(items []gorse.Item, userLabels []string) []gorse.Score {
	type scored struct {
		id    string
		score float64
	}
	labelSet := make(map[string]bool, len(userLabels))
	for _, l := range userLabels {
		labelSet[strings.ToLower(l)] = true
	}
	var candidates []scored
	for _, item := range items {
		overlap := 0
		for _, l := range item.Labels {
			if labelSet[strings.ToLower(l)] {
				overlap++
			}
		}
		if overlap > 0 {
			candidates = append(candidates, scored{id: item.ItemId, score: 0.75 + float64(overlap)*0.05})
		}
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].score > candidates[j].score })
	out := make([]gorse.Score, len(candidates))
	for i, c := range candidates {
		out[i] = gorse.Score{Id: c.id, Score: c.score}
	}
	return out
}
