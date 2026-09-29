package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/recolab/backend/internal/gorse"
	"github.com/recolab/backend/internal/store"
)

// OnboardHandler handles new-user cold-start via interest picking (Solution 1).
type OnboardHandler struct {
	gorseClient *gorse.Client
	itemStore   *store.ItemStore
	tenantID    string
}

func NewOnboardHandler(gc *gorse.Client, is *store.ItemStore, tenantID string) *OnboardHandler {
	return &OnboardHandler{gorseClient: gc, itemStore: is, tenantID: tenantID}
}

// InterestOption is one selectable interest tag returned to the UI.
type InterestOption struct {
	Label    string `json:"label"`
	Category string `json:"category"` // which scenario it applies to
	Emoji    string `json:"emoji"`
}

// availableInterests is the full catalogue shown in the onboarding picker.
var availableInterests = []InterestOption{
	// News interests
	{Label: "sports", Category: "news", Emoji: "⚽"},
	{Label: "finance", Category: "news", Emoji: "📈"},
	{Label: "technology", Category: "news", Emoji: "💻"},
	{Label: "health", Category: "news", Emoji: "🏥"},
	{Label: "politics", Category: "news", Emoji: "🏛️"},
	{Label: "entertainment", Category: "news", Emoji: "🎬"},
	{Label: "travel", Category: "news", Emoji: "✈️"},
	{Label: "autos", Category: "news", Emoji: "🚗"},
	{Label: "music", Category: "news", Emoji: "🎵"},
	// Product interests
	{Label: "electronics", Category: "products", Emoji: "📱"},
	{Label: "fashion", Category: "products", Emoji: "👗"},
	{Label: "home", Category: "products", Emoji: "🏠"},
	{Label: "sports", Category: "products", Emoji: "🏃"},
	{Label: "beauty", Category: "products", Emoji: "💄"},
	{Label: "computers", Category: "products", Emoji: "🖥️"},
	// Expo interests
	{Label: "solar", Category: "expo", Emoji: "☀️"},
	{Label: "textile", Category: "expo", Emoji: "🧵"},
	{Label: "chemicals", Category: "expo", Emoji: "⚗️"},
	{Label: "food", Category: "expo", Emoji: "🍽️"},
	{Label: "ai", Category: "expo", Emoji: "🤖"},
	{Label: "trade", Category: "expo", Emoji: "📦"},
	{Label: "manufacturing", Category: "expo", Emoji: "🏭"},
}

// GetInterestOptions handles GET /api/onboard/interests
// Returns the full list of interest tags the user can pick from.
func (h *OnboardHandler) GetInterestOptions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"interests":   availableInterests,
		"min_pick":    3,
		"max_pick":    8,
		"tenant_id":   h.tenantID,
	})
}

type OnboardRequest struct {
	UserId    string   `json:"userId"    binding:"required"`
	Interests []string `json:"interests" binding:"required,min=1"`
}

// PostOnboard handles POST /api/onboard
// Creates/updates the user in Gorse with chosen labels, then immediately
// returns personalised recommendations across all scenarios — no training needed.
func (h *OnboardHandler) PostOnboard(c *gin.Context) {
	var req OnboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Normalise labels to lowercase
	labels := make([]string, 0, len(req.Interests))
	for _, l := range req.Interests {
		l = strings.TrimSpace(strings.ToLower(l))
		if l != "" {
			labels = append(labels, l)
		}
	}
	if len(labels) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one interest required"})
		return
	}

	// 1. Upsert user in Gorse with the picked labels
	gorseUser := gorse.User{
		UserId:  req.UserId,
		Labels:  labels,
		Comment: "onboarded_" + time.Now().UTC().Format("20060102"),
	}
	if err := h.gorseClient.InsertUsers([]gorse.User{gorseUser}); err != nil {
		// Non-fatal — still serve recommendations from local store
		c.Header("X-Gorse-Warning", "user upsert failed: "+err.Error())
	}

	// 2. Update local ItemStore so recommendations are immediately label-aware
	h.itemStore.UpsertUser(store.UserMeta{
		UserId:  req.UserId,
		Labels:  labels,
		Comment: gorseUser.Comment,
	})

	// 3. Build immediate recommendations across all scenarios using label fallback
	//    (Solution 1 + 2: label-matched, stratified by category)
	newsRecs   := h.itemStore.GetStratifiedFallback("news", labels, 10)
	productRecs := h.itemStore.GetStratifiedFallback("products", labels, 10)
	expoRecs   := h.itemStore.GetStratifiedFallback("expo_booth", labels, 8)

	toItems := func(scores []gorse.Score) []gin.H {
		out := make([]gin.H, 0, len(scores))
		for _, s := range scores {
			meta := h.itemStore.GetItem(s.Id)
			if meta == nil {
				continue
			}
			out = append(out, gin.H{
				"item_id":  s.Id,
				"title":    meta.Comment,
				"score":    s.Score,
				"labels":   meta.Labels,
				"category": func() string {
					for _, cat := range meta.Categories {
						return cat
					}
					return ""
				}(),
				"reason": buildOnboardReason(labels, meta.Labels),
			})
		}
		return out
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":          req.UserId,
		"interests":        labels,
		"tenant_id":        h.tenantID,
		"cold_start":       true,
		"message":          "Interests saved. Showing label-matched recommendations while your personal model trains.",
		"news":             toItems(newsRecs),
		"products":         toItems(productRecs),
		"expo":             toItems(expoRecs),
	})
}

func buildOnboardReason(userLabels, itemLabels []string) string {
	set := make(map[string]bool, len(userLabels))
	for _, l := range userLabels {
		set[l] = true
	}
	var matched []string
	for _, l := range itemLabels {
		if set[strings.ToLower(l)] {
			matched = append(matched, l)
		}
	}
	if len(matched) == 0 {
		return "Recommended based on your interests"
	}
	if len(matched) > 3 {
		matched = matched[:3]
	}
	return "Matches your interests: " + strings.Join(matched, ", ")
}
