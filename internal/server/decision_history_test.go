package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecisionHistoryEndpointReturnsConfiguredPayload(t *testing.T) {
	decisionHistoryFile := filepath.Join(t.TempDir(), "decision-history.json")
	content := []byte(`{"schemaVersion":"decision-history/v1","cases":[{"id":"private-test-case","title":"Private test case"}]}`)
	require.NoError(t, os.WriteFile(decisionHistoryFile, content, 0600))
	t.Setenv("PAISA_DECISION_HISTORY_FILE", decisionHistoryFile)

	router := gin.New()
	router.GET("/api/decision_history", decisionHistoryHandler())
	record := httptest.NewRecorder()
	router.ServeHTTP(record, httptest.NewRequest(http.MethodGet, "/api/decision_history", nil))

	require.Equal(t, http.StatusOK, record.Code)
	var payload DecisionHistoryPayload
	require.NoError(t, json.Unmarshal(record.Body.Bytes(), &payload))
	assert.Equal(t, "decision-history/v1", payload.SchemaVersion)
	require.Len(t, payload.Cases, 1)
	assert.JSONEq(t, `{"id":"private-test-case","title":"Private test case"}`, string(payload.Cases[0]))
}

func TestDecisionHistoryEndpointReturns503WhenSourceIsMissing(t *testing.T) {
	decisionHistoryFile := filepath.Join(t.TempDir(), "missing-decision-history.json")
	t.Setenv("PAISA_DECISION_HISTORY_FILE", decisionHistoryFile)

	router := gin.New()
	router.GET("/api/decision_history", decisionHistoryHandler())
	record := httptest.NewRecorder()
	router.ServeHTTP(record, httptest.NewRequest(http.MethodGet, "/api/decision_history?source=synthetic", nil))

	assert.Equal(t, http.StatusServiceUnavailable, record.Code)
	assert.Contains(t, record.Body.String(), "read decision history")
}

func TestDecisionHistoryEndpointReturns503WhenSourceIsInvalid(t *testing.T) {
	decisionHistoryFile := filepath.Join(t.TempDir(), "decision-history.json")
	require.NoError(t, os.WriteFile(decisionHistoryFile, []byte(`{"schemaVersion":"decision-history/v0","cases":[]}`), 0600))
	t.Setenv("PAISA_DECISION_HISTORY_FILE", decisionHistoryFile)

	router := gin.New()
	router.GET("/api/decision_history", decisionHistoryHandler())
	record := httptest.NewRecorder()
	router.ServeHTTP(record, httptest.NewRequest(http.MethodGet, "/api/decision_history", nil))

	assert.Equal(t, http.StatusServiceUnavailable, record.Code)
	assert.Contains(t, record.Body.String(), "unsupported decision history schema version")
}
