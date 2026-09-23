package server

import (
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/reporting"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeIncomeTimelineKeepsExistingMonthlyOutput(t *testing.T) {
	utils.SetNow("2024-03-15")
	postings := []posting.Posting{
		{ID: 1, Date: time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC), Account: "Income:Salary", Amount: decimal.NewFromInt(-100)},
		{ID: 2, Date: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), Account: "Income:Interest:Savings", Amount: decimal.NewFromInt(-5)},
		{ID: 3, Date: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), Account: "Income:Refund:Card", Amount: decimal.NewFromInt(-20)},
		{ID: 4, Date: time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC), Account: "Income:Salary", Amount: decimal.NewFromInt(-200)},
	}

	result := computeIncomeTimeline(postings)
	require.Len(t, result, 3)
	assert.Equal(t, []uint{1}, postingIDs(result[0].Postings))
	assert.Equal(t, []uint{2}, postingIDs(result[1].Postings))
	assert.Equal(t, []uint{3}, postingIDs(result[2].Postings))
	assert.Equal(t, "2024-03-01", result[2].Date.Format("2006-01-02"))
}

func TestComputePassiveIncomeAnalysisSeparatesRealizedSourcesAndCoverage(t *testing.T) {
	utils.SetNow("2024-06-15")
	period := reporting.Period{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	postings := []posting.Posting{
		{ID: 1, Date: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC), Account: "Income:Salary", Amount: decimal.NewFromInt(-1200)},
		{ID: 2, Date: time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC), Account: "Income:Interest:Savings", Amount: decimal.NewFromInt(-40)},
		{ID: 3, Date: time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC), Account: "Expenses:Personal:Food", Amount: decimal.NewFromInt(1000)},
		{ID: 4, Date: time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC), Account: "Expenses:Personal:Food", Amount: decimal.NewFromInt(1100)},
		{ID: 5, Date: time.Date(2024, 3, 10, 0, 0, 0, 0, time.UTC), Account: "Income:Interest:Investments:Coupons", Amount: decimal.NewFromInt(-75)},
		{ID: 6, Date: time.Date(2024, 3, 12, 0, 0, 0, 0, time.UTC), Account: "Expenses:Personal:Food", Amount: decimal.NewFromInt(900)},
		{ID: 7, Date: time.Date(2024, 4, 12, 0, 0, 0, 0, time.UTC), Account: "Expenses:Personal:Food", Amount: decimal.NewFromInt(1000)},
		{ID: 8, Date: time.Date(2024, 5, 10, 0, 0, 0, 0, time.UTC), Account: "Income:Dividend:Fund", Amount: decimal.NewFromInt(-100)},
		{ID: 9, Date: time.Date(2024, 5, 12, 0, 0, 0, 0, time.UTC), Account: "Expenses:Personal:Food", Amount: decimal.NewFromInt(1050)},
		{ID: 10, Date: time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC), Account: "Income:Interest:Savings", Amount: decimal.NewFromInt(-35)},
		{ID: 11, Date: time.Date(2024, 6, 12, 0, 0, 0, 0, time.UTC), Account: "Expenses:Personal:Food", Amount: decimal.NewFromInt(950)},
		{ID: 12, Date: time.Date(2024, 6, 13, 0, 0, 0, 0, time.UTC), Account: "Income:CapitalGains:FX", Amount: decimal.NewFromInt(-999)},
	}

	result := computePassiveIncomeAnalysis(postings, period, decimal.NewFromInt(1250))
	require.Len(t, result.Timeline, 6)
	assert.True(t, result.Timeline[0].Total.Equal(decimal.NewFromInt(40)))
	assert.True(t, result.Timeline[2].Coupons.Equal(decimal.NewFromInt(75)))
	assert.True(t, result.Timeline[4].Dividends.Equal(decimal.NewFromInt(100)))
	assert.True(t, result.Trailing12Months.Total.Equal(decimal.NewFromInt(250)))
	assert.True(t, result.RegularPersonalBurn.Equal(decimal.NewFromInt(1000)))
	assert.True(t, result.FreedomIncomeFloor.Equal(decimal.NewFromInt(1250)))
	assert.True(t, result.CurrentMonth.CoverageRegularPersonalBurn.Equal(decimal.RequireFromString("0.035")))
	assert.True(t, result.CurrentMonth.CoverageFreedomIncomeFloor.Equal(decimal.RequireFromString("0.028")))
	assert.Zero(t, result.Timeline[5].Coupons)
}

func postingIDs(postings []posting.Posting) []uint {
	ids := make([]uint, 0, len(postings))
	for _, p := range postings {
		ids = append(ids, p.ID)
	}
	return ids
}
