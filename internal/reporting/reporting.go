// Package reporting contains read-only adapters for the account-path rules
// already used by Paisa's reporting services.
//
// The package deliberately has no persistence model. Entries retain the
// original posting and its source coordinates so a later screen can explain a
// number without reconstructing a parallel financial model.
package reporting

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/shopspring/decimal"
)

// Period uses an inclusive start and an exclusive end.
type Period struct {
	Start time.Time
	End   time.Time
}

func NewPeriod(start, end time.Time) (Period, error) {
	if end.Before(start) {
		return Period{}, errors.New("reporting period end must not be before start")
	}
	return Period{Start: start, End: end}, nil
}

func (p Period) Contains(date time.Time) bool {
	return !date.Before(p.Start) && date.Before(p.End)
}

// MonthPeriods returns calendar-month buckets clipped to the requested
// period. Empty months are retained so timelines remain stable for charts.
func MonthPeriods(p Period) []Period {
	if p.Start.Equal(p.End) || p.Start.After(p.End) {
		return []Period{}
	}

	periods := make([]Period, 0)
	monthStart := time.Date(p.Start.Year(), p.Start.Month(), 1, 0, 0, 0, 0, p.Start.Location())
	for monthStart.Before(p.End) {
		monthEnd := monthStart.AddDate(0, 1, 0)
		start := monthStart
		if p.Start.After(start) {
			start = p.Start
		}
		end := monthEnd
		if p.End.Before(end) {
			end = p.End
		}
		periods = append(periods, Period{Start: start, End: end})
		monthStart = monthEnd
	}
	return periods
}

type Report string

const (
	IncomeReport              Report = "income"
	ExpenseReport             Report = "expenses"
	TaxReport                 Report = "tax"
	PassiveIncomeReport       Report = "passive_income"
	RegularPersonalBurnReport Report = "regular_personal_burn"
	CapitalGainsReport        Report = "capital_gains"
	InternalTransfersReport   Report = "internal_transfers"
	RefundsReport             Report = "refunds"
	AssetMovementsReport      Report = "asset_movements"
)

type Kind string

const (
	IncomeKind           Kind = "income"
	ExpenseKind          Kind = "expense"
	TaxKind              Kind = "tax"
	InternalTransferKind Kind = "internal_transfer"
	AssetMovementKind    Kind = "asset_movement"
	OtherKind            Kind = "other"
)

type PassiveIncomeCategory string

const (
	PassiveInterestCategory  PassiveIncomeCategory = "interest"
	PassiveCouponsCategory   PassiveIncomeCategory = "coupons"
	PassiveDividendsCategory PassiveIncomeCategory = "dividends"
)

// Provenance identifies the source posting without copying or normalizing its
// account path away. Posting contains the complete original source record.
type Provenance struct {
	PostingID            uint   `json:"posting_id"`
	AccountPath          string `json:"account_path"`
	TransactionID        string `json:"transaction_id"`
	FileName             string `json:"file_name"`
	TransactionBeginLine uint64 `json:"transaction_begin_line"`
	TransactionEndLine   uint64 `json:"transaction_end_line"`
}

type Entry struct {
	Posting       posting.Posting `json:"posting"`
	Provenance    Provenance      `json:"provenance"`
	Kind          Kind            `json:"kind"`
	PassiveIncome bool            `json:"passive_income"`
	CapitalGain   bool            `json:"capital_gain"`
	Refund        bool            `json:"refund"`
}

type Options struct {
	// Query.All excludes forecast postings by default. The seam keeps that
	// compatibility default and only includes them when explicitly requested.
	IncludeForecast bool
}

type Predicate func(posting.Posting) bool

func AccountOrChild(account, root string) bool {
	return utils.IsSameOrParent(account, root)
}

func IsIncome(p posting.Posting) bool {
	return AccountOrChild(p.Account, "Income")
}

func IsTax(p posting.Posting) bool {
	return AccountOrChild(p.Account, "Expenses:Tax")
}

// IsExpense matches the existing Expenses report: tax is exposed separately.
func IsExpense(p posting.Posting) bool {
	return AccountOrChild(p.Account, "Expenses") && !IsTax(p)
}

func IsInterest(p posting.Posting) bool {
	return AccountOrChild(p.Account, "Income:Interest") && !IsCoupon(p)
}

func IsCoupon(p posting.Posting) bool {
	return AccountOrChild(p.Account, "Income:Coupon") ||
		AccountOrChild(p.Account, "Income:Interest:Investments:Coupons")
}

func IsDividend(p posting.Posting) bool {
	return AccountOrChild(p.Account, "Income:Dividend")
}

func PassiveIncomeCategoryFor(p posting.Posting) (PassiveIncomeCategory, bool) {
	switch {
	case IsCoupon(p):
		return PassiveCouponsCategory, true
	case IsInterest(p):
		return PassiveInterestCategory, true
	case IsDividend(p):
		return PassiveDividendsCategory, true
	default:
		return "", false
	}
}

func IsPassiveIncome(p posting.Posting) bool {
	_, ok := PassiveIncomeCategoryFor(p)
	return ok
}

var regularPersonalBurnExcludedRoots = []string{
	"Expenses:Business",
	"Expenses:Investments",
	"Expenses:Tax",
	"Expenses:Мошенничество",
	"Expenses:Наличные:Неразобранное",
	"Expenses:Семейный бизнес",
}

func hasOneOffBurnClass(p posting.Posting) bool {
	const marker = "burn-class: one-off"
	return strings.Contains(p.Note, marker) || strings.Contains(p.TransactionNote, marker)
}

// IsRegularPersonalBurn follows the existing household-burn boundary: all
// Expenses postings except explicitly excluded business, investment, tax,
// fraud, unclassified-cash, and family-business branches. A posting marked
// burn-class: one-off is excluded only from this report.
func IsRegularPersonalBurn(p posting.Posting) bool {
	if !AccountOrChild(p.Account, "Expenses") || hasOneOffBurnClass(p) {
		return false
	}
	for _, root := range regularPersonalBurnExcludedRoots {
		if AccountOrChild(p.Account, root) {
			return false
		}
	}
	return true
}

func IsCapitalGain(p posting.Posting) bool {
	return service.IsCapitalGains(p)
}

func IsInternalTransfer(p posting.Posting) bool {
	return AccountOrChild(p.Account, "Equity:Transfers")
}

func IsRefund(p posting.Posting) bool {
	return service.IsRefund(p)
}

func IsAssetMovement(p posting.Posting) bool {
	return AccountOrChild(p.Account, "Assets")
}

func Matches(report Report, p posting.Posting) bool {
	switch report {
	case IncomeReport:
		return IsIncome(p)
	case ExpenseReport:
		return IsExpense(p)
	case TaxReport:
		return IsTax(p)
	case PassiveIncomeReport:
		return IsPassiveIncome(p)
	case RegularPersonalBurnReport:
		return IsRegularPersonalBurn(p)
	case CapitalGainsReport:
		return IsCapitalGain(p)
	case InternalTransfersReport:
		return IsInternalTransfer(p)
	case RefundsReport:
		return IsRefund(p)
	case AssetMovementsReport:
		return IsAssetMovement(p)
	default:
		return false
	}
}

func kindFor(p posting.Posting) Kind {
	switch {
	case IsIncome(p):
		return IncomeKind
	case IsTax(p):
		return TaxKind
	case IsExpense(p):
		return ExpenseKind
	case IsInternalTransfer(p):
		return InternalTransferKind
	case IsAssetMovement(p):
		return AssetMovementKind
	default:
		return OtherKind
	}
}

func entryFor(p posting.Posting) Entry {
	return Entry{
		Posting: p,
		Provenance: Provenance{
			PostingID:            p.ID,
			AccountPath:          p.Account,
			TransactionID:        p.TransactionID,
			FileName:             p.FileName,
			TransactionBeginLine: p.TransactionBeginLine,
			TransactionEndLine:   p.TransactionEndLine,
		},
		Kind:          kindFor(p),
		PassiveIncome: IsPassiveIncome(p),
		CapitalGain:   IsCapitalGain(p),
		Refund:        IsRefund(p),
	}
}

// Select filters source postings without mutating them. The returned Entry
// always carries the original posting and source provenance.
func Select(postings []posting.Posting, period Period, options Options, predicate Predicate) []Entry {
	entries := make([]Entry, 0)
	for _, p := range postings {
		if !options.IncludeForecast && p.Forecast {
			continue
		}
		if !period.Contains(p.Date) || (predicate != nil && !predicate(p)) {
			continue
		}
		entries = append(entries, entryFor(p))
	}

	sort.SliceStable(entries, func(i, j int) bool {
		left, right := entries[i].Posting, entries[j].Posting
		if !left.Date.Equal(right.Date) {
			return left.Date.Before(right.Date)
		}
		if !left.Amount.Equal(right.Amount) {
			return left.Amount.GreaterThan(right.Amount)
		}
		if left.Account != right.Account {
			return left.Account < right.Account
		}
		return left.ID < right.ID
	})
	return entries
}

func SelectReport(postings []posting.Posting, period Period, options Options, report Report) []Entry {
	return Select(postings, period, options, func(p posting.Posting) bool {
		return Matches(report, p)
	})
}

type Month struct {
	Period  Period          `json:"period"`
	Key     string          `json:"key"`
	Entries []Entry         `json:"entries"`
	Amount  decimal.Decimal `json:"amount"`
}

// AggregateMonthly groups already-selected entries by calendar month. Amount
// keeps the Ledger sign: income postings are negative and expense postings are
// positive, matching existing server calculations.
func AggregateMonthly(entries []Entry, period Period) []Month {
	periods := MonthPeriods(period)
	months := make([]Month, 0, len(periods))
	for _, monthPeriod := range periods {
		month := Month{
			Period:  monthPeriod,
			Key:     monthPeriod.Start.Format("2006-01"),
			Entries: make([]Entry, 0),
			Amount:  decimal.Zero,
		}
		for _, entry := range entries {
			if monthPeriod.Contains(entry.Posting.Date) {
				month.Entries = append(month.Entries, entry)
				month.Amount = month.Amount.Add(entry.Posting.Amount)
			}
		}
		months = append(months, month)
	}
	return months
}

func AggregateReportMonthly(postings []posting.Posting, period Period, options Options, report Report) []Month {
	return AggregateMonthly(SelectReport(postings, period, options, report), period)
}
