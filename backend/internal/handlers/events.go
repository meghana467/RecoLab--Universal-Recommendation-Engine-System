package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/recolab/backend/internal/gorse"
)

type EventHandler struct {
	gorseClient *gorse.Client
	tenantID    string
}

func NewEventHandler(gc *gorse.Client, tenantID string) *EventHandler {
	return &EventHandler{gorseClient: gc, tenantID: tenantID}
}

type EventRequest struct {
	FeedbackType string `json:"feedbackType" binding:"required"`
	UserId       string `json:"userId" binding:"required"`
	ItemId       string `json:"itemId" binding:"required"`
}

// PostEvent handles POST /api/events
func (h *EventHandler) PostEvent(c *gin.Context) {
	var req EventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	feedback := []gorse.Feedback{
		{
			FeedbackType: req.FeedbackType,
			UserId:       req.UserId,
			ItemId:       req.ItemId,
			Timestamp:    time.Now().UTC().Format(time.RFC3339),
		},
	}

	if err := h.gorseClient.InsertFeedback(feedback); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record event: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "Event recorded",
		"feedbackType":  req.FeedbackType,
		"userId":        req.UserId,
		"itemId":        req.ItemId,
		"timestamp":     feedback[0].Timestamp,
		"tenant_id":     h.tenantID,
	})
}
