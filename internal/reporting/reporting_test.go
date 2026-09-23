package reporting

import (
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var reportingLocation = time.UTC

func reportingPosting(id uint, date, account string, amount string) posting.Posting {
	parsedDate, err := time.ParseInLocation("2006-01-02", date, reportingLocation)
	if err != nil {
		panic(err)
	}
	return posting.Posting{
		ID:                   id,
		TransactionID:        "tx-" + decimal.NewFromInt(int64(id)).String(),
		Date:                 parsedDate,
		Account:              account,
		Amount:               decimal.RequireFromString(amount),
		FileName:             "fixture/main.ledger",
		TransactionBeginLine: 10 + uint64(id),
		TransactionEndLine:   11 + uint64(id),
	}
}

func reportingPeriod(t *testing.T, start, end string) Period {
	t.Helper()
	startDate, err := time.ParseInLocation("2006-01-02", start, reportingLocation)
	require.NoError(t, err)
	endDate, err := time.ParseInLocation("2006-01-02", end, reportingLocation)
	require.NoError(t, err)
	period, err := NewPeriod(startDate, endDate)
	require.NoError(t, err)
	return period
}

func ids(entries []Entry) []uint {
	result := make([]uint, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entry.Posting.ID)
	}
	return result
}

func TestPeriodIsInclusiveExclusive(t *testing.T) {
	period := reportingPeriod(t, "2024-01-01", "2024-02-01")

	assert.True(t, period.Contains(time.Date(2024, 1, 1, 0, 0, 0, 0, reportingLocation)))
	assert.True(t, period.Contains(time.Date(2024, 1, 31, 23, 59, 59, 0, reportingLocation)))
	assert.False(t, period.Contains(time.Date(2024, 2, 1, 0, 0, 0, 0, reportingLocation)))

	_, err := NewPeriod(period.End, period.Start)
	assert.Error(t, err)
}

func TestAccountPathReportsReuseExistingSemantics(t *testing.T) {
	postings := []posting.Posting{
		reportingPosting(1, "2024-01-02", "Income:Salary:Acme", "-100"),
		reportingPosting(2, "2024-01-03", "Expenses:Food", "25"),
		reportingPosting(3, "2024-01-04", "Expenses:Tax:Business", "10"),
		reportingPosting(4, "2024-01-05", "Income:Interest:TBank", "-5"),
		reportingPosting(5, "2024-01-06", "Income:Dividend:Broker", "-7"),
		reportingPosting(6, "2024-01-07", "Income:CapitalGains:Broker:SBER", "-12"),
		reportingPosting(7, "2024-01-08", "Equity:Transfers:Cleared", "100"),
		reportingPosting(8, "2024-01-09", "Income:Refund:CreditCard", "-20"),
		reportingPosting(9, "2024-01-10", "Assets:Checking:TBank", "1000"),
	}
	period := reportingPeriod(t, "2024-01-01", "2024-02-01")

	assert.Equal(t, []uint{1, 4, 5, 6, 8}, ids(SelectReport(postings, period, Options{}, IncomeReport)))
	assert.Equal(t, []uint{2}, ids(SelectReport(postings, period, Options{}, ExpenseReport)))
	assert.Equal(t, []uint{3}, ids(SelectReport(postings, period, Options{}, TaxReport)))
	assert.Equal(t, []uint{4, 5}, ids(SelectReport(postings, period, Options{}, PassiveIncomeReport)))
	assert.Equal(t, []uint{6}, ids(SelectReport(postings, period, Options{}, CapitalGainsReport)))
	assert.Equal(t, []uint{7}, ids(SelectReport(postings, period, Options{}, InternalTransfersReport)))
	assert.Equal(t, []uint{8}, ids(SelectReport(postings, period, Options{}, RefundsReport)))
	assert.Equal(t, []uint{9}, ids(SelectReport(postings, period, Options{}, AssetMovementsReport)))
}

func TestSelectionExcludesForecastUnlessExplicitlyIncluded(t *testing.T) {
	postings := []posting.Posting{
		reportingPosting(1, "2024-01-02", "Expenses:Food", "25"),
		reportingPosting(2, "2024-01-03", "Expenses:Food", "30"),
	}
	postings[1].Forecast = true
	period := reportingPeriod(t, "2024-01-01", "2024-02-01")

	assert.Equal(t, []uint{1}, ids(SelectReport(postings, period, Options{}, ExpenseReport)))
	assert.Equal(t, []uint{1, 2}, ids(SelectReport(postings, period, Options{IncludeForecast: true}, ExpenseReport)))
}

func TestEntriesPreservePostingAndProvenance(t *testing.T) {
	sourcePosting := reportingPosting(42, "2024-01-02", "Income:Interest:Savings", "-15")
	sourcePosting.TransactionID = "bank-row-42"
	period := reportingPeriod(t, "2024-01-01", "2024-02-01")

	entries := SelectReport([]posting.Posting{sourcePosting}, period, Options{}, PassiveIncomeReport)
	require.Len(t, entries, 1)
	assert.Equal(t, sourcePosting, entries[0].Posting)
	assert.Equal(t, Provenance{
		PostingID:            42,
		AccountPath:          "Income:Interest:Savings",
		TransactionID:        "bank-row-42",
		FileName:             "fixture/main.ledger",
		TransactionBeginLine: 52,
		TransactionEndLine:   53,
	}, entries[0].Provenance)
	assert.True(t, entries[0].PassiveIncome)
	assert.Equal(t, IncomeKind, entries[0].Kind)
}

func TestAggregateMonthlyUsesClippedPartialPeriods(t *testing.T) {
	postings := []posting.Posting{
		reportingPosting(1, "2024-01-14", "Income:Salary", "-1"),
		reportingPosting(2, "2024-01-15", "Income:Salary", "-2"),
		reportingPosting(3, "2024-01-31", "Income:Salary", "-3"),
		reportingPosting(4, "2024-02-01", "Income:Salary", "-4"),
		reportingPosting(5, "2024-03-09", "Income:Salary", "-5"),
		reportingPosting(6, "2024-03-10", "Income:Salary", "-6"),
	}
	period := reportingPeriod(t, "2024-01-15", "2024-03-10")
	months := AggregateReportMonthly(postings, period, Options{}, IncomeReport)

	require.Len(t, months, 3)
	assert.Equal(t, "2024-01", months[0].Key)
	assert.Equal(t, "2024-02", months[1].Key)
	assert.Equal(t, "2024-03", months[2].Key)
	assert.Equal(t, "2024-01-15", months[0].Period.Start.Format("2006-01-02"))
	assert.Equal(t, "2024-02-01", months[0].Period.End.Format("2006-01-02"))
	assert.Equal(t, "2024-03-10", months[2].Period.End.Format("2006-01-02"))
	assert.Equal(t, []uint{2, 3}, ids(months[0].Entries))
	assert.Equal(t, []uint{4}, ids(months[1].Entries))
	assert.Equal(t, []uint{5}, ids(months[2].Entries))
	assert.True(t, months[0].Amount.Equal(decimal.RequireFromString("-5")))
	assert.True(t, months[1].Amount.Equal(decimal.RequireFromString("-4")))
	assert.True(t, months[2].Amount.Equal(decimal.RequireFromString("-5")))
}
