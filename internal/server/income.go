package server

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/reporting"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type IncomeYearlyCard struct {
	StartDate   time.Time         `json:"start_date"`
	EndDate     time.Time         `json:"end_date"`
	Postings    []posting.Posting `json:"postings"`
	GrossIncome decimal.Decimal   `json:"gross_income"`
	NetTax      decimal.Decimal   `json:"net_tax"`
	NetIncome   decimal.Decimal   `json:"net_income"`
}

type Income struct {
	Date     time.Time         `json:"date"`
	Postings []posting.Posting `json:"postings"`
}

type Tax struct {
	StartDate time.Time         `json:"start_date"`
	EndDate   time.Time         `json:"end_date"`
	Postings  []posting.Posting `json:"postings"`
}

func GetIncome(db *gorm.DB) gin.H {
	incomePostings := query.Init(db).Like("Income:%").All()
	taxPostings := query.Init(db).AccountPrefix("Expenses:Tax").All()
	p := query.Init(db).First()

	if p == nil {
		return gin.H{"income_timeline": []Income{}, "tax_timeline": []Tax{}, "yearly_cards": []IncomeYearlyCard{}}
	}

	return gin.H{"income_timeline": computeIncomeTimeline(incomePostings), "tax_timeline": computeTaxTimeline(taxPostings), "yearly_cards": computeIncomeYearlyCard(p.Date, taxPostings, incomePostings)}
}

func computeIncomeTimeline(postings []posting.Posting) []Income {
	incomes := make([]Income, 0)

	if len(postings) == 0 {
		return incomes
	}

	start := utils.BeginningOfMonth(postings[0].Date)
	end := utils.BeginningOfMonth(utils.EndOfToday()).AddDate(0, 1, 0)
	period := reporting.Period{Start: start, End: end}
	for _, month := range reporting.AggregateReportMonthly(postings, period, reporting.Options{}, reporting.IncomeReport) {
		incomes = append(incomes, Income{Date: month.Period.Start, Postings: postingsFromEntries(month.Entries)})
	}
	return incomes
}

func computeTaxTimeline(postings []posting.Posting) []Tax {
	taxes := make([]Tax, 0)

	if len(postings) == 0 {
		return taxes
	}

	end := utils.EndOfToday()
	for start := utils.BeginningOfFinancialYear(postings[0].Date); start.Before(end); start = start.AddDate(1, 0, 0) {
		yearEnd := utils.EndOfFinancialYear(start)
		period := reporting.Period{Start: start, End: yearEnd.Add(time.Nanosecond)}
		entries := reporting.SelectReport(postings, period, reporting.Options{}, reporting.TaxReport)
		taxes = append(taxes, Tax{StartDate: start, EndDate: yearEnd, Postings: postingsFromEntries(entries)})

	}
	return taxes
}

func computeIncomeYearlyCard(start time.Time, taxes []posting.Posting, incomes []posting.Posting) []IncomeYearlyCard {
	yearlyCards := make([]IncomeYearlyCard, 0)

	end := utils.EndOfToday()
	for start = utils.BeginningOfFinancialYear(start); start.Before(end); start = start.AddDate(1, 0, 0) {
		yearEnd := utils.EndOfFinancialYear(start)
		period := reporting.Period{Start: start, End: yearEnd.Add(time.Nanosecond)}
		taxEntries := reporting.SelectReport(taxes, period, reporting.Options{}, reporting.TaxReport)
		incomeEntries := reporting.SelectReport(incomes, period, reporting.Options{}, reporting.IncomeReport)
		netTax := decimal.Zero
		for _, entry := range taxEntries {
			netTax = netTax.Add(entry.Posting.Amount)
		}
		currentYearIncomes := postingsFromEntries(incomeEntries)

		grossIncome := utils.SumBy(currentYearIncomes, func(p posting.Posting) decimal.Decimal {
			return p.Amount.Neg()
		})

		yearlyCards = append(yearlyCards, IncomeYearlyCard{
			StartDate:   start,
			EndDate:     yearEnd,
			Postings:    currentYearIncomes,
			NetTax:      netTax,
			GrossIncome: grossIncome,
			NetIncome:   grossIncome.Sub(netTax),
		})

	}
	return yearlyCards
}

func postingsFromEntries(entries []reporting.Entry) []posting.Posting {
	postings := make([]posting.Posting, 0, len(entries))
	for _, entry := range entries {
		postings = append(postings, entry.Posting)
	}
	return postings
}
