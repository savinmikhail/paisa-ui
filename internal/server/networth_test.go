package server

import (
	"testing"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/model/price"
	"github.com/ananthakumaran/paisa/internal/reporting"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

func TestGetNetworthReconciliationUsesAllTransactionPostings(t *testing.T) {
	utils.SetNow("2024-02-03")
	service.ClearPriceCache()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&posting.Posting{}, &price.Price{}))

	date := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	postings := []posting.Posting{
		{ID: 1, TransactionID: "opening", Date: date.AddDate(0, 0, -1), Account: "Assets:Checking:Fixture", Commodity: "RUB", Amount: decimal.NewFromInt(1_000), Quantity: decimal.NewFromInt(1_000)},
		{ID: 2, TransactionID: "salary", Date: date.AddDate(0, 0, 1), Payee: "Employer", Account: "Assets:Checking:Fixture", Commodity: "RUB", Amount: decimal.NewFromInt(100), Quantity: decimal.NewFromInt(100)},
		{ID: 3, TransactionID: "salary", Date: date.AddDate(0, 0, 1), Payee: "Employer", Account: "Income:Salary", Commodity: "RUB", Amount: decimal.NewFromInt(-100), Quantity: decimal.NewFromInt(-100)},
		{ID: 4, TransactionID: "card-purchase", Date: date.AddDate(0, 0, 1), Payee: "Groceries", Account: "Expenses:Household", Commodity: "RUB", Amount: decimal.NewFromInt(30), Quantity: decimal.NewFromInt(30)},
		{ID: 5, TransactionID: "card-purchase", Date: date.AddDate(0, 0, 1), Payee: "Groceries", Account: "Liabilities:CreditCards:Fixture", Commodity: "RUB", Amount: decimal.NewFromInt(-30), Quantity: decimal.NewFromInt(-30)},
	}
	for i := range postings {
		require.NoError(t, db.Create(&postings[i]).Error)
	}

	result := GetNetworthReconciliation(db, date, date.AddDate(0, 0, 1))
	decomposition := result.Decomposition
	assert.True(t, decomposition.NetSavings.Equal(decimal.NewFromInt(70)))
	assert.True(t, decomposition.Debt.IsZero())
	assert.True(t, decomposition.Corrections.IsZero())

	require.Len(t, result.Operations, 2)
	assert.Equal(t, "salary", result.Operations[0].TransactionID)
	assert.Equal(t, "netSavings", result.Operations[0].Category)
	assert.True(t, result.Operations[0].Amount.Equal(decimal.NewFromInt(100)))
	assert.Equal(t, "Employer", result.Operations[0].Payee)
	assert.Equal(t, []string{"Assets:Checking:Fixture", "Income:Salary"}, result.Operations[0].Accounts)
	assert.Equal(t, "card-purchase", result.Operations[1].TransactionID)
	assert.Equal(t, "netSavings", result.Operations[1].Category)
	assert.True(t, result.Operations[1].Amount.Equal(decimal.NewFromInt(-30)))
}

func TestNetworthOperationsKeepDebtAndUnattributedSourcesSeparate(t *testing.T) {
	date := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	postings := []posting.Posting{
		{ID: 1, TransactionID: "liability-adjustment", Date: date, Payee: "Card correction", Account: "Liabilities:CreditCards:Fixture", Commodity: "RUB", Amount: decimal.NewFromInt(15)},
		{ID: 2, TransactionID: "manual-adjustment", Date: date, Payee: "Opening balance", Account: "Assets:Checking:Fixture", Commodity: "RUB", Amount: decimal.NewFromInt(200)},
		{ID: 3, TransactionID: "borrow", Date: date, Payee: "Card draw", Account: "Assets:Checking:Fixture", Commodity: "RUB", Amount: decimal.NewFromInt(50)},
		{ID: 4, TransactionID: "borrow", Date: date, Payee: "Card draw", Account: "Liabilities:CreditCards:Fixture", Commodity: "RUB", Amount: decimal.NewFromInt(-50)},
		{ID: 5, TransactionID: "transfer", Date: date, Payee: "Own account transfer", Account: "Assets:Checking:Fixture", Commodity: "RUB", Amount: decimal.NewFromInt(80)},
		{ID: 6, TransactionID: "transfer", Date: date, Payee: "Own account transfer", Account: "Equity:Transfers:Cleared", Commodity: "RUB", Amount: decimal.NewFromInt(-80)},
	}

	operations := networthOperations(postings, date, date)
	require.Len(t, operations, 3)
	assert.Equal(t, "manual-adjustment", operations[0].TransactionID)
	assert.Equal(t, "unattributed", operations[0].Category)
	assert.True(t, operations[0].Amount.Equal(decimal.NewFromInt(200)))
	assert.Equal(t, "transfer", operations[1].TransactionID)
	assert.Equal(t, "internalTransfer", operations[1].Category)
	assert.True(t, operations[1].Amount.Equal(decimal.NewFromInt(80)))
	assert.Equal(t, "liability-adjustment", operations[2].TransactionID)
	assert.Equal(t, "debt", operations[2].Category)
	assert.True(t, operations[2].Amount.Equal(decimal.NewFromInt(15)))
}
