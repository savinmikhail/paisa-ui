package server

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func importHealthHandler(path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		content, err := os.ReadFile(path)
		var report struct {
			GeneratedAt string            `json:"generated_at"`
			Services    []json.RawMessage `json:"services"`
			Statements  []json.RawMessage `json:"statements"`
			Accounts    []json.RawMessage `json:"accounts"`
			Issues      []struct {
				ID string `json:"id"`
			} `json:"issues"`
		}
		if err != nil || json.Unmarshal(content, &report) != nil || report.GeneratedAt == "" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Import health report unavailable"})
			return
		}
		c.JSON(http.StatusOK, report)
	}
}
