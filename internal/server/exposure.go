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

type CuratedExposureMetadata struct {
	Value      *string `json:"value"`
	Source     *string `json:"source"`
	VerifiedAt *string `json:"verified_at"`
	Confidence string  `json:"confidence"`
}

type ExposureAsset struct {
	ID             string                              `json:"id"`
	Label          string                              `json:"label"`
	Kind           string                              `json:"kind"`
	MarketValue    float64                             `json:"market_value"`
	Currency       string                              `json:"currency"`
	FailureDomains map[string]*CuratedExposureMetadata `json:"failure_domains"`
}

type ExposureReport struct {
	ContractVersion string          `json:"contract_version"`
	AsOf            string          `json:"as_of"`
	BaseCurrency    string          `json:"base_currency"`
	Source          string          `json:"source"`
	Assets          []ExposureAsset `json:"assets"`
}

func exposureHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Query("source") == "synthetic" {
			c.JSON(http.StatusOK, syntheticExposureReport())
			return
		}

		report, err := loadExposureReport(config.GetExposureMetadataPath())
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, report)
	}
}

func loadExposureReport(path string) (ExposureReport, error) {
	if strings.TrimSpace(path) == "" {
		return ExposureReport{}, errors.New("exposure metadata source is not configured")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return ExposureReport{}, fmt.Errorf("read exposure metadata: %w", err)
	}

	var report ExposureReport
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&report); err != nil {
		return ExposureReport{}, fmt.Errorf("decode exposure metadata: %w", err)
	}
	if report.ContractVersion != "exposure.v1" {
		return ExposureReport{}, fmt.Errorf("unsupported exposure contract %q", report.ContractVersion)
	}
	if report.Source != "private-adapter" {
		return ExposureReport{}, fmt.Errorf("private exposure metadata must declare source private-adapter")
	}
	if report.Assets == nil {
		return ExposureReport{}, errors.New("exposure metadata assets must be an array")
	}

	return report, nil
}

func syntheticExposureReport() ExposureReport {
	return ExposureReport{
		ContractVersion: "exposure.v1",
		AsOf:            "2026-09-01",
		BaseCurrency:    "RUB",
		Source:          "synthetic-fixture",
		Assets: []ExposureAsset{
			{
				ID: "asset-aurora-reserve", Label: "Aurora Reserve", Kind: "Cash reserve", MarketValue: 210000, Currency: "RUB",
				FailureDomains: map[string]*CuratedExposureMetadata{
					"institution":        metadata("Northstar Cooperative", "synthetic-fixture/exposure-01", "2026-08-28", "high"),
					"custodian":          metadata("Northstar Cooperative", "synthetic-fixture/exposure-01", "2026-08-28", "high"),
					"jurisdiction":       metadata("RU", "synthetic-fixture/exposure-01", "2026-08-28", "high"),
					"correspondent_bank": metadata("Harbor Correspondent", "synthetic-fixture/exposure-01", "2026-08-28", "medium"),
					"withdrawal_route":   metadata("Instant bank transfer", "synthetic-fixture/exposure-01", "2026-08-28", "high"),
					"key_control":        metadata("Two-person approval", "synthetic-fixture/exposure-01", "2026-08-28", "medium"),
				},
			},
			{
				ID: "asset-lumen-notes", Label: "Lumen Treasury Notes", Kind: "Bond", MarketValue: 180000, Currency: "RUB",
				FailureDomains: map[string]*CuratedExposureMetadata{
					"institution":        metadata("Northstar Cooperative", "synthetic-fixture/exposure-02", "2026-08-27", "high"),
					"broker":             metadata("Lumen Securities", "synthetic-fixture/exposure-02", "2026-08-27", "high"),
					"custodian":          metadata("Orbit Custody", "synthetic-fixture/exposure-02", "2026-08-27", "high"),
					"depositary":         metadata("Orbit Depositary", "synthetic-fixture/exposure-02", "2026-08-27", "medium"),
					"jurisdiction":       metadata("RU", "synthetic-fixture/exposure-02", "2026-08-27", "high"),
					"correspondent_bank": metadata("Harbor Correspondent", "synthetic-fixture/exposure-02", "2026-08-27", "medium"),
					"withdrawal_route":   metadata("T+1 bank transfer", "synthetic-fixture/exposure-02", "2026-08-27", "medium"),
					"key_control":        metadata("Settlement confirmation", "synthetic-fixture/exposure-02", "2026-08-27", "medium"),
				},
			},
			{
				ID: "asset-helio-etf", Label: "Helio Global ETF", Kind: "Fund", MarketValue: 145000, Currency: "USD",
				FailureDomains: map[string]*CuratedExposureMetadata{
					"institution":      metadata("Lumen Securities", "synthetic-fixture/exposure-03", "2026-08-26", "high"),
					"broker":           metadata("Lumen Securities", "synthetic-fixture/exposure-03", "2026-08-26", "high"),
					"custodian":        metadata("Orbit Custody", "synthetic-fixture/exposure-03", "2026-08-26", "high"),
					"depositary":       metadata("Orbit Depositary", "synthetic-fixture/exposure-03", "2026-08-26", "high"),
					"jurisdiction":     metadata("IE", "synthetic-fixture/exposure-03", "2026-08-26", "high"),
					"withdrawal_route": metadata("Broker withdrawal", "synthetic-fixture/exposure-03", "2026-08-26", "high"),
					"key_control":      metadata("Segregated custody", "synthetic-fixture/exposure-03", "2026-08-26", "high"),
				},
			},
			{
				ID: "asset-nova-reserve", Label: "Nova Digital Reserve", Kind: "Digital asset", MarketValue: 95000, Currency: "USD",
				FailureDomains: map[string]*CuratedExposureMetadata{
					"broker":           metadata("Northstar Prime", "synthetic-fixture/exposure-04", "2026-08-25", "low"),
					"custodian":        metadata("Aurora Vault", "synthetic-fixture/exposure-04", "2026-08-25", "medium"),
					"withdrawal_route": metadata("Whitelisted wallet", "synthetic-fixture/exposure-04", "2026-08-25", "medium"),
					"key_control":      metadata("Hardware signer quorum", "synthetic-fixture/exposure-04", "2026-08-25", "low"),
				},
			},
			{
				ID: "asset-cedar-note", Label: "Cedar Opportunity Note", Kind: "Private note", MarketValue: 70000, Currency: "RUB",
				FailureDomains: map[string]*CuratedExposureMetadata{
					"institution":        metadata("Northstar Cooperative", "synthetic-fixture/exposure-05", "2026-08-20", "medium"),
					"custodian":          metadata("Orbit Custody", "synthetic-fixture/exposure-05", "2026-08-20", "medium"),
					"jurisdiction":       metadata("RU", "synthetic-fixture/exposure-05", "2026-08-20", "medium"),
					"correspondent_bank": metadata("Harbor Correspondent", "synthetic-fixture/exposure-05", "2026-08-20", "low"),
				},
			},
		},
	}
}

func metadata(value, source, verifiedAt, confidence string) *CuratedExposureMetadata {
	return &CuratedExposureMetadata{Value: stringPointer(value), Source: stringPointer(source), VerifiedAt: stringPointer(verifiedAt), Confidence: confidence}
}

func stringPointer(value string) *string {
	return &value
}
