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

func TestExposureEndpointReturnsDeterministicSyntheticContract(t *testing.T) {
	router := gin.New()
	router.GET("/api/exposure", exposureHandler())

	record := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/exposure?source=synthetic", nil)
	router.ServeHTTP(record, request)

	require.Equal(t, http.StatusOK, record.Code)
	var report ExposureReport
	require.NoError(t, json.Unmarshal(record.Body.Bytes(), &report))
	assert.Equal(t, "exposure.v1", report.ContractVersion)
	assert.Equal(t, "synthetic-fixture", report.Source)
	assert.Len(t, report.Assets, 5)
	assert.Equal(t, "Northstar Cooperative", *report.Assets[0].FailureDomains["institution"].Value)
}

func TestExposureEndpointLoadsPrivateReportWithoutTransformingUnknownMetadata(t *testing.T) {
	metadataFile := filepath.Join(t.TempDir(), "exposure.json")
	unknownBroker := (*CuratedExposureMetadata)(nil)
	report := ExposureReport{
		ContractVersion: "exposure.v1",
		AsOf:            "2026-09-01",
		BaseCurrency:    "RUB",
		Source:          "private-adapter",
		Assets: []ExposureAsset{{
			ID: "private-asset-01", Label: "Synthetic test asset", Kind: "test", MarketValue: 123.45, Currency: "RUB",
			FailureDomains: map[string]*CuratedExposureMetadata{
				"institution": metadata("Private adapter institution", "private-test-source", "2026-08-31", "medium"),
				"broker":      unknownBroker,
			},
		}},
	}
	content, err := json.Marshal(report)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(metadataFile, content, 0600))
	t.Setenv("PAISA_EXPOSURE_METADATA_FILE", metadataFile)

	router := gin.New()
	router.GET("/api/exposure", exposureHandler())
	record := httptest.NewRecorder()
	router.ServeHTTP(record, httptest.NewRequest(http.MethodGet, "/api/exposure", nil))

	require.Equal(t, http.StatusOK, record.Code)
	var loaded ExposureReport
	require.NoError(t, json.Unmarshal(record.Body.Bytes(), &loaded))
	assert.Equal(t, "private-adapter", loaded.Source)
	assert.Nil(t, loaded.Assets[0].FailureDomains["broker"])
	assert.Equal(t, "private-test-source", *loaded.Assets[0].FailureDomains["institution"].Source)
}

func TestExposureEndpointFailsClosedWhenPrivateSourceIsNotConfigured(t *testing.T) {
	t.Setenv("PAISA_EXPOSURE_METADATA_FILE", "")
	router := gin.New()
	router.GET("/api/exposure", exposureHandler())

	record := httptest.NewRecorder()
	router.ServeHTTP(record, httptest.NewRequest(http.MethodGet, "/api/exposure", nil))

	assert.Equal(t, http.StatusServiceUnavailable, record.Code)
	assert.Contains(t, record.Body.String(), "not configured")
}
