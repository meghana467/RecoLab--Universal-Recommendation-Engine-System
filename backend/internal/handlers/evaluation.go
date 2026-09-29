package handlers

import (
	"encoding/json"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/recolab/backend/internal/gorse"
	"github.com/recolab/backend/internal/rules"
	"github.com/recolab/backend/internal/store"
)

type EvaluationHandler struct {
	gorseClient *gorse.Client
	itemStore   *store.ItemStore
	dataDir     string
	tenantID    string
}

func NewEvaluationHandler(gc *gorse.Client, is *store.ItemStore, dataDir string, tenantID string) *EvaluationHandler {
	return &EvaluationHandler{gorseClient: gc, itemStore: is, dataDir: dataDir, tenantID: tenantID}
}

type strategyMetrics struct {
	Strategy  string  `json:"strategy"`
	Users     int     `json:"users"`
	Precision float64 `json:"precision_at_k"`
	Recall    float64 `json:"recall_at_k"`
	HitRate   float64 `json:"hit_rate_at_k"`
	MRR       float64 `json:"mrr_at_k"`
	NDCG      float64 `json:"ndcg_at_k"`
}

type heldOutCase struct {
	UserId   string
	ItemIds  map[string]bool
	UserRank int
}

// GetEvaluation handles GET /api/demo/evaluation?category=news&k=10&users=25&mode=personalized.
func (h *EvaluationHandler) GetEvaluation(c *gin.Context) {
	category := c.DefaultQuery("category", "news")
	mode := c.DefaultQuery("mode", "personalized")
	k := parsePositiveInt(c.Query("k"), 10)
	maxUsers := parsePositiveInt(c.Query("users"), 25)

	metrics, best, cases, err := h.evaluateStrategies(category, mode, k, maxUsers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tenant_id":       h.tenantID,
		"category":        category,
		"mode":            mode,
		"k":               k,
		"evaluated_users": len(cases),
		"holdout":         "latest positive item per user from feedback.json",
		"note":            "Demo offline evaluation. For strict evaluation, seed Gorse with train data only and keep holdout feedback out of training.",
		"best_strategy":   best,
		"metrics":         metrics,
	})
}

// GetEvaluatedNews handles GET /api/demo/evaluated-news/:userId.
func (h *EvaluationHandler) GetEvaluatedNews(c *gin.Context) {
	userId := c.Param("userId")
	limit := parsePositiveInt(c.Query("limit"), 10)
	maxUsers := parsePositiveInt(c.Query("users"), 25)

	metrics, best, _, err := h.evaluateStrategies("news", "personalized", limit, maxUsers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userLabels := h.getUserLabels(userId)
	scores := h.recommendByStrategy(userId, "news", best.Strategy, userLabels, limit)
	if len(scores) < limit {
		scores = h.padNewsScores(scores, userLabels, limit)
	}
	scores = rules.DeduplicateItems(scores)

	recs := make([]gin.H, 0, len(scores))
	for _, s := range scores {
		meta := h.itemStore.GetItem(s.Id)
		if meta == nil {
			continue
		}
		recs = append(recs, newsRec(s, meta, buildNewsReason(userLabels, meta.Labels)))
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":            userId,
		"scenario":           "evaluated_personalized_news",
		"tenant_id":          h.tenantID,
		"user_interests":     userLabels,
		"selected_strategy":  best.Strategy,
		"selection_metric":   "highest ndcg_at_k, then recall_at_k, then precision_at_k",
		"evaluation":         metrics,
		"count":              len(recs),
		"recommendations":    recs,
	})
}

func (h *EvaluationHandler) evaluateStrategies(category string, mode string, k int, maxUsers int) ([]strategyMetrics, strategyMetrics, []heldOutCase, error) {
	cases, err := h.loadHoldoutCases(category, maxUsers)
	if err != nil {
		return nil, strategyMetrics{}, nil, err
	}

	strategies := strategiesForMode(mode)
	metrics := make([]strategyMetrics, 0, len(strategies))
	for _, strategy := range strategies {
		var totalPrecision, totalRecall, totalHitRate, totalMRR, totalNDCG float64
		evaluated := 0
		for _, tc := range cases {
			labels := h.getUserLabels(tc.UserId)
			scores := h.recommendByStrategy(tc.UserId, category, strategy, labels, k)
			if len(scores) == 0 {
				continue
			}
			pm := rankingMetrics(scores, tc.ItemIds, k)
			totalPrecision += pm.Precision
			totalRecall += pm.Recall
			totalHitRate += pm.HitRate
			totalMRR += pm.MRR
			totalNDCG += pm.NDCG
			evaluated++
		}
		if evaluated == 0 {
			metrics = append(metrics, strategyMetrics{Strategy: strategy})
			continue
		}
		metrics = append(metrics, strategyMetrics{
			Strategy:  strategy,
			Users:     evaluated,
			Precision: round4(totalPrecision / float64(evaluated)),
			Recall:    round4(totalRecall / float64(evaluated)),
			HitRate:   round4(totalHitRate / float64(evaluated)),
			MRR:       round4(totalMRR / float64(evaluated)),
			NDCG:      round4(totalNDCG / float64(evaluated)),
		})
	}

	best := metrics[0]
	for _, m := range metrics[1:] {
		if betterMetrics(m, best) {
			best = m
		}
	}
	return metrics, best, cases, nil
}

func (h *EvaluationHandler) loadHoldoutCases(category string, maxUsers int) ([]heldOutCase, error) {
	path := filepath.Join(h.dataDir, "feedback.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var feedback []gorse.Feedback
	if err := json.Unmarshal(data, &feedback); err != nil {
		return nil, err
	}

	byUser := make(map[string][]gorse.Feedback)
	for _, f := range feedback {
		// Only treat genuinely positive feedback types as positive — not view/read
		if !isStrictlyPositiveFeedback(f.FeedbackType) {
			continue
		}
		meta := h.itemStore.GetItem(f.ItemId)
		if meta == nil || !hasCategory(meta.Categories, category) {
			continue
		}
		byUser[f.UserId] = append(byUser[f.UserId], f)
	}

	users := make([]string, 0, len(byUser))
	for userId := range byUser {
		users = append(users, userId)
	}
	sort.Strings(users)

	cases := make([]heldOutCase, 0, maxUsers)
	for _, userId := range users {
		items := byUser[userId]
		if len(items) < 2 {
			continue
		}
		sort.Slice(items, func(i, j int) bool {
			return feedbackTime(items[i]).Before(feedbackTime(items[j]))
		})
		last := items[len(items)-1]
		cases = append(cases, heldOutCase{
			UserId:   userId,
			ItemIds:  map[string]bool{last.ItemId: true},
			UserRank: len(cases) + 1,
		})
		if len(cases) >= maxUsers {
			break
		}
	}
	return cases, nil
}

func (h *EvaluationHandler) recommendByStrategy(userId string, category string, strategy string, userLabels []string, limit int) []gorse.Score {
	switch strategy {
	case "label_fallback":
		return h.localPersonalizedFallback(category, userLabels, limit)
	case "gorse_popular":
		scores, _ := h.gorseClient.GetPopular(category, limit)
		return scores
	case "local_popular":
		return h.localPopularFallback(category, limit)
	default: // "gorse_personalized"
		scores, _ := h.gorseClient.GetRecommendations(userId, category, limit*3)
		if category == "news" && len(userLabels) > 0 {
			return filterByLabels(scores, userLabels, h.itemStore, limit)
		}
		if category == "products" && len(userLabels) > 0 {
			return filterProductsByLabels(scores, userLabels, h.itemStore, limit)
		}
		if (category == "expo_booth" || category == "session") && len(userLabels) > 0 {
			return filterExpoByLabels(scores, userLabels, h.itemStore, limit)
		}
		if len(scores) > limit {
			return scores[:limit]
		}
		return scores
	}
}

func (h *EvaluationHandler) localPersonalizedFallback(category string, userLabels []string, limit int) []gorse.Score {
	scores := h.itemStore.GetPersonalizedFallback(category, userLabels, limit)
	if len(scores) > 0 {
		return scores
	}
	return h.localPopularFallback(category, limit)
}

func (h *EvaluationHandler) localPopularFallback(category string, limit int) []gorse.Score {
	switch category {
	case "news":
		return h.itemStore.GetFallbackNewsScores(limit)
	case "products":
		return h.itemStore.GetFallbackProductScores(limit)
	case "expo_booth":
		return h.itemStore.GetFallbackExpoScores(limit)
	default:
		return h.itemStore.GetPersonalizedFallback(category, nil, limit)
	}
}

func (h *EvaluationHandler) padNewsScores(existing []gorse.Score, userLabels []string, limit int) []gorse.Score {
	seen := make(map[string]bool, len(existing))
	for _, s := range existing {
		seen[s.Id] = true
	}
	for _, s := range h.itemStore.GetPersonalizedFallback("news", userLabels, limit) {
		if !seen[s.Id] {
			existing = append(existing, s)
			seen[s.Id] = true
		}
		if len(existing) >= limit {
			return existing
		}
	}
	for _, s := range h.itemStore.GetFallbackNewsScores(limit) {
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

func (h *EvaluationHandler) getUserLabels(userId string) []string {
	if localUser := h.itemStore.GetUser(userId); localUser != nil {
		return localUser.Labels
	}
	if gorseUser, _ := h.gorseClient.GetUser(userId); gorseUser != nil {
		return gorseUser.Labels
	}
	return nil
}

type pointMetrics struct {
	Precision float64
	Recall    float64
	HitRate   float64
	MRR       float64
	NDCG      float64
}

func rankingMetrics(scores []gorse.Score, relevant map[string]bool, k int) pointMetrics {
	if k <= 0 || len(relevant) == 0 {
		return pointMetrics{}
	}
	if len(scores) < k {
		k = len(scores)
	}

	hits := 0
	var dcg float64
	var mrr float64
	for i := 0; i < k; i++ {
		if relevant[scores[i].Id] {
			hits++
			rank := i + 1
			dcg += 1 / math.Log2(float64(rank)+1)
			if mrr == 0 {
				mrr = 1 / float64(rank)
			}
		}
	}

	idealHits := len(relevant)
	if idealHits > k {
		idealHits = k
	}
	var idcg float64
	for i := 1; i <= idealHits; i++ {
		idcg += 1 / math.Log2(float64(i)+1)
	}
	ndcg := 0.0
	if idcg > 0 {
		ndcg = dcg / idcg
	}

	hitRate := 0.0
	if hits > 0 {
		hitRate = 1
	}
	return pointMetrics{
		Precision: float64(hits) / float64(k),
		Recall:    float64(hits) / float64(len(relevant)),
		HitRate:   hitRate,
		MRR:       mrr,
		NDCG:      ndcg,
	}
}

func betterMetrics(a, b strategyMetrics) bool {
	if a.NDCG != b.NDCG {
		return a.NDCG > b.NDCG
	}
	if a.Recall != b.Recall {
		return a.Recall > b.Recall
	}
	return a.Precision > b.Precision
}

func strategiesForMode(mode string) []string {
	switch mode {
	case "global":
		return []string{"gorse_popular", "local_popular"}
	default:
		return []string{"gorse_personalized", "label_fallback"}
	}
}

func defaultStrategy(mode string) string {
	switch mode {
	case "global":
		return "gorse_popular"
	default:
		return "gorse_personalized"
	}
}

func parsePositiveInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func hasCategory(categories []string, category string) bool {
	for _, c := range categories {
		if c == category {
			return true
		}
	}
	return false
}

// isStrictlyPositiveFeedback treats only genuine engagement signals as positive.
// view and read are read_feedback_types in Gorse — NOT positive — so excluding
// them here prevents inflated holdout evaluation metrics.
func isStrictlyPositiveFeedback(feedbackType string) bool {
	switch feedbackType {
	case "cart", "purchase", "click", "like", "bookmark", "share", "inquiry", "meeting_request":
		return true
	default:
		return false
	}
}

func feedbackTime(f gorse.Feedback) time.Time {
	t, err := time.Parse(time.RFC3339, f.Timestamp)
	if err != nil {
		return time.Time{}
	}
	return t
}

func round4(v float64) float64 {
	return math.Round(v*10000) / 10000
}
