package server

import (
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func recurringExpensePosting(id uint, transactionID, date, account, amount string) posting.Posting {
	parsedDate, err := time.ParseInLocation("2006-01-02", date, time.UTC)
	if err != nil {
		panic(err)
	}
	return posting.Posting{
		ID: id, TransactionID: transactionID, Date: parsedDate,
		Payee: "Synthetic rent", Account: account,
		Amount: decimal.RequireFromString(amount), TagRecurring: "Synthetic rent",
		TagPeriod: "0 9 1 * *", TransactionBeginLine: uint64(id), TransactionEndLine: uint64(id + 1),
	}
}

func TestComputeRecurringExpenseCommitmentsKeepsPersonalCommitments(t *testing.T) {
	utils.SetNow("2024-07-15")
	postings := []posting.Posting{
		recurringExpensePosting(1, "rent-1", "2024-05-01", "Expenses:Rent", "100"),
		recurringExpensePosting(2, "rent-1", "2024-05-01", "Assets:Checking", "-100"),
		recurringExpensePosting(3, "rent-2", "2024-06-01", "Expenses:Rent", "100"),
		recurringExpensePosting(4, "rent-2", "2024-06-01", "Assets:Checking", "-100"),
		recurringExpensePosting(5, "business-1", "2024-05-03", "Expenses:Business:Software", "900"),
		recurringExpensePosting(6, "business-1", "2024-05-03", "Assets:Checking", "-900"),
		recurringExpensePosting(7, "business-2", "2024-06-03", "Expenses:Business:Software", "900"),
		recurringExpensePosting(8, "business-2", "2024-06-03", "Assets:Checking", "-900"),
	}
	for i := 4; i < len(postings); i++ {
		postings[i].TagRecurring = "Synthetic business"
	}

	commitments := computeRecurringExpenseCommitments(postings)

	require.Len(t, commitments, 1)
	assert.Equal(t, "Synthetic rent", commitments[0].Key)
	assert.Equal(t, "Rent", commitments[0].Category)
	assert.Equal(t, int64(100), commitments[0].Amount.IntPart())
	assert.Equal(t, 31, commitments[0].Interval)
}
