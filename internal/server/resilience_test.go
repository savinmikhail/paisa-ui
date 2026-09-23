package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResilienceEndpointReturnsDeterministicSyntheticContract(t *testing.T) {
	router := gin.New()
	router.GET("/api/resilience", resilienceHandler())

	record := httptest.NewRecorder()
	router.ServeHTTP(record, httptest.NewRequest(http.MethodGet, "/api/resilience?source=synthetic", nil))

	require.Equal(t, http.StatusOK, record.Code)
	var report ResilienceReport
	require.NoError(t, json.Unmarshal(record.Body.Bytes(), &report))

	assert.Equal(t, "resilience.v1", report.ContractVersion)
	assert.Equal(t, "synthetic-fixture", report.Source)
	assert.Len(t, report.Points, 6)
	assert.Equal(t, "RegularPersonalBurnMedian", report.BurnProvenance.Metric)
	assert.Equal(t, "ExpenseBaseline.RegularPersonalBurnMedian", report.BurnProvenance.Source)
	assert.Equal(t, 30000.0, report.BurnProvenance.Value)
	assert.InDelta(t, report.Points[0].RUB/report.BurnProvenance.Value, report.Points[0].LifeMonths, 0.000001)

	assert.Equal(t, []string{"Checking", "Savings", "Cash", "Reserve Asset USDT"}, scopeNames(report.LiquidScope.Included))
	assert.Equal(t, []string{"Investments", "IIS", "Receivables"}, scopeNames(report.LiquidScope.Excluded))
}

func TestResilienceEndpointFailsClosedWithoutPrivateAdapter(t *testing.T) {
	router := gin.New()
	router.GET("/api/resilience", resilienceHandler())

	record := httptest.NewRecorder()
	router.ServeHTTP(record, httptest.NewRequest(http.MethodGet, "/api/resilience", nil))

	assert.Equal(t, http.StatusServiceUnavailable, record.Code)
	assert.Contains(t, record.Body.String(), "private resilience adapter is not configured")
}

func scopeNames(items []ResilienceScopeItem) []string {
	names := make([]string, len(items))
	for i, item := range items {
		names[i] = item.Name
	}
	return names
}
