package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/recolab/backend/internal/db"
	"github.com/recolab/backend/internal/gorse"
	"github.com/recolab/backend/internal/store"
)

type DataHandler struct {
	gorseClient *gorse.Client
	itemStore   *store.ItemStore
	pgStore     *db.Store // may be nil
	dataDir     string
	tenantID    string
}

func NewDataHandler(gc *gorse.Client, is *store.ItemStore, pg *db.Store, dataDir string, tenantID string) *DataHandler {
	return &DataHandler{gorseClient: gc, itemStore: is, pgStore: pg, dataDir: dataDir, tenantID: tenantID}
}

// LoadData handles POST /api/demo/load-data
func (h *DataHandler) LoadData(c *gin.Context) {
	results := gin.H{}

	// --- Users ---
	usersPath := filepath.Join(h.dataDir, "users.json")
	usersData, err := os.ReadFile(usersPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot read users.json: " + err.Error()})
		return
	}
	var users []gorse.User
	if err := json.Unmarshal(usersData, &users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot parse users.json: " + err.Error()})
		return
	}
	const userBatch = 1000
	usersLoaded := 0
	var usersErr string
	for i := 0; i < len(users); i += userBatch {
		end := i + userBatch
		if end > len(users) {
			end = len(users)
		}
		if err := h.gorseClient.InsertUsers(users[i:end]); err != nil {
			usersErr = err.Error()
			break
		}
		usersLoaded += end - i
	}
	if usersErr != "" {
		results["users_error"] = usersErr
	}
	results["users_loaded"] = usersLoaded

	// --- Items (all categories, including sessions.json) ---
	itemFiles := []string{"news_items.json", "product_items.json", "expo_booths.json", "sessions.json"}
	totalItems := 0
	const itemBatch = 1000
	for _, f := range itemFiles {
		path := filepath.Join(h.dataDir, f)
		data, err := os.ReadFile(path)
		if err != nil {
			results[f+"_error"] = "cannot read: " + err.Error()
			continue
		}
		var items []gorse.Item
		if err := json.Unmarshal(data, &items); err != nil {
			results[f+"_error"] = "cannot parse: " + err.Error()
			continue
		}
		fileLoaded := 0
		var fileErr string
		for i := 0; i < len(items); i += itemBatch {
			end := i + itemBatch
			if end > len(items) {
				end = len(items)
			}
			if err := h.gorseClient.InsertItems(items[i:end]); err != nil {
				fileErr = err.Error()
				break
			}
			fileLoaded += end - i
		}
		if fileErr != "" {
			results[f+"_error"] = fileErr
		}
		results[f+"_loaded"] = fileLoaded
		totalItems += fileLoaded
	}
	results["total_items_loaded"] = totalItems

	// --- Feedback (PUT overwrites duplicates safely) ---
	feedbackPath := filepath.Join(h.dataDir, "feedback.json")
	feedbackData, err := os.ReadFile(feedbackPath)
	if err != nil {
		results["feedback_error"] = "cannot read: " + err.Error()
	} else {
		var feedback []gorse.Feedback
		if err := json.Unmarshal(feedbackData, &feedback); err != nil {
			results["feedback_error"] = "cannot parse: " + err.Error()
		} else {
			const batchSize = 2000
			loaded := 0
			for i := 0; i < len(feedback); i += batchSize {
				end := i + batchSize
				if end > len(feedback) {
					end = len(feedback)
				}
				if err := h.gorseClient.InsertFeedback(feedback[i:end]); err != nil {
					results["feedback_batch_error"] = err.Error()
					break
				}
				loaded += end - i
			}
			results["feedback_loaded"] = loaded
		}
	}

	// Refresh in-memory store so dashboard counts are immediately correct
	_ = h.itemStore.LoadFromDataDir(h.dataDir)

	results["status"] = "complete"
	results["tenant_id"] = h.tenantID
	c.JSON(http.StatusOK, results)
}

// GetDashboard handles GET /api/demo/dashboard
// totalFeedback is read from Postgres when available, otherwise from the in-memory store.
func (h *DataHandler) GetDashboard(c *gin.Context) {
	totalUsers := h.itemStore.TotalUsers()
	totalItems := h.itemStore.TotalItems()

	// Prefer Postgres count (includes live events), fall back to seeded count
	var totalFeedback int64
	if h.pgStore != nil {
		if n, err := h.pgStore.FeedbackCount(h.tenantID); err == nil {
			totalFeedback = n
		}
	}
	if totalFeedback == 0 {
		totalFeedback = int64(h.itemStore.TotalFeedback())
	}

	// Recent events from Postgres (empty slice when DB unavailable)
	recentEvents := []gin.H{}
	if h.pgStore != nil {
		if events, err := h.pgStore.RecentEvents(h.tenantID, 10); err == nil {
			for _, e := range events {
				recentEvents = append(recentEvents, gin.H{
					"feedback_type": e.FeedbackType,
					"user_id":       e.UserID,
					"item_id":       e.ItemID,
					"timestamp":     e.Timestamp,
				})
			}
		}
	}

	// Category breakdown
	categories := map[string]int{
		"news":       0,
		"products":   0,
		"expo_booth": 0,
		"session":    0,
	}
	for _, item := range h.itemStore.AllItems() {
		for _, cat := range item.Categories {
			if _, ok := categories[cat]; ok {
				categories[cat]++
			}
		}
	}

	topCategories := []gin.H{
		{"name": "News Articles", "category": "news", "count": categories["news"]},
		{"name": "Products", "category": "products", "count": categories["products"]},
		{"name": "Expo Booths", "category": "expo_booth", "count": categories["expo_booth"]},
		{"name": "Sessions", "category": "session", "count": categories["session"]},
	}

	c.JSON(http.StatusOK, gin.H{
		"tenant_id":      h.tenantID,
		"totalUsers":     totalUsers,
		"totalItems":     totalItems,
		"totalFeedback":  totalFeedback,
		"topCategories":  topCategories,
		"recentEvents":   recentEvents,
	})
}
