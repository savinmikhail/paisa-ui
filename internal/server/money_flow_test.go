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

func TestMoneyFlowSyntheticEndpointIsDeterministicAndSeparatesUnresolvedTransfers(t *testing.T) {
	router := gin.New()
	router.GET("/api/money_flow", moneyFlowHandler(nil))

	first := httptest.NewRecorder()
	router.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/api/money_flow?source=synthetic", nil))
	second := httptest.NewRecorder()
	router.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/api/money_flow?source=synthetic", nil))

	require.Equal(t, http.StatusOK, first.Code)
	assert.Equal(t, first.Body.String(), second.Body.String())
	var report MoneyFlowReport
	require.NoError(t, json.Unmarshal(first.Body.Bytes(), &report))
	assert.Equal(t, moneyFlowContractVersion, report.ContractVersion)
	assert.Equal(t, "synthetic-fixture", report.Source)
	assert.Contains(t, report.FlowEvents, MoneyFlowEvent{Date: report.FlowEvents[len(report.FlowEvents)-1].Date, Category: "Internal Transfers", Amount: 11000, Currency: "RUB", Unresolved: true})
	assert.Len(t, report.WaterfallEvents, 6)
}

func TestMoneyFlowPeriodRejectsReversedRange(t *testing.T) {
	router := gin.New()
	router.GET("/api/money_flow", moneyFlowHandler(nil))
	record := httptest.NewRecorder()
	router.ServeHTTP(record, httptest.NewRequest(http.MethodGet, "/api/money_flow?from=2024-04-01&to=2024-03-01", nil))
	assert.Equal(t, http.StatusBadRequest, record.Code)
}
