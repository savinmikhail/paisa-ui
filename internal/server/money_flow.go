package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/reporting"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const moneyFlowContractVersion = "money-flow.v1"

type MoneyFlowEvent struct {
	Date       time.Time `json:"date"`
	Category   string    `json:"category"`
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	Label      string    `json:"label,omitempty"`
	Unresolved bool      `json:"unresolved,omitempty"`
}

type MoneyFlowWaterfallEvent struct {
	Date   time.Time `json:"date"`
	Key    string    `json:"key"`
	Amount float64   `json:"amount"`
}

type MoneyFlowSnapshot struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

type MoneyFlowReport struct {
	ContractVersion   string                    `json:"contractVersion"`
	BaseCurrency      string                    `json:"baseCurrency"`
	Source            string                    `json:"source"`
	FlowEvents        []MoneyFlowEvent          `json:"flowEvents"`
	WaterfallEvents   []MoneyFlowWaterfallEvent `json:"waterfallEvents"`
	NetWorthSnapshots []MoneyFlowSnapshot       `json:"netWorthSnapshots"`
}

func moneyFlowHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Query("source") == "synthetic" {
			c.JSON(http.StatusOK, syntheticMoneyFlowReport())
			return
		}

		from, to, err := moneyFlowPeriod(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, getMoneyFlow(db, from, to))
	}
}

func moneyFlowPeriod(c *gin.Context) (time.Time, time.Time, error) {
	var from, to time.Time
	var err error
	if value := c.Query("from"); value != "" {
		from, err = time.Parse("2006-01-02", value)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}
	if value := c.Query("to"); value != "" {
		to, err = time.Parse("2006-01-02", value)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}
	if !from.IsZero() && !to.IsZero() && to.Before(from) {
		return time.Time{}, time.Time{}, &moneyFlowPeriodError{}
	}
	return from, to, nil
}

type moneyFlowPeriodError struct{}

func (*moneyFlowPeriodError) Error() string {
	return "money flow period end must not be before start"
}

func getMoneyFlow(db *gorm.DB, from, to time.Time) MoneyFlowReport {
	allPostings := query.Init(db).UntilToday()
	if !from.IsZero() {
		allPostings = allPostings.Where("date >= ?", from)
	}
	if !to.IsZero() {
		allPostings = allPostings.Where("date < ?", to.AddDate(0, 0, 1))
	}
	postings := allPostings.All()
	baseCurrency := config.DefaultCurrency()
	result := MoneyFlowReport{
		ContractVersion:   moneyFlowContractVersion,
		BaseCurrency:      baseCurrency,
		Source:            "private-adapter",
		FlowEvents:        make([]MoneyFlowEvent, 0),
		WaterfallEvents:   make([]MoneyFlowWaterfallEvent, 0),
		NetWorthSnapshots: make([]MoneyFlowSnapshot, 0),
	}
	for _, p := range postings {
		if event, ok := classifyMoneyFlowPosting(p, baseCurrency); ok {
			result.FlowEvents = append(result.FlowEvents, event)
		}
	}

	reconciliation := GetNetworthReconciliation(db, from, to)
	if !reconciliation.From.IsZero() {
		result.NetWorthSnapshots = append(result.NetWorthSnapshots,
			MoneyFlowSnapshot{Date: reconciliation.From, Value: networthValue(reconciliation.Opening).InexactFloat64()},
			MoneyFlowSnapshot{Date: reconciliation.To, Value: networthValue(reconciliation.Closing).InexactFloat64()},
		)
		decomposition := reconciliation.Decomposition
		result.WaterfallEvents = append(result.WaterfallEvents,
			MoneyFlowWaterfallEvent{Date: reconciliation.To, Key: "netSavings", Amount: decomposition.NetSavings.InexactFloat64()},
			MoneyFlowWaterfallEvent{Date: reconciliation.To, Key: "passiveIncome", Amount: decomposition.PassiveIncome.InexactFloat64()},
			MoneyFlowWaterfallEvent{Date: reconciliation.To, Key: "market", Amount: decomposition.Market.InexactFloat64()},
			MoneyFlowWaterfallEvent{Date: reconciliation.To, Key: "fx", Amount: decomposition.FX.InexactFloat64()},
			MoneyFlowWaterfallEvent{Date: reconciliation.To, Key: "debt", Amount: decomposition.Debt.InexactFloat64()},
			MoneyFlowWaterfallEvent{Date: reconciliation.To, Key: "corrections", Amount: decomposition.Corrections.InexactFloat64()},
		)
	}
	return result
}

func classifyMoneyFlowPosting(p posting.Posting, currency string) (MoneyFlowEvent, bool) {
	if p.Amount.IsZero() {
		return MoneyFlowEvent{}, false
	}
	event := MoneyFlowEvent{Date: p.Date, Amount: p.Amount.Abs().InexactFloat64(), Currency: currency}
	switch {
	case reporting.IsInternalTransfer(p):
		event.Category = "Internal Transfers"
		account := strings.ToLower(p.Account)
		event.Unresolved = strings.Contains(account, "unclassified") || strings.Contains(account, "pending")
	case reporting.IsIncome(p):
		event.Category = "Income"
	case reporting.IsTax(p):
		event.Category = "Tax"
	case reporting.IsExpense(p):
		if reporting.IsBusinessExpense(p) {
			event.Category = "Business"
		} else {
			event.Category = "Personal Expense"
		}
	case reporting.IsAssetMovement(p):
		if p.Amount.LessThanOrEqual(decimal.Zero) {
			return MoneyFlowEvent{}, false
		}
		account := strings.ToLower(p.Account)
		if strings.Contains(account, "savings") || strings.Contains(account, "reserve") || strings.Contains(account, "cash") {
			event.Category = "Reserve"
		} else {
			event.Category = "Capital Placement"
		}
	default:
		return MoneyFlowEvent{}, false
	}
	return event, true
}

func syntheticMoneyFlowReport() MoneyFlowReport {
	date := func(value string) time.Time {
		parsed, _ := time.Parse("2006-01-02", value)
		return parsed
	}
	events := []MoneyFlowEvent{
		{Date: date("2024-01-31"), Category: "Income", Amount: 176000, Currency: "RUB", Label: "Salary"},
		{Date: date("2024-02-29"), Category: "Income", Amount: 92000, Currency: "RUB", Label: "Contract"},
		{Date: date("2024-03-31"), Category: "Income", Amount: 138000, Currency: "RUB", Label: "Salary"},
		{Date: date("2024-01-31"), Category: "Personal Expense", Amount: 70000, Currency: "RUB"},
		{Date: date("2024-02-29"), Category: "Personal Expense", Amount: 58000, Currency: "RUB"},
		{Date: date("2024-03-31"), Category: "Business", Amount: 26000, Currency: "RUB"},
		{Date: date("2024-02-29"), Category: "Tax", Amount: 42000, Currency: "RUB"},
		{Date: date("2024-01-31"), Category: "Reserve", Amount: 64000, Currency: "RUB"},
		{Date: date("2024-03-31"), Category: "Capital Placement", Amount: 86000, Currency: "RUB"},
		{Date: date("2024-02-29"), Category: "Internal Transfers", Amount: 24000, Currency: "RUB"},
		{Date: date("2024-03-31"), Category: "Internal Transfers", Amount: 11000, Currency: "RUB", Unresolved: true},
	}
	return MoneyFlowReport{
		ContractVersion: moneyFlowContractVersion,
		BaseCurrency:    "RUB",
		Source:          "synthetic-fixture",
		FlowEvents:      events,
		WaterfallEvents: []MoneyFlowWaterfallEvent{
			{Date: date("2024-03-31"), Key: "netSavings", Amount: 82000},
			{Date: date("2024-03-31"), Key: "passiveIncome", Amount: 14000},
			{Date: date("2024-03-31"), Key: "market", Amount: 26000},
			{Date: date("2024-03-31"), Key: "fx", Amount: -9000},
			{Date: date("2024-03-31"), Key: "debt", Amount: -18000},
			{Date: date("2024-03-31"), Key: "corrections", Amount: 5000},
		},
		NetWorthSnapshots: []MoneyFlowSnapshot{
			{Date: date("2023-12-31"), Value: 640000},
			{Date: date("2024-03-31"), Value: 740000},
		},
	}
}
