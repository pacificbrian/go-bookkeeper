/*
 * SPDX-FileCopyrightText: 2023 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"log"
	"math"
	"time"
	"github.com/pacificbrian/edgar"
	"github.com/pacificbrian/go-bookkeeper/config"
)

type FilingData edgar.Filing

const (
	FilingTypeUndefined string = ""
	FilingTypeAnnual = "10-K"
	FilingTypeQuarterly = "10-Q"
)

const (
	BalanceSheet string = "Balance Sheet"
	IncomeStatement = "Income Statement"
	CashFlowStatement = "CashFlow Statement"
	ConsolidatedStatements = "Consolidated Financials"
)

const (
	Cash uint = iota
	Investments
	CurrentAssets
	Goodwill
	Intangibles
	Assets
	ShortTermDebt
	UnearnedRevenue
	CurrentLiabilities
	LongTermDebt
	Liabilities
	RetainedEarnings
	Equity
	BasicShares
	DilutedShares
	Revenue
	CostOfRevenue
	GrossProfit
	OperatingExpense
	OperatingIncome
	Interest
	NetIncome
	OperatingCashFlow
	CapEx
	DividendPaid
)

// order of below must match above for map keys to be correct
var ConsolidatedFilingItemKeys = []string {
	// Balance Sheet
	"Cash",
	"Investments",
	"Current Assets",
	"Goodwill",
	"Intangibles",
	"Assets",
	"Short-Term Debt",
	"Unearned Revenue",
	"Current Liabilities",
	"Long-Term Debt",
	"Liabilities",
	"Retained Earnings",
	"Equity",
	// Income Statement
	"Basic Shares",
	"Diluted Shares",
	"Revenue",
	"Cost Of Revenue",
	"Gross Profit",
	"Operating Expense",
	"Operating Income",
	"Interest",
	"Net Income",
	// CashFlow
	"Operating CashFlow",
	"CapEx",
	"Dividend Paid",
}

var ConsolidatedFilingItemMap = map[string]func(edgar.Filing) (float64, error){
	// Balance Sheet
	ConsolidatedFilingItemKeys[Cash] : edgar.Filing.Cash,
	ConsolidatedFilingItemKeys[Investments] : edgar.Filing.Securities,
	ConsolidatedFilingItemKeys[CurrentAssets] : edgar.Filing.CurrentAssets,
	ConsolidatedFilingItemKeys[Goodwill] : edgar.Filing.Goodwill,
	ConsolidatedFilingItemKeys[Intangibles] : edgar.Filing.Intangibles,
	ConsolidatedFilingItemKeys[Assets] : edgar.Filing.Assets,
	ConsolidatedFilingItemKeys[ShortTermDebt] : edgar.Filing.ShortTermDebt,
	ConsolidatedFilingItemKeys[UnearnedRevenue] : edgar.Filing.DeferredRevenue,
	ConsolidatedFilingItemKeys[CurrentLiabilities] : edgar.Filing.CurrentLiabilities,
	ConsolidatedFilingItemKeys[LongTermDebt] : edgar.Filing.LongTermDebt,
	ConsolidatedFilingItemKeys[Liabilities] : edgar.Filing.Liabilities,
	ConsolidatedFilingItemKeys[RetainedEarnings] : edgar.Filing.RetainedEarnings,
	ConsolidatedFilingItemKeys[Equity] : edgar.Filing.TotalEquity,
	// Income Statement
	ConsolidatedFilingItemKeys[BasicShares] : edgar.Filing.ShareCount,
	ConsolidatedFilingItemKeys[DilutedShares] : edgar.Filing.WAShares,
	ConsolidatedFilingItemKeys[Revenue] : edgar.Filing.Revenue,
	ConsolidatedFilingItemKeys[CostOfRevenue] : edgar.Filing.CostOfRevenue,
	ConsolidatedFilingItemKeys[GrossProfit] : edgar.Filing.GrossMargin,
	ConsolidatedFilingItemKeys[OperatingExpense] : edgar.Filing.OperatingExpense,
	ConsolidatedFilingItemKeys[OperatingIncome] : edgar.Filing.OperatingIncome,
	ConsolidatedFilingItemKeys[Interest] : edgar.Filing.Interest,
	ConsolidatedFilingItemKeys[NetIncome] : edgar.Filing.NetIncome,
	// CashFlow
	ConsolidatedFilingItemKeys[OperatingCashFlow] : edgar.Filing.OperatingCashFlow,
	ConsolidatedFilingItemKeys[CapEx] : edgar.Filing.CapitalExpenditure,
	ConsolidatedFilingItemKeys[DividendPaid] : edgar.Filing.Dividend,
}

var edgarHandle edgar.FilingFetcher

func init() {
	edgarHandle = edgar.NewFilingFetcher()
}

func (c *Company) edgarFilings() edgar.CompanyFolder {
	globals := config.GlobalConfig()
	if !globals.EnableSecurityFilings || c.Symbol == "" {
		return nil
	}
	filings, err := edgarHandle.CompanyFolder(c.Symbol,
						  FilingTypeAnnual,
						  FilingTypeQuarterly)
	if err != nil {
		log.Printf("[MODEL] SET EDGAR FOLDER FAILED FOR (%s): %v",
			   c.Symbol, err)
	}
	return filings
}

// Note to caller:
//   convert floats to string with: strconv.FormatFloat(f.item(), 'f', -1, 64)
func (c *Company) GetFiling(fType string, date time.Time) edgar.Filing {
	filings := c.edgarFilings()
	if filings == nil {
		return nil
	}

	f, err := filings.Filing(edgar.FilingType(fType), date)
	if err != nil {
		log.Printf("[MODEL] GET EDGAR FILING FAILED FOR (%s) (%s, %s): %v",
			   c.Symbol, fType, timeToString(&date), err)
	}
	return f
}

func (c *Company) GetFilingDates(fType string) []time.Time {
	switch fType {
	case FilingTypeAnnual:
		break
	case FilingTypeQuarterly:
		break
	default:
		return nil
	}

	filings := c.edgarFilings()
	if filings == nil {
		return nil
	}
	return filings.AvailableFilings(edgar.FilingType(fType))
}

func (c *Company) HasFilings() bool {
	return c.edgarFilings() != nil
}

func (c Company) NumFilings(fType string) int {
	if fType == "" {
		return len(c.GetFilingDates("10-Q")) +
		       len(c.GetFilingDates("10-K"))
	}
	return len(c.GetFilingDates(fType))
}

func (c *Company) GetFilingDate(f FilingData) string {
	t := f.FiledOn()
	return dateToString(&t)
}

func (c *Company) GetFilingItemNames(viewName string) []string {
	// can't use below as maps don't store insertion order
	//keys := make([]string, 0, len(ConsolidatedFilingItemMap))
	//for k := range ConsolidatedFilingItemMap {
	return ConsolidatedFilingItemKeys
}

func (c *Company) GetFilingItem(f FilingData, item string) float64 {
	queryFunc := ConsolidatedFilingItemMap[item]
	data, _ := queryFunc(f)
	return data
}

func (c *Company) GetFilingItemString(f FilingData, item string) string {
	p := message.NewPrinter(language.English)
	applyScaling := true
	suffix := ""

	data := int64(math.Round(c.GetFilingItem(f, item)))
	if applyScaling {
		switch f.ScaleMoney() {
		case 1000000000:
			data = data / f.ScaleMoney()
			suffix = " B"
		case 1000000:
			data = data / f.ScaleMoney()
			suffix = " M"
		case 1000:
			data = data / f.ScaleMoney()
			suffix = " K"
		}
	}

	strData := p.Sprintf("%d", data) // add commas
	switch item {
	case ConsolidatedFilingItemKeys[BasicShares]:
		fallthrough
	case ConsolidatedFilingItemKeys[DilutedShares]:
		return strData + suffix
	}
	return "$" + strData + suffix
}
