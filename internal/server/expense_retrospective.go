package server

import (
	"sort"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/reporting"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/shopspring/decimal"
)

type ExpenseCommitment struct {
	Key      string          `json:"key"`
	Category string          `json:"category"`
	Payee    string          `json:"payee"`
	Amount   decimal.Decimal `json:"amount"`
	Interval int             `json:"interval"`
	Period   string          `json:"period"`
	LastDate time.Time       `json:"last_date"`
	Count    int             `json:"count"`
}

type ExpenseRetrospective struct {
	reporting.ExpenseBaseline
	RecurringCommitments []ExpenseCommitment `json:"recurring_commitments"`
}

func computeRecurringExpenseCommitments(postings []posting.Posting) []ExpenseCommitment {
	sequences := ComputeRecurringTransactions(postings)
	commitments := make([]ExpenseCommitment, 0, len(sequences))
	for _, sequence := range sequences {
		if len(sequence.Transactions) == 0 {
			continue
		}
		latest := sequence.Transactions[0]
		categories := make(map[string]struct{})
		amount := decimal.Zero
		for _, p := range latest.Postings {
			if !reporting.IsPersonalExpense(p) || reporting.IsOneOff(p) {
				continue
			}
			amount = amount.Add(p.Amount)
			categories[reporting.ReportingCategory(p.Account)] = struct{}{}
		}
		if !amount.GreaterThan(decimal.Zero) {
			continue
		}
		categoryNames := make([]string, 0, len(categories))
		for category := range categories {
			categoryNames = append(categoryNames, category)
		}
		sort.Strings(categoryNames)
		category := ""
		if len(categoryNames) > 0 {
			category = categoryNames[0]
			if len(categoryNames) > 1 {
				category += " +"
			}
		}
		commitments = append(commitments, ExpenseCommitment{
			Key: sequence.Key, Category: category, Payee: latest.Payee,
			Amount: amount, Interval: sequence.Interval, Period: sequence.Period,
			LastDate: latest.Date, Count: len(sequence.Transactions),
		})
	}
	sort.SliceStable(commitments, func(i, j int) bool {
		if !commitments[i].Amount.Equal(commitments[j].Amount) {
			return commitments[i].Amount.GreaterThan(commitments[j].Amount)
		}
		return commitments[i].Key < commitments[j].Key
	})
	return commitments
}

func buildExpenseRetrospective(postings []posting.Posting) ExpenseRetrospective {
	baseline := reporting.BuildExpenseBaseline(postings, utils.Now())
	return ExpenseRetrospective{
		ExpenseBaseline:      baseline,
		RecurringCommitments: computeRecurringExpenseCommitments(postings),
	}
}
