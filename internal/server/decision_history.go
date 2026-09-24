package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/gin-gonic/gin"
)

type DecisionHistoryPayload struct {
	SchemaVersion string            `json:"schemaVersion"`
	Source        string            `json:"source"`
	Overview      json.RawMessage   `json:"overview,omitempty"`
	Themes        []json.RawMessage `json:"themes,omitempty"`
	Cases         []json.RawMessage `json:"cases"`
}

func decisionHistoryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		payload, err := loadDecisionHistory(config.GetDecisionHistoryPath())
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, payload)
	}
}

func loadDecisionHistory(path string) (DecisionHistoryPayload, error) {
	if strings.TrimSpace(path) == "" {
		return DecisionHistoryPayload{}, errors.New("decision history source is not configured")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return DecisionHistoryPayload{}, fmt.Errorf("read decision history: %w", err)
	}

	var payload DecisionHistoryPayload
	if err := json.Unmarshal(content, &payload); err != nil {
		return DecisionHistoryPayload{}, fmt.Errorf("decode decision history: %w", err)
	}
	if payload.SchemaVersion != "decision-history/v1" {
		return DecisionHistoryPayload{}, fmt.Errorf("unsupported decision history schema version %q", payload.SchemaVersion)
	}
	if payload.Source != "private-adapter" {
		return DecisionHistoryPayload{}, errors.New("decision history source must declare private-adapter")
	}
	if payload.Cases == nil {
		return DecisionHistoryPayload{}, errors.New("decision history cases must be an array")
	}

	return payload, nil
}
