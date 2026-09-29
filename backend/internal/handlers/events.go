package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/recolab/backend/internal/db"
	"github.com/recolab/backend/internal/gorse"
)

type EventHandler struct {
	gorseClient *gorse.Client
	pgStore     *db.Store // may be nil if Postgres is unavailable
	tenantID    string
}

func NewEventHandler(gc *gorse.Client, pg *db.Store, tenantID string) *EventHandler {
	return &EventHandler{gorseClient: gc, pgStore: pg, tenantID: tenantID}
}

type EventRequest struct {
	FeedbackType string `json:"feedbackType" binding:"required"`
	UserId       string `json:"userId"       binding:"required"`
	ItemId       string `json:"itemId"       binding:"required"`
}

// PostEvent handles POST /api/events
// It writes to both Gorse (for real-time recommendations) and Postgres (for durable storage).
func (h *EventHandler) PostEvent(c *gin.Context) {
	var req EventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	now := time.Now().UTC()
	ts := now.Format(time.RFC3339)

	// 1. Send to Gorse for real-time recommendation updates
	feedback := []gorse.Feedback{
		{
			FeedbackType: req.FeedbackType,
			UserId:       req.UserId,
			ItemId:       req.ItemId,
			Timestamp:    ts,
		},
	}
	if err := h.gorseClient.InsertFeedback(feedback); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record event in Gorse: " + err.Error()})
		return
	}

	// 2. Persist to PostgreSQL (best-effort — do not fail the request if DB is down)
	if h.pgStore != nil {
		if err := h.pgStore.InsertEvent(req.FeedbackType, req.UserId, req.ItemId, h.tenantID, now); err != nil {
			log.Printf("WARNING: failed to persist event to Postgres: %v", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      "Event recorded",
		"feedbackType": req.FeedbackType,
		"userId":       req.UserId,
		"itemId":       req.ItemId,
		"timestamp":    ts,
		"tenant_id":    h.tenantID,
	})
}
