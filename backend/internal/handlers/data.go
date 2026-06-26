package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/recolab/backend/internal/gorse"
	"github.com/recolab/backend/internal/store"
)

type DataHandler struct {
	gorseClient *gorse.Client
	itemStore   *store.ItemStore
	dataDir     string
	tenantID    string
}

func NewDataHandler(gc *gorse.Client, is *store.ItemStore, dataDir string, tenantID string) *DataHandler {
	return &DataHandler{gorseClient: gc, itemStore: is, dataDir: dataDir, tenantID: tenantID}
}

// LoadData handles POST /api/demo/load-data
func (h *DataHandler) LoadData(c *gin.Context) {
	results := gin.H{}

	// Load users
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
	if err := h.gorseClient.InsertUsers(users); err != nil {
		results["users_error"] = err.Error()
	} else {
		results["users_loaded"] = len(users)
	}

	// Load items (news + products + expo)
	itemFiles := []string{"news_items.json", "product_items.json", "expo_booths.json", "sessions.json"}
	totalItems := 0
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
		if err := h.gorseClient.InsertItems(items); err != nil {
			results[f+"_error"] = err.Error()
		} else {
			results[f+"_loaded"] = len(items)
			totalItems += len(items)
		}
	}
	results["total_items_loaded"] = totalItems

	// Load feedback
	feedbackPath := filepath.Join(h.dataDir, "feedback.json")
	feedbackData, err := os.ReadFile(feedbackPath)
	if err != nil {
		results["feedback_error"] = "cannot read: " + err.Error()
	} else {
		var feedback []gorse.Feedback
		if err := json.Unmarshal(feedbackData, &feedback); err != nil {
			results["feedback_error"] = "cannot parse: " + err.Error()
		} else {
			// Send in batches of 200
			batchSize := 200
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

	// Refresh in-memory store
	_ = h.itemStore.LoadFromDataDir(h.dataDir)

	results["status"] = "complete"
	results["tenant_id"] = h.tenantID
	c.JSON(http.StatusOK, results)
}

// GetDashboard handles GET /api/demo/dashboard
func (h *DataHandler) GetDashboard(c *gin.Context) {
	totalUsers := h.itemStore.TotalUsers()
	totalItems := h.itemStore.TotalItems()
	totalFeedback := h.itemStore.TotalFeedback()

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
		"tenant_id":       h.tenantID,
		"totalUsers":      totalUsers,
		"totalItems":      totalItems,
		"totalFeedback":   totalFeedback,
		"topCategories":   topCategories,
		"recentEvents":    []gin.H{},
	})
}
