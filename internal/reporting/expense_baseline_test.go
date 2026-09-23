package reporting

import (
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func baselinePosting(id uint, date, account, amount string) posting.Posting {
	parsedDate, err := time.ParseInLocation("2006-01-02", date, time.UTC)
	if err != nil {
		panic(err)
	}
	return posting.Posting{
		ID: id, Date: parsedDate, Account: account,
		Amount:        decimal.RequireFromString(amount),
		TransactionID: "tx-" + decimal.NewFromInt(int64(id)).String(),
	}
}

func exclusionByClass(items []ExpenseExclusion, class string) ExpenseExclusion {
	for _, item := range items {
		if item.Class == class {
			return item
		}
	}
	return ExpenseExclusion{}
}

func TestBuildExpenseBaselineFiltersFlowsAndKeepsOneOffCategory(t *testing.T) {
	postings := []posting.Posting{
		baselinePosting(1, "2024-01-10", "Expenses:Food", "100"),
		baselinePosting(2, "2024-02-10", "Expenses:Food", "100"),
		baselinePosting(3, "2024-03-10", "Expenses:Food", "100"),
		baselinePosting(4, "2024-04-10", "Expenses:Food", "100"),
		baselinePosting(5, "2024-05-10", "Expenses:Food", "100"),
		baselinePosting(6, "2024-06-10", "Expenses:Food", "500"),
		baselinePosting(7, "2024-06-11", "Expenses:Travel", "700"),
		baselinePosting(8, "2024-06-12", "Expenses:Business:Software", "900"),
		baselinePosting(9, "2024-06-13", "Expenses:Tax:Income", "200"),
		baselinePosting(10, "2024-06-14", "Expenses:Investments:Broker Fees", "300"),
		baselinePosting(11, "2024-06-15", "Expenses:Food", "-20"),
	}
	postings[6].TransactionNote = "; burn-class: one-off"

	result := BuildExpenseBaseline(postings, time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC))

	assert.Equal(t, "2024-01-01", result.StartDate.Format("2006-01-02"))
	assert.Equal(t, "2024-07-01", result.EndDate.Format("2006-01-02"))
	require.Len(t, result.Months, 6)
	assert.Equal(t, "2024-02", result.Months[1].Date.Format("2006-01"))
	assert.True(t, result.Months[5].PersonalExpense.Equal(decimal.RequireFromString("1180")))
	assert.True(t, result.Months[5].RegularPersonalBurn.Equal(decimal.RequireFromString("480")))
	assert.True(t, result.PersonalExpenseMedian.Equal(decimal.RequireFromString("100")))
	assert.True(t, result.RegularPersonalBurnMedian.Equal(decimal.RequireFromString("100")))

	require.Len(t, result.OneOffs, 1)
	assert.Equal(t, "Travel", result.OneOffs[0].Category)
	assert.Equal(t, uint(7), result.OneOffs[0].Posting.ID)

	assert.Equal(t, int64(900), exclusionByClass(result.Excluded, ExpenseClassBusiness).Amount.IntPart())
	assert.Equal(t, int64(200), exclusionByClass(result.Excluded, ExpenseClassTax).Amount.IntPart())
	assert.Equal(t, int64(300), exclusionByClass(result.Excluded, ExpenseClassCapitalPlacement).Amount.IntPart())

	require.Len(t, result.CategorySignals, 1)
	assert.Equal(t, "Food", result.CategorySignals[0].Category)
	assert.True(t, result.CategorySignals[0].BaselineMedian.Equal(decimal.RequireFromString("100")))
	assert.True(t, result.CategorySignals[0].Current.Equal(decimal.RequireFromString("480")))
	assert.True(t, result.CategorySignals[0].Delta.Equal(decimal.RequireFromString("380")))
	assert.True(t, result.CategorySignals[0].GrowthPercent.Equal(decimal.RequireFromString("380")))
}

func TestBuildExpenseBaselineRetainsMissingMonthsAndHandlesRefunds(t *testing.T) {
	postings := []posting.Posting{
		baselinePosting(1, "2024-01-10", "Expenses:Rent", "100"),
		baselinePosting(2, "2024-03-10", "Expenses:Rent", "120"),
		baselinePosting(3, "2024-06-10", "Expenses:Rent", "80"),
		baselinePosting(4, "2024-06-11", "Expenses:Rent", "-20"),
	}

	result := BuildExpenseBaseline(postings, time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC))

	assert.Equal(t, []string{"2024-01", "2024-02", "2024-03", "2024-04", "2024-05", "2024-06"}, []string{
		result.Months[0].Date.Format("2006-01"), result.Months[1].Date.Format("2006-01"), result.Months[2].Date.Format("2006-01"),
		result.Months[3].Date.Format("2006-01"), result.Months[4].Date.Format("2006-01"), result.Months[5].Date.Format("2006-01"),
	})
	assert.True(t, result.Months[1].PersonalExpense.IsZero())
	assert.True(t, result.Months[3].PersonalExpense.IsZero())
	assert.True(t, result.Months[5].PersonalExpense.Equal(decimal.RequireFromString("60")))
}
