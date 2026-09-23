package reporting

import "github.com/shopspring/decimal"

// NetWorthFlows contains the explainable contributions that are available from
// the current read-only reporting seam. Values are expressed in the default
// currency and preserve the sign of their contribution to net worth.
type NetWorthFlows struct {
	NetSavings    decimal.Decimal
	PassiveIncome decimal.Decimal
	Market        decimal.Decimal
	FX            decimal.Decimal
	Debt          decimal.Decimal
}

// NetWorthDecomposition is an additive explanation of an interval's net worth
// change. Corrections is deliberately explicit: it absorbs valuation events
// that do not yet have a source-level market/FX event in the reporting seam.
type NetWorthDecomposition struct {
	NetSavings    decimal.Decimal `json:"netSavings"`
	PassiveIncome decimal.Decimal `json:"passiveIncome"`
	Market        decimal.Decimal `json:"market"`
	FX            decimal.Decimal `json:"fx"`
	Debt          decimal.Decimal `json:"debt"`
	Corrections   decimal.Decimal `json:"corrections"`
	TotalDelta    decimal.Decimal `json:"totalDelta"`
}

func (d NetWorthDecomposition) Sum() decimal.Decimal {
	return d.NetSavings.
		Add(d.PassiveIncome).
		Add(d.Market).
		Add(d.FX).
		Add(d.Debt).
		Add(d.Corrections)
}

// ReconcileNetWorth closes the accounting identity for the selected interval.
// Keeping the residual as a named component prevents a partial classifier from
// presenting a plausible but unreconciled chart.
func ReconcileNetWorth(opening, closing decimal.Decimal, flows NetWorthFlows) NetWorthDecomposition {
	totalDelta := closing.Sub(opening)
	known := flows.NetSavings.
		Add(flows.PassiveIncome).
		Add(flows.Market).
		Add(flows.FX).
		Add(flows.Debt)

	return NetWorthDecomposition{
		NetSavings:    flows.NetSavings,
		PassiveIncome: flows.PassiveIncome,
		Market:        flows.Market,
		FX:            flows.FX,
		Debt:          flows.Debt,
		Corrections:   totalDelta.Sub(known),
		TotalDelta:    totalDelta,
	}
}
