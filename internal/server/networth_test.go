package server

import (
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/reporting"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNetworthFlowsClassifySyntheticInterval(t *testing.T) {
	date := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	postings := []posting.Posting{
		{ID: 1, TransactionID: "salary", Date: date, Account: "Assets:Checking:Fixture", Amount: decimal.NewFromInt(100)},
		{ID: 2, TransactionID: "salary", Date: date, Account: "Income:Salary", Amount: decimal.NewFromInt(-100)},
		{ID: 3, TransactionID: "passive", Date: date, Account: "Assets:Savings:Fixture", Amount: decimal.NewFromInt(5)},
		{ID: 4, TransactionID: "passive", Date: date, Account: "Income:Interest:Fixture", Amount: decimal.NewFromInt(-5)},
		{ID: 5, TransactionID: "market", Date: date, Account: "Income:CapitalGains:Fixture", Amount: decimal.NewFromInt(-20)},
		{ID: 6, TransactionID: "expense", Date: date, Account: "Assets:Checking:Fixture", Amount: decimal.NewFromInt(-30)},
		{ID: 7, TransactionID: "expense", Date: date, Account: "Expenses:Household", Amount: decimal.NewFromInt(30)},
		{ID: 8, TransactionID: "borrow", Date: date, Account: "Assets:Checking:Fixture", Amount: decimal.NewFromInt(50)},
		{ID: 9, TransactionID: "borrow", Date: date, Account: "Liabilities:Fixture", Amount: decimal.NewFromInt(-50)},
		{ID: 10, TransactionID: "forgiven", Date: date, Account: "Liabilities:Fixture", Amount: decimal.NewFromInt(15)},
	}

	flows := networthFlows(postings, date, date)
	assert.True(t, flows.NetSavings.Equal(decimal.NewFromInt(70)))
	assert.True(t, flows.PassiveIncome.Equal(decimal.NewFromInt(5)))
	assert.True(t, flows.Market.Equal(decimal.NewFromInt(20)))
	assert.True(t, flows.Debt.Equal(decimal.NewFromInt(15)))
	assert.True(t, flows.FX.IsZero())

	decomposition := reporting.ReconcileNetWorth(
		decimal.NewFromInt(1_000),
		decimal.NewFromInt(1_110),
		flows,
	)
	assert.True(t, decomposition.Corrections.IsZero())
	assert.True(t, decomposition.TotalDelta.Equal(decomposition.Sum()))
}
