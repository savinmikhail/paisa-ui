package server

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/reporting"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Networth struct {
	Date                time.Time       `json:"date"`
	InvestmentAmount    decimal.Decimal `json:"investmentAmount"`
	WithdrawalAmount    decimal.Decimal `json:"withdrawalAmount"`
	GainAmount          decimal.Decimal `json:"gainAmount"`
	BalanceAmount       decimal.Decimal `json:"balanceAmount"`
	BalanceUnits        decimal.Decimal `json:"balanceUnits"`
	NetInvestmentAmount decimal.Decimal `json:"netInvestmentAmount"`
}

type NetworthReconciliation struct {
	From          time.Time                       `json:"from"`
	To            time.Time                       `json:"to"`
	Opening       Networth                        `json:"opening"`
	Closing       Networth                        `json:"closing"`
	Decomposition reporting.NetWorthDecomposition `json:"decomposition"`
	Operations    []NetworthOperation             `json:"operations"`
}

type NetworthOperation struct {
	Date                 time.Time       `json:"date"`
	Payee                string          `json:"payee"`
	Category             string          `json:"category"`
	Amount               decimal.Decimal `json:"amount"`
	TransactionID        string          `json:"transactionId"`
	Accounts             []string        `json:"accounts"`
	FileName             string          `json:"fileName,omitempty"`
	TransactionBeginLine uint64          `json:"transactionBeginLine,omitempty"`
	TransactionEndLine   uint64          `json:"transactionEndLine,omitempty"`
}

func GetNetworth(db *gorm.DB) gin.H {
	postings := query.Init(db).Like("Assets:%", "Income:CapitalGains:%", "Liabilities:%").UntilToday().All()

	postings = service.PopulateMarketPrice(db, postings)
	networthTimeline := computeNetworthTimeline(db, postings, false)
	xirr := service.XIRR(db, postings)
	return gin.H{"networthTimeline": networthTimeline, "xirr": xirr}
}

func GetCurrentNetworth(db *gorm.DB) gin.H {
	postings := query.Init(db).Like("Assets:%", "Income:CapitalGains:%", "Liabilities:%").UntilToday().All()
	postings = service.PopulateMarketPrice(db, postings)
	networth := computeNetworth(db, postings)
	xirr := service.XIRR(db, postings)
	return gin.H{"networth": networth, "xirr": xirr}
}

func GetNetworthReconciliation(db *gorm.DB, from, to time.Time) NetworthReconciliation {
	timelinePostings := query.Init(db).Like("Assets:%", "Income:CapitalGains:%", "Liabilities:%").UntilToday().All()
	timelinePostings = service.PopulateMarketPrice(db, timelinePostings)
	timeline := computeNetworthTimeline(db, timelinePostings, false)
	if len(timeline) == 0 {
		return NetworthReconciliation{}
	}

	if from.IsZero() {
		from = timeline[0].Date
	}
	if to.IsZero() {
		to = timeline[len(timeline)-1].Date
	}
	if to.Before(from) {
		from, to = to, from
	}

	opening := timelineAtOrBefore(timeline, from)
	closing := timelineAtOrBefore(timeline, to)
	flowPostings := query.Init(db).UntilToday().All()
	flows := networthFlows(flowPostings, from, to)

	return NetworthReconciliation{
		From:          opening.Date,
		To:            closing.Date,
		Opening:       opening,
		Closing:       closing,
		Decomposition: reporting.ReconcileNetWorth(networthValue(opening), networthValue(closing), flows),
		Operations:    networthOperations(flowPostings, from, to),
	}
}

func timelineAtOrBefore(timeline []Networth, date time.Time) Networth {
	selected := timeline[0]
	for _, point := range timeline {
		if point.Date.After(date) {
			break
		}
		selected = point
	}
	return selected
}

func networthValue(point Networth) decimal.Decimal {
	return point.InvestmentAmount.Add(point.GainAmount).Sub(point.WithdrawalAmount)
}

func networthFlows(postings []posting.Posting, from, to time.Time) reporting.NetWorthFlows {
	byTransaction := make(map[string][]posting.Posting)
	for _, p := range postings {
		if p.Date.Before(from) || !p.Date.Before(to.AddDate(0, 0, 1)) {
			continue
		}
		key := p.TransactionID
		if key == "" {
			key = fmt.Sprintf("%s/%s/%s/%d", p.Date.Format(time.RFC3339), p.Account, p.Commodity, p.ID)
		}
		byTransaction[key] = append(byTransaction[key], p)
	}

	var flows reporting.NetWorthFlows
	for _, transactionPostings := range byTransaction {
		hasFlowAccount := false
		hasLiability := false
		for _, p := range transactionPostings {
			switch {
			case reporting.IsPassiveIncome(p):
				flows.PassiveIncome = flows.PassiveIncome.Add(p.Amount.Neg())
				hasFlowAccount = true
			case reporting.IsCapitalGain(p):
				flows.Market = flows.Market.Add(p.Amount.Neg())
				hasFlowAccount = true
			case strings.HasPrefix(p.Account, "Income:"):
				flows.NetSavings = flows.NetSavings.Add(p.Amount.Neg())
				hasFlowAccount = true
			case strings.HasPrefix(p.Account, "Expenses:"):
				flows.NetSavings = flows.NetSavings.Add(p.Amount.Neg())
				hasFlowAccount = true
			case strings.HasPrefix(p.Account, "Liabilities:"):
				hasLiability = true
			}
		}

		// A pure asset/liability transaction is usually borrowing or repayment:
		// the asset leg offsets the liability leg. Only the unmatched net leg is
		// shown as debt; expense-backed card purchases stay in net savings.
		if hasLiability && !hasFlowAccount {
			for _, p := range transactionPostings {
				if strings.HasPrefix(p.Account, "Assets:") || strings.HasPrefix(p.Account, "Liabilities:") {
					flows.Debt = flows.Debt.Add(p.Amount)
				}
			}
		}
	}
	return flows
}

func networthOperations(postings []posting.Posting, from, to time.Time) []NetworthOperation {
	byTransaction := make(map[string][]posting.Posting)
	for _, p := range postings {
		if p.Date.Before(from) || !p.Date.Before(to.AddDate(0, 0, 1)) {
			continue
		}
		key := p.TransactionID
		if key == "" {
			key = fmt.Sprintf("%s/%s/%s/%d", p.Date.Format(time.RFC3339), p.Account, p.Commodity, p.ID)
		}
		byTransaction[key] = append(byTransaction[key], p)
	}

	categoryOrder := []string{"netSavings", "passiveIncome", "market", "debt", "internalTransfer", "unattributed"}
	operations := make([]NetworthOperation, 0, len(byTransaction))
	for transactionID, transactionPostings := range byTransaction {
		contributions := make(map[string]decimal.Decimal)
		hasFlowAccount := false
		hasLiability := false
		hasInternalTransfer := false
		unattributed := decimal.Zero
		accounts := make([]string, 0, len(transactionPostings))
		seenAccounts := make(map[string]struct{})

		for _, p := range transactionPostings {
			if _, seen := seenAccounts[p.Account]; !seen {
				accounts = append(accounts, p.Account)
				seenAccounts[p.Account] = struct{}{}
			}
			if reporting.IsInternalTransfer(p) {
				hasInternalTransfer = true
			}
			switch {
			case reporting.IsPassiveIncome(p):
				contributions["passiveIncome"] = contributions["passiveIncome"].Add(p.Amount.Neg())
				hasFlowAccount = true
			case reporting.IsCapitalGain(p):
				contributions["market"] = contributions["market"].Add(p.Amount.Neg())
				hasFlowAccount = true
			case strings.HasPrefix(p.Account, "Income:"), strings.HasPrefix(p.Account, "Expenses:"):
				contributions["netSavings"] = contributions["netSavings"].Add(p.Amount.Neg())
				hasFlowAccount = true
			case strings.HasPrefix(p.Account, "Liabilities:"):
				hasLiability = true
			}

			if strings.HasPrefix(p.Account, "Assets:") || strings.HasPrefix(p.Account, "Liabilities:") {
				unattributed = unattributed.Add(p.Amount)
			}
		}

		if !hasFlowAccount {
			if hasLiability {
				contributions["debt"] = unattributed
			} else if hasInternalTransfer && !unattributed.IsZero() {
				contributions["internalTransfer"] = unattributed
			} else if !unattributed.IsZero() {
				contributions["unattributed"] = unattributed
			}
		}

		if len(accounts) == 0 {
			continue
		}
		sort.Strings(accounts)
		for _, category := range categoryOrder {
			amount, ok := contributions[category]
			if !ok || amount.IsZero() {
				continue
			}
			sample := transactionPostings[0]
			operations = append(operations, NetworthOperation{
				Date:                 sample.Date,
				Payee:                sample.Payee,
				Category:             category,
				Amount:               amount,
				TransactionID:        transactionID,
				Accounts:             accounts,
				FileName:             sample.FileName,
				TransactionBeginLine: sample.TransactionBeginLine,
				TransactionEndLine:   sample.TransactionEndLine,
			})
		}
	}

	sort.SliceStable(operations, func(i, j int) bool {
		if operations[i].Date.Equal(operations[j].Date) {
			if operations[i].Amount.Abs().Equal(operations[j].Amount.Abs()) {
				return operations[i].TransactionID < operations[j].TransactionID
			}
			return operations[i].Amount.Abs().GreaterThan(operations[j].Amount.Abs())
		}
		return operations[i].Date.Before(operations[j].Date)
	})
	return operations
}

func computeNetworth(db *gorm.DB, postings []posting.Posting) Networth {
	var networth Networth

	if len(postings) == 0 {
		return networth
	}

	var investment decimal.Decimal = decimal.Zero
	var withdrawal decimal.Decimal = decimal.Zero
	var balance decimal.Decimal = decimal.Zero

	now := utils.EndOfToday()
	for _, p := range postings {
		isInterest := service.IsInterest(db, p)
		isInterestRepayment := service.IsInterestRepayment(db, p)
		isStockSplit := service.IsStockSplit(db, p)
		isCapitalGains := service.IsCapitalGains(p)

		if isInterest || isInterestRepayment {
			balance = balance.Add(p.Amount)
		} else if isCapitalGains {
			withdrawal = withdrawal.Add(p.Amount.Neg())
		} else {
			if p.Amount.GreaterThan(decimal.Zero) && !isStockSplit {
				investment = investment.Add(p.Amount)
			}

			if p.Amount.LessThan(decimal.Zero) && !isStockSplit {
				withdrawal = withdrawal.Add(p.Amount.Neg())
			}

			balance = balance.Add(service.GetMarketPrice(db, p, now))
		}
	}

	gain := balance.Add(withdrawal).Sub(investment)
	netInvestment := investment.Sub(withdrawal)
	networth = Networth{
		Date:                now,
		InvestmentAmount:    investment,
		WithdrawalAmount:    withdrawal,
		GainAmount:          gain,
		BalanceAmount:       balance,
		NetInvestmentAmount: netInvestment,
	}

	return networth
}

func computeNetworthTimeline(db *gorm.DB, postings []posting.Posting, computeBalanceUnits bool) []Networth {
	var networths []Networth

	var p posting.Posting

	if len(postings) == 0 {
		return []Networth{}
	}

	type RunningSum struct {
		investment   decimal.Decimal
		withdrawal   decimal.Decimal
		balance      decimal.Decimal
		balanceUnits decimal.Decimal
	}

	accumulator := make(map[string]RunningSum)

	end := utils.EndOfToday()
	for start := postings[0].Date; start.Before(end); start = start.AddDate(0, 0, 1) {
		for len(postings) > 0 && (postings[0].Date.Before(start) || postings[0].Date.Equal(start)) {
			p, postings = postings[0], postings[1:]
			rs := accumulator[p.Commodity]

			isInterest := service.IsInterest(db, p)
			isCapitalGains := service.IsCapitalGains(p)

			if p.Amount.GreaterThan(decimal.Zero) && !isInterest {
				rs.investment = rs.investment.Add(p.Amount)
			}

			if p.Amount.LessThan(decimal.Zero) && !isInterest {
				rs.withdrawal = rs.withdrawal.Add(p.Amount.Neg())
			}

			if !isCapitalGains {
				rs.balance = rs.balance.Add(service.GetMarketPrice(db, p, start))
				rs.balanceUnits = rs.balanceUnits.Add(p.Quantity)
			}

			accumulator[p.Commodity] = rs

		}

		var investment decimal.Decimal = decimal.Zero
		var withdrawal decimal.Decimal = decimal.Zero
		var balance decimal.Decimal = decimal.Zero
		var balanceUnits decimal.Decimal = decimal.Zero

		for commodity, rs := range accumulator {
			investment = investment.Add(rs.investment)
			withdrawal = withdrawal.Add(rs.withdrawal)

			if utils.IsCurrency(commodity) {
				balance = balance.Add(rs.balance)
			} else {
				if computeBalanceUnits {
					balanceUnits = balanceUnits.Add(rs.balanceUnits)
				}
				price := service.GetUnitPrice(db, commodity, start)
				if !price.Value.Equal(decimal.Zero) {
					balance = balance.Add(rs.balanceUnits.Mul(price.Value))
				} else {
					balance = balance.Add(rs.balance)
				}
			}

		}

		gain := balance.Add(withdrawal).Sub(investment)
		netInvestment := investment.Sub(withdrawal)
		networths = append(networths, Networth{
			Date:                start,
			InvestmentAmount:    investment,
			WithdrawalAmount:    withdrawal,
			GainAmount:          gain,
			BalanceAmount:       balance,
			BalanceUnits:        balanceUnits,
			NetInvestmentAmount: netInvestment,
		})

		if len(postings) == 0 && balance.Abs().LessThan(decimal.NewFromFloat(0.01)) {
			break
		}
	}
	return networths
}
