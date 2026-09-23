package reporting

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestReconcileNetWorthSumsToClosingMinusOpening(t *testing.T) {
	decomposition := ReconcileNetWorth(
		decimal.NewFromInt(1_000),
		decimal.NewFromInt(1_275),
		NetWorthFlows{
			NetSavings:    decimal.NewFromInt(140),
			PassiveIncome: decimal.NewFromInt(25),
			Market:        decimal.NewFromInt(80),
			FX:            decimal.NewFromInt(10),
			Debt:          decimal.NewFromInt(-15),
		},
	)

	assert.Equal(t, decimal.NewFromInt(275), decomposition.TotalDelta)
	assert.Equal(t, decimal.NewFromInt(35), decomposition.Corrections)
	assert.Equal(t, decomposition.TotalDelta, decomposition.Sum())
}

func TestReconcileNetWorthKeepsExplicitZeroCategories(t *testing.T) {
	decomposition := ReconcileNetWorth(
		decimal.NewFromInt(500),
		decimal.NewFromInt(450),
		NetWorthFlows{NetSavings: decimal.NewFromInt(-50)},
	)

	assert.True(t, decomposition.PassiveIncome.IsZero())
	assert.True(t, decomposition.Market.IsZero())
	assert.True(t, decomposition.FX.IsZero())
	assert.True(t, decomposition.Debt.IsZero())
	assert.True(t, decomposition.Corrections.IsZero())
	assert.True(t, decomposition.TotalDelta.Equal(decomposition.Sum()))
}
