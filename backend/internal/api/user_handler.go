package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CompleteUserOnboarding permanently flips a user's onboarding status boolean flag to true.
func (h *Handler) CompleteUserOnboarding(c *gin.Context) {
	var req struct {
		UserID int64 `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is a mandatory parameter"})
		return
	}

	// Flip the status flag and update the record's tracking timeline
	const q = `
		UPDATE users 
		SET is_onboarded = true, 
		    updated_at = CURRENT_TIMESTAMP AT TIME ZONE 'UTC' 
		WHERE id = $1`

	_, err := h.repo.GetDB().Exec(c.Request.Context(), q, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("failed to complete onboarding status flip: %w", err).Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "onboarding_finalized"})
}
