package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CompleteUserOnboarding marks the authenticated user's onboarding as complete.
func (h *Handler) CompleteUserOnboarding(c *gin.Context) {
	userID, ok := h.deriveUserIDQuery(c)
	if !ok {
		return
	}

	const q = `
		UPDATE users 
		SET is_onboarded = true, 
		    updated_at = CURRENT_TIMESTAMP AT TIME ZONE 'UTC' 
		WHERE id = $1`

	tag, err := h.repo.GetDB().Exec(c.Request.Context(), q, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("failed to complete onboarding status flip: %w", err).Error()})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "authenticated user was not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "onboarding_finalized"})
}
