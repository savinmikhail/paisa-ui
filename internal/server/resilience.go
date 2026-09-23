package server

import (
	"net/http"
	"time"

	"github.com/ananthakumaran/paisa/internal/reporting"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type ResiliencePoint struct {
	Date       time.Time `json:"date"`
	RUB        float64   `json:"rub"`
	USD        float64   `json:"usd"`
	LifeMonths float64   `json:"life_months"`
}

type ResilienceScopeItem struct {
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Currency string `json:"currency"`
	Reason   string `json:"reason"`
}

type ResilienceLiquidScope struct {
	Included []ResilienceScopeItem `json:"included"`
	Excluded []ResilienceScopeItem `json:"excluded"`
}

type ResilienceBurnProvenance struct {
	Metric      string    `json:"metric"`
	Source      string    `json:"source"`
	Value       float64   `json:"value"`
	Currency    string    `json:"currency"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	Description string    `json:"description"`
}

type ResilienceReport struct {
	ContractVersion string                   `json:"contract_version"`
	Source          string                   `json:"source"`
	AsOf            string                   `json:"as_of"`
	BaseCurrency    string                   `json:"base_currency"`
	RangeStart      time.Time                `json:"range_start"`
	RangeEnd        time.Time                `json:"range_end"`
	LiquidScope     ResilienceLiquidScope    `json:"liquid_scope"`
	BurnProvenance  ResilienceBurnProvenance `json:"burn_provenance"`
	Points          []ResiliencePoint        `json:"points"`
}

func resilienceHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Query("source") == "synthetic" {
			c.JSON(http.StatusOK, syntheticResilienceReport())
			return
		}

		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "private resilience adapter is not configured; use source=synthetic for the deterministic fixture",
		})
	}
}

func syntheticResilienceReport() ResilienceReport {
	utc := time.UTC
	baseline := reporting.ExpenseBaseline{
		StartDate:                 time.Date(2024, 7, 1, 0, 0, 0, 0, utc),
		EndDate:                   time.Date(2025, 1, 1, 0, 0, 0, 0, utc),
		RegularPersonalBurnMedian: decimal.NewFromInt(30000),
	}
	burn, _ := baseline.RegularPersonalBurnMedian.Float64()
	points := []ResiliencePoint{
		{Date: time.Date(2025, 1, 1, 0, 0, 0, 0, utc), RUB: 720000, USD: 9000},
		{Date: time.Date(2025, 2, 1, 0, 0, 0, 0, utc), RUB: 690000, USD: 8800},
		{Date: time.Date(2025, 3, 1, 0, 0, 0, 0, utc), RUB: 750000, USD: 9200},
		{Date: time.Date(2025, 4, 1, 0, 0, 0, 0, utc), RUB: 710000, USD: 9000},
		{Date: time.Date(2025, 5, 1, 0, 0, 0, 0, utc), RUB: 680000, USD: 8600},
		{Date: time.Date(2025, 6, 1, 0, 0, 0, 0, utc), RUB: 760000, USD: 9500},
	}
	for i := range points {
		points[i].LifeMonths = points[i].RUB / burn
	}

	return ResilienceReport{
		ContractVersion: "resilience.v1",
		Source:          "synthetic-fixture",
		AsOf:            "2025-06-30",
		BaseCurrency:    "RUB",
		RangeStart:      points[0].Date,
		RangeEnd:        points[len(points)-1].Date,
		LiquidScope: ResilienceLiquidScope{
			Included: []ResilienceScopeItem{
				{Name: "Checking", Kind: "cash account", Currency: "RUB", Reason: "operational liquidity"},
				{Name: "Savings", Kind: "cash account", Currency: "RUB", Reason: "available reserve"},
				{Name: "Cash", Kind: "cash account", Currency: "RUB", Reason: "available reserve"},
				{Name: "Reserve Asset USDT", Kind: "reserve asset", Currency: "USDT", Reason: "deliberate liquid reserve inclusion"},
			},
			Excluded: []ResilienceScopeItem{
				{Name: "Investments", Kind: "investment account", Currency: "mixed", Reason: "not liquid runway"},
				{Name: "IIS", Kind: "locked investment account", Currency: "RUB", Reason: "locked / restricted access"},
				{Name: "Receivables", Kind: "receivable", Currency: "mixed", Reason: "not accessible cash"},
			},
		},
		BurnProvenance: ResilienceBurnProvenance{
			Metric:      "RegularPersonalBurnMedian",
			Source:      "ExpenseBaseline.RegularPersonalBurnMedian",
			Value:       burn,
			Currency:    "RUB",
			PeriodStart: baseline.StartDate,
			PeriodEnd:   baseline.EndDate,
			Description: "Median of six complete months of regular personal burn; one-offs and excluded classes are outside this denominator.",
		},
		Points: points,
	}
}
