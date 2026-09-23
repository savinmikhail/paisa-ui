package reporting

import (
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/shopspring/decimal"
)

const (
	ExpenseClassBusiness         = "business"
	ExpenseClassTax              = "tax"
	ExpenseClassCapitalPlacement = "capital_placement"
	ExpenseClassOtherExcluded    = "other_excluded"
)

var oneOffMarker = regexp.MustCompile(`(?im)^\s*;?\s*burn-class\s*:\s*one-off\s*$`)

// ExpenseBaseline is a read-only retrospective view. It intentionally keeps
// source postings on one-offs so the original Reporting Category remains
// inspectable while the regular burn calculation can exclude them.
type ExpenseBaseline struct {
	StartDate                 time.Time               `json:"start_date"`
	EndDate                   time.Time               `json:"end_date"`
	Months                    []ExpenseBaselineMonth  `json:"months"`
	PersonalExpenseMedian     decimal.Decimal         `json:"personal_expense_median"`
	RegularPersonalBurnMedian decimal.Decimal         `json:"regular_personal_burn_median"`
	Excluded                  []ExpenseExclusion      `json:"excluded"`
	OneOffs                   []ExpenseOneOff         `json:"one_offs"`
	CategorySignals           []ExpenseCategorySignal `json:"category_signals"`
}

type ExpenseBaselineMonth struct {
	Date                time.Time       `json:"date"`
	PersonalExpense     decimal.Decimal `json:"personal_expense"`
	RegularPersonalBurn decimal.Decimal `json:"regular_personal_burn"`
}

type ExpenseExclusion struct {
	Class  string          `json:"class"`
	Label  string          `json:"label"`
	Amount decimal.Decimal `json:"amount"`
	Count  int             `json:"count"`
}

type ExpenseOneOff struct {
	Posting  posting.Posting `json:"posting"`
	Category string          `json:"category"`
}

type ExpenseCategorySignal struct {
	Category       string          `json:"category"`
	Current        decimal.Decimal `json:"current"`
	BaselineMedian decimal.Decimal `json:"baseline_median"`
	Delta          decimal.Decimal `json:"delta"`
	GrowthPercent  decimal.Decimal `json:"growth_percent"`
}

type monthTotals struct {
	personal decimal.Decimal
	regular  decimal.Decimal
}

func IsBusinessExpense(p posting.Posting) bool {
	return AccountOrChild(p.Account, "Expenses:Business") ||
		AccountOrChild(p.Account, "Expenses:Семейный бизнес")
}

// IsCapitalPlacementExpense covers legacy expense postings used for broker,
// savings, and investment fees. Asset postings are not expenses and therefore
// never enter this baseline in the first place.
func IsCapitalPlacementExpense(p posting.Posting) bool {
	return AccountOrChild(p.Account, "Expenses:Investments")
}

func IsOtherExcludedPersonalExpense(p posting.Posting) bool {
	return AccountOrChild(p.Account, "Expenses:Мошенничество") ||
		AccountOrChild(p.Account, "Expenses:Наличные:Неразобранное")
}

func IsOneOff(p posting.Posting) bool {
	return oneOffMarker.MatchString(p.Note + "\n" + p.TransactionNote)
}

func reportingCategory(account string) string {
	parts := strings.Split(account, ":")
	if len(parts) < 2 {
		return account
	}
	return parts[1]
}

func ReportingCategory(account string) string {
	return reportingCategory(account)
}

func IsPersonalExpense(p posting.Posting) bool {
	return IsExpense(p) && expenseClass(p) == ""
}

func expenseClass(p posting.Posting) string {
	switch {
	case IsTax(p):
		return ExpenseClassTax
	case IsBusinessExpense(p):
		return ExpenseClassBusiness
	case IsCapitalPlacementExpense(p):
		return ExpenseClassCapitalPlacement
	case IsOtherExcludedPersonalExpense(p):
		return ExpenseClassOtherExcluded
	default:
		return ""
	}
}

func expenseClassLabel(class string) string {
	switch class {
	case ExpenseClassBusiness:
		return "Business"
	case ExpenseClassTax:
		return "Tax"
	case ExpenseClassCapitalPlacement:
		return "Capital Placement"
	case ExpenseClassOtherExcluded:
		return "Other excluded flows"
	default:
		return class
	}
}

func medianDecimal(values []decimal.Decimal) decimal.Decimal {
	if len(values) == 0 {
		return decimal.Zero
	}
	sorted := append([]decimal.Decimal(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].LessThan(sorted[j]) })
	middle := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[middle]
	}
	return sorted[middle-1].Add(sorted[middle]).Div(decimal.NewFromInt(2))
}

func latestExpenseMonth(postings []posting.Posting) (time.Time, bool) {
	var latest time.Time
	for _, p := range postings {
		if !strings.HasPrefix(p.Account, "Expenses") || p.Forecast {
			continue
		}
		if latest.IsZero() || p.Date.After(latest) {
			latest = p.Date
		}
	}
	if latest.IsZero() {
		return time.Time{}, false
	}
	return time.Date(latest.Year(), latest.Month(), 1, 0, 0, 0, 0, latest.Location()), true
}

func baselineEnd(latestExpense time.Time, now time.Time) time.Time {
	nowMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	if latestExpense.Before(nowMonth) {
		return latestExpense.AddDate(0, 1, 0)
	}
	return nowMonth
}

// BuildExpenseBaseline returns six complete calendar months ending at the
// latest complete month available in the ledger. Missing months are retained.
// Personal Expense includes explicit one-offs; Regular Personal Burn does not.
func BuildExpenseBaseline(postings []posting.Posting, now time.Time) ExpenseBaseline {
	latest, ok := latestExpenseMonth(postings)
	if !ok {
		return ExpenseBaseline{Months: []ExpenseBaselineMonth{}, Excluded: []ExpenseExclusion{}, OneOffs: []ExpenseOneOff{}, CategorySignals: []ExpenseCategorySignal{}}
	}
	end := baselineEnd(latest, now)
	start := end.AddDate(0, -6, 0)
	period := Period{Start: start, End: end}
	periods := MonthPeriods(period)

	months := make([]ExpenseBaselineMonth, len(periods))
	monthIndex := make(map[string]int, len(periods))
	for i, month := range periods {
		months[i] = ExpenseBaselineMonth{Date: month.Start, PersonalExpense: decimal.Zero, RegularPersonalBurn: decimal.Zero}
		monthIndex[month.Key()] = i
	}

	totals := make([]monthTotals, len(months))
	categoryTotals := make(map[string][]decimal.Decimal)
	excluded := make(map[string]*ExpenseExclusion)
	oneOffs := make([]ExpenseOneOff, 0)

	for _, p := range postings {
		if p.Forecast || !period.Contains(p.Date) || !strings.HasPrefix(p.Account, "Expenses") {
			continue
		}
		month, found := monthIndex[p.Date.Format("2006-01")]
		if !found {
			continue
		}
		if class := expenseClass(p); class != "" {
			item := excluded[class]
			if item == nil {
				item = &ExpenseExclusion{Class: class, Label: expenseClassLabel(class), Amount: decimal.Zero}
				excluded[class] = item
			}
			item.Amount = item.Amount.Add(p.Amount)
			item.Count++
			continue
		}
		if !IsExpense(p) {
			continue
		}

		totals[month].personal = totals[month].personal.Add(p.Amount)
		if IsOneOff(p) {
			oneOffs = append(oneOffs, ExpenseOneOff{Posting: p, Category: reportingCategory(p.Account)})
			continue
		}
		totals[month].regular = totals[month].regular.Add(p.Amount)
		category := reportingCategory(p.Account)
		if categoryTotals[category] == nil {
			categoryTotals[category] = make([]decimal.Decimal, len(months))
		}
		categoryTotals[category][month] = categoryTotals[category][month].Add(p.Amount)
	}

	personalValues := make([]decimal.Decimal, len(months))
	regularValues := make([]decimal.Decimal, len(months))
	for i := range months {
		months[i].PersonalExpense = totals[i].personal
		months[i].RegularPersonalBurn = totals[i].regular
		personalValues[i] = totals[i].personal
		regularValues[i] = totals[i].regular
	}

	excludedList := make([]ExpenseExclusion, 0, len(excluded))
	for _, item := range excluded {
		excludedList = append(excludedList, *item)
	}
	sort.Slice(excludedList, func(i, j int) bool { return excludedList[i].Class < excludedList[j].Class })

	categorySignals := make([]ExpenseCategorySignal, 0)
	lastMonth := len(months) - 1
	for category, values := range categoryTotals {
		baseline := medianDecimal(values)
		current := values[lastMonth]
		delta := current.Sub(baseline)
		if !delta.GreaterThan(decimal.Zero) {
			continue
		}
		growth := decimal.Zero
		if baseline.GreaterThan(decimal.Zero) {
			growth = delta.Div(baseline).Mul(decimal.NewFromInt(100))
		}
		categorySignals = append(categorySignals, ExpenseCategorySignal{
			Category: category, Current: current, BaselineMedian: baseline, Delta: delta, GrowthPercent: growth,
		})
	}
	sort.Slice(categorySignals, func(i, j int) bool {
		if !categorySignals[i].Delta.Equal(categorySignals[j].Delta) {
			return categorySignals[i].Delta.GreaterThan(categorySignals[j].Delta)
		}
		return categorySignals[i].Category < categorySignals[j].Category
	})

	return ExpenseBaseline{
		StartDate: start, EndDate: end, Months: months,
		PersonalExpenseMedian:     medianDecimal(personalValues),
		RegularPersonalBurnMedian: medianDecimal(regularValues),
		Excluded:                  excludedList, OneOffs: oneOffs, CategorySignals: categorySignals,
	}
}

func (p Period) Key() string {
	return p.Start.Format("2006-01")
}
