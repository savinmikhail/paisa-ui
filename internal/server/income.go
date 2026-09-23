package server

import (
	"time"

	"github.com/ananthakumaran/paisa/internal/config"
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

type IncomeBaselinePoint struct {
	Date   time.Time       `json:"date"`
	Amount decimal.Decimal `json:"amount"`
}

type PassiveIncomePoint struct {
	Date      time.Time       `json:"date"`
	Interest  decimal.Decimal `json:"interest"`
	Coupons   decimal.Decimal `json:"coupons"`
	Dividends decimal.Decimal `json:"dividends"`
	Total     decimal.Decimal `json:"total"`
}

type PassiveIncomeWindow struct {
	Total                       decimal.Decimal `json:"total"`
	Average                     decimal.Decimal `json:"average"`
	CoverageRegularPersonalBurn decimal.Decimal `json:"coverage_regular_personal_burn"`
	CoverageFreedomIncomeFloor  decimal.Decimal `json:"coverage_freedom_income_floor"`
}

type PassiveIncomeAnalysis struct {
	Timeline            []PassiveIncomePoint `json:"timeline"`
	CurrentMonth        PassiveIncomeWindow  `json:"current_month"`
	Trailing3Months     PassiveIncomeWindow  `json:"trailing_3_months"`
	Trailing12Months    PassiveIncomeWindow  `json:"trailing_12_months"`
	RegularPersonalBurn decimal.Decimal      `json:"regular_personal_burn"`
	FreedomIncomeFloor  decimal.Decimal      `json:"freedom_income_floor"`
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

// GetIncomeAnalysis extends the legacy income endpoint without changing its
// response contract. All derived values come from the shared reporting seam.
func GetIncomeAnalysis(db *gorm.DB) gin.H {
	postings := query.Init(db).All()
	if len(postings) == 0 {
		return gin.H{
			"income_baseline": []IncomeBaselinePoint{},
			"passive_income":  emptyPassiveIncomeAnalysis(),
		}
	}

	period := incomeReportPeriod(postings)
	return gin.H{
		"income_baseline": computeIncomeBaseline(postings, period),
		"passive_income": computePassiveIncomeAnalysis(
			postings,
			period,
			configuredFreedomIncomeFloor(),
		),
	}
}

func incomeReportPeriod(postings []posting.Posting) reporting.Period {
	start := utils.BeginningOfMonth(postings[0].Date)
	end := utils.BeginningOfMonth(utils.EndOfToday()).AddDate(0, 1, 0)
	return reporting.Period{Start: start, End: end}
}

func computeIncomeBaseline(postings []posting.Posting, period reporting.Period) []IncomeBaselinePoint {
	months := reporting.AggregateReportMonthly(postings, period, reporting.Options{}, reporting.IncomeReport)
	result := make([]IncomeBaselinePoint, 0, len(months))
	for _, month := range months {
		result = append(result, IncomeBaselinePoint{
			Date:   month.Period.Start,
			Amount: month.Amount.Neg(),
		})
	}
	return result
}

func computePassiveIncomeAnalysis(
	postings []posting.Posting,
	period reporting.Period,
	freedomIncomeFloor decimal.Decimal,
) PassiveIncomeAnalysis {
	passiveMonths := reporting.AggregateReportMonthly(postings, period, reporting.Options{}, reporting.PassiveIncomeReport)
	burnMonths := reporting.AggregateReportMonthly(postings, period, reporting.Options{}, reporting.RegularPersonalBurnReport)

	points := make([]PassiveIncomePoint, 0, len(passiveMonths))
	for _, month := range passiveMonths {
		point := PassiveIncomePoint{Date: month.Period.Start}
		for _, entry := range month.Entries {
			amount := entry.Posting.Amount.Neg()
			category, ok := reporting.PassiveIncomeCategoryFor(entry.Posting)
			if !ok {
				continue
			}
			switch category {
			case reporting.PassiveInterestCategory:
				point.Interest = point.Interest.Add(amount)
			case reporting.PassiveCouponsCategory:
				point.Coupons = point.Coupons.Add(amount)
			case reporting.PassiveDividendsCategory:
				point.Dividends = point.Dividends.Add(amount)
			}
		}
		point.Total = point.Interest.Add(point.Coupons).Add(point.Dividends)
		points = append(points, point)
	}

	regularPersonalBurn := averageMonths(burnMonths)
	return PassiveIncomeAnalysis{
		Timeline:            points,
		CurrentMonth:        passiveIncomeWindow(lastMonths(points, 1), regularPersonalBurn, freedomIncomeFloor),
		Trailing3Months:     passiveIncomeWindow(lastMonths(points, 3), regularPersonalBurn, freedomIncomeFloor),
		Trailing12Months:    passiveIncomeWindow(lastMonths(points, 12), regularPersonalBurn, freedomIncomeFloor),
		RegularPersonalBurn: regularPersonalBurn,
		FreedomIncomeFloor:  freedomIncomeFloor,
	}
}

func emptyPassiveIncomeAnalysis() PassiveIncomeAnalysis {
	return PassiveIncomeAnalysis{
		Timeline:         []PassiveIncomePoint{},
		CurrentMonth:     PassiveIncomeWindow{},
		Trailing3Months:  PassiveIncomeWindow{},
		Trailing12Months: PassiveIncomeWindow{},
	}
}

func lastMonths(points []PassiveIncomePoint, count int) []PassiveIncomePoint {
	if len(points) <= count {
		return points
	}
	return points[len(points)-count:]
}

func averageMonths(months []reporting.Month) decimal.Decimal {
	if len(months) == 0 {
		return decimal.Zero
	}
	total := decimal.Zero
	for _, month := range months {
		total = total.Add(month.Amount)
	}
	return total.Div(decimal.NewFromInt(int64(len(months))))
}

func passiveIncomeWindow(
	points []PassiveIncomePoint,
	regularPersonalBurn decimal.Decimal,
	freedomIncomeFloor decimal.Decimal,
) PassiveIncomeWindow {
	window := PassiveIncomeWindow{}
	for _, point := range points {
		window.Total = window.Total.Add(point.Total)
	}
	if len(points) == 0 {
		return window
	}
	window.Average = window.Total.Div(decimal.NewFromInt(int64(len(points))))
	if regularPersonalBurn.GreaterThan(decimal.Zero) {
		window.CoverageRegularPersonalBurn = window.Average.Div(regularPersonalBurn)
	}
	if freedomIncomeFloor.GreaterThan(decimal.Zero) {
		window.CoverageFreedomIncomeFloor = window.Average.Div(freedomIncomeFloor)
	}
	return window
}

func configuredFreedomIncomeFloor() decimal.Decimal {
	for _, goal := range config.GetConfig().Goals.Retirement {
		if goal.YearlyExpenses > 0 {
			return decimal.NewFromFloat(goal.YearlyExpenses).Div(decimal.NewFromInt(12))
		}
	}
	return decimal.Zero
}

func computeIncomeTimeline(postings []posting.Posting) []Income {
	incomes := make([]Income, 0)

	if len(postings) == 0 {
		return incomes
	}

	period := incomeReportPeriod(postings)
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
