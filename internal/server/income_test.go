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

func postingIDs(postings []posting.Posting) []uint {
	ids := make([]uint, 0, len(postings))
	for _, p := range postings {
		ids = append(ids, p.ID)
	}
	return ids
}
