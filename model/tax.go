/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"errors"
	"log"
	"time"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var FilingStatusLabels = [5]string{"","S","MFJ","MFS","HH"}
var TaxTypeNames = [9]string{"", "Income", "Income Capital Gain",
			     "Deductions for AGI", "Deductions from AGI",
			     "Itemized Deductions", "Tax", "Tax Credits",
			     "Tax Payments"}
// for ItemizedDeduction, Credits, Payments: entry.Amount will be set
// from Debit CashFlows, but we expected TaxEntry to have Positive Amount
var FlipAutomaticTaxEntries = [9]bool{false, false, false, false, false,
				      true, false, true, true}

const (
	FilingStatusUndefined uint = iota
	Single
	MarriedJointly
	MarriedSeparately
	HeadOfHousehold
)

const (
	TaxTypeUndefined uint = iota
	TaxTypeIncome
	TaxTypeIncomeCapitalGain
	TaxTypeDeductionsForAGI
	TaxTypeDeductionFromAGI
	TaxTypeItemizedDeduction
	TaxTypeTax
	TaxTypeCredits
	TaxTypePayments
)

type TaxCategory struct {
	Model
	TaxItemID uint `form:"tax_category.tax_item_id"`
	CategoryID uint `form:"tax_category.category_id"`
	TradeTypeID uint `form:"tax_category.trade_type_id"`
	TaxItem TaxItem
}

type TaxFilingStatus struct {
	Model
	Name string
	Label string
}

type TaxItem struct {
	Model
	TaxTypeID uint `form:"tax_item.tax_type_id"`
	//TaxCategoryID uint `form:"tax_item.tax_category_id"
	Name string `form:"tax_item.Name"`
	Type string
	TaxType TaxType
}

type TaxRegion struct {
	Model
	Name string `form:"tax_region.Name"`
}

type TaxType struct {
	Model
	Name string `form:"tax_type.Name"`
}

//db.Table("taxes")
type TaxEntry struct {
	Model
	Year time.Time
	DateYear int `form:"year" gorm:"-:all"`
	TaxItemID uint `form:"tax_item_id"`
	TaxRegionID uint `form:"tax_region_id"`
	TaxTypeID uint `form:"tax_type_id"`
	UserID uint
	Amount decimal.Decimal `form:"amount" gorm:"not null"`
	Memo string `form:"memo"`
	TaxItem TaxItem
	TaxRegion TaxRegion
	TaxType TaxType
	User User
}

//db.Table("tax_users")
type TaxReturn struct {
	Model
	FilingStatus uint `form:"tax_filing_status"`
	TaxRegionID uint `form:"tax_region_id"`
	UserID uint
	Year int `form:"year"`
	Exemptions int32 `form:"exemptions"`
	Income decimal.Decimal
	AgiIncome decimal.Decimal
	TaxableIncome decimal.Decimal
	ForAGI decimal.Decimal
	FromAGI decimal.Decimal
	StandardDeduction decimal.Decimal
	ItemizedDeduction decimal.Decimal
	Exemption decimal.Decimal
	Credits decimal.Decimal
	Payments decimal.Decimal
	BaseTax decimal.Decimal
	OtherTax decimal.Decimal
	OwedTax decimal.Decimal
	UnpaidTax decimal.Decimal
	LongCapgainIncome decimal.Decimal
	Session *Session `gorm:"-:all"`
	TaxRegion TaxRegion
	User User
}

func (TaxEntry) Currency(value decimal.Decimal) string {
	return currency(value)
}

func (TaxReturn) Currency(value decimal.Decimal) string {
	return currency(value)
}

// cannot get GORM to read this table using Preload,
// this is faster to just compute and avoid DB lookup
func (t TaxReturn) FilingStatusLabel() string {
	var label string

	if t.FilingStatus < uint(len(FilingStatusLabels)) {
		label = FilingStatusLabels[t.FilingStatus]
	}
	return label
}

func (*TaxCategory) List(db *gorm.DB, taxType uint) []TaxCategory {
	entries := []TaxCategory{}

	dbPrefix := db.Order("tax_item_id").Order("category_id")
	if taxType > 0 {
		dbPrefix = dbPrefix.Where("tax_type_id = ?", taxType).Joins("TaxItem")
	} else {
		dbPrefix = dbPrefix.Preload("TaxItem")
	}
	dbPrefix.Find(&entries)

	if len(entries) > 0 {
		log.Printf("[MODEL] LIST TAX CATEGORIES(%d: %d)", taxType, len(entries))
	}
	return entries
}

func (*TaxFilingStatus) List(db *gorm.DB) []TaxFilingStatus {
	entries := []TaxFilingStatus{}
	db.Table("tax_filing_status").Find(&entries)

	return entries
}

func (*TaxItem) List(db *gorm.DB) []TaxItem {
	entries := []TaxItem{}
	db.Find(&entries)

	return entries
}

func (ti *TaxItem) setTaxType() {
	switch ti.Type {
	case "TaxIncomeItem":
		ti.TaxType.ID = TaxTypeIncome
	case "TaxIncomeCapitalGainItem":
		ti.TaxType.ID = TaxTypeIncomeCapitalGain
	case "TaxDeductionForAGIItem":
		ti.TaxType.ID = TaxTypeDeductionsForAGI
	case "TaxDeductionFromAGIItem":
		ti.TaxType.ID = TaxTypeDeductionFromAGI
	case "TaxDeductionItemizedItem":
		ti.TaxType.ID = TaxTypeItemizedDeduction
	case "TaxTaxItem":
		ti.TaxType.ID = TaxTypeTax
	case "TaxCreditItem":
		ti.TaxType.ID = TaxTypeCredits
	case "TaxPaymentItem":
		ti.TaxType.ID = TaxTypePayments
	}
	ti.TaxTypeID = ti.TaxType.ID
	ti.TaxType.Name = TaxTypeNames[ti.TaxTypeID]
}

func (*TaxRegion) List(db *gorm.DB) []TaxRegion {
	entries := []TaxRegion{}
	db.Find(&entries)

	return entries
}

func (*TaxType) List(db *gorm.DB) []TaxType {
	entries := []TaxType{}
	db.Find(&entries)

	return entries
}

// some TaxTypes expect to have Positive Amount
// (see comment for FlipAutomaticTaxEntries)
func (entry *TaxEntry) setAmount(total decimal.Decimal) {
	if FlipAutomaticTaxEntries[entry.TaxTypeID] {
		entry.Amount = total.Neg()
	} else {
		entry.Amount = total
	}
}

func (taxCat *TaxCategory) makeTaxEntry(session *Session, year int, total decimal.Decimal) *TaxEntry {
	db := session.DB
	u := session.GetCurrentUser()
	entry := new(TaxEntry)
	entry.Year = yearToDate(year)
	entry.UserID = u.ID
	entry.TaxItem = taxCat.TaxItem
	entry.TaxItemID = entry.TaxItem.ID
	entry.TaxItem.setTaxType()
	entry.TaxType = entry.TaxItem.TaxType
	entry.TaxTypeID = entry.TaxType.ID
	entry.setAmount(total)

	c := new(CashFlow)
	c.CategoryID = taxCat.CategoryID
	c.Account.setSession(session)
	c.setCategoryName(db)
	entry.Memo = c.CategoryName
	entry.TaxRegion.Name = "AUTO"

	return entry
}

func (*TaxEntry) List(session *Session, year int) []TaxEntry {
	db := session.DB
	u := session.GetCurrentUser()
	autoEntries := []TaxEntry{}
	entries := []TaxEntry{}
	db.Preload("TaxRegion").
	   Preload("TaxType").
	   Preload("TaxItem").
	   Table("taxes").
	   Where("year >= ? AND year < ?", yearToDate(year), yearToDate(year+1)).
	   Where(&TaxEntry{UserID: u.ID}).Find(&entries)

	// Get "AUTO" entries
	taxCategories := new(TaxCategory).List(db, 0)
	for i := 0; i < len(taxCategories); i++ {
		taxCategory := &taxCategories[i]
		if taxCategory.CategoryID > 0 {
			_, total := u.ListTaxCategory(db, year, taxCategory)
			if !total.IsZero() {
				autoEntry := taxCategory.makeTaxEntry(session, year, total)
				autoEntries = append(autoEntries, *autoEntry)
			}
		}
	}

	log.Printf("[MODEL] LIST TAX ENTRIES(AUTO:%d, USER:%d)",
		   len(autoEntries), len(entries))
	return append(autoEntries, entries...)
}

func (t *TaxEntry) Create(session *Session) error {
	db := session.DB
	u := session.GetCurrentUser()
	if u != nil {
		t.UserID = u.ID
		t.Year = yearToDate(t.DateYear)
		spewModel(t)
		result := db.Omit(clause.Associations).Table("taxes").Create(t)
		return result.Error
	}
	return errors.New("Permission Denied")
}

func (*TaxReturn) List(session *Session, year int) []TaxReturn {
	db := session.DB
	u := session.GetCurrentUser()
	entries := []TaxReturn{}
	db.Preload("TaxRegion").
	   Table("tax_users").
	   Where(&TaxReturn{UserID: u.ID, Year: year}).Find(&entries)

	log.Printf("[MODEL] LIST TAX RETURNS(%d)", len(entries))
	return entries
}

func (t *TaxReturn) Create(session *Session) error {
	db := session.DB
	u := session.GetCurrentUser()
	if u != nil {
		t.UserID = u.ID
		spewModel(t)
		result := db.Omit(clause.Associations).Table("tax_users").Create(t)
		if result.Error != nil {
			log.Panic(result.Error)
		}
		return result.Error
	}
	return errors.New("Permission Denied")
}

func (item *TaxItem) GetByName(db *gorm.DB, name string) *TaxItem {
	item.Name = name
	db.Where(&item).First(&item)

	if item.ID == 0 {
		return nil
	}
	return item
}

// Some TaxTypes may need Neg() of the CashFlows (handled in makeTaxEntry)
// Possibly we should add Round(2) to be cautious
func (*TaxType) Sum(db *gorm.DB, r *TaxReturn, taxType uint) decimal.Decimal {
	var total decimal.Decimal

	entries := []TaxEntry{}
	t := new(TaxEntry)
	t.UserID = r.UserID
	t.TaxRegionID = r.TaxRegionID
	t.TaxTypeID = taxType
	db.Table("taxes").Where(&t).
	   Where("year >= ? AND year < ?", yearToDate(r.Year), yearToDate(r.Year+1)).
	   Find(&entries)

	for i := 0; i < len(entries); i++ {
		t = &entries[i]
		total = total.Add(t.Amount)
	}

	// Include "AUTO" entries
	taxCategories := new(TaxCategory).List(db, taxType)
	taxEntry := new(TaxEntry)
	for i := 0; i < len(taxCategories); i++ {
		taxCategory := &taxCategories[i]
		if taxCategory.CategoryID > 0 {
			_, categoryTotal := r.User.ListTaxCategory(db, r.Year, taxCategory)
			if !categoryTotal.IsZero() {
				taxEntry.TaxTypeID = taxType
				taxEntry.setAmount(categoryTotal)
				total = total.Add(taxEntry.Amount)
			}
		}
	}

	if len(entries) > 0 {
		log.Printf("[MODEL] TAX TYPE(%d) SUM(%f)", taxType, total.InexactFloat64())
	}
	return total
}

func (item *TaxItem) Sum(db *gorm.DB, r *TaxReturn, name string) decimal.Decimal {
	var total decimal.Decimal

	item = item.GetByName(db, name)
	if item == nil {
		return total
	}

	entries := []TaxEntry{}
	t := new(TaxEntry)
	t.UserID = r.UserID
	t.TaxRegionID = r.TaxRegionID
	t.TaxTypeID = item.TaxTypeID
	t.TaxItemID = item.ID
	db.Table("taxes").Where(&t).
	   Where("year >= ? AND year < ?", yearToDate(r.Year), yearToDate(r.Year+1)).
	   Find(&entries)

	for i := 0; i < len(entries); i++ {
		t = &entries[i]
		total = total.Add(t.Amount)
	}

	if len(entries) > 0 {
		log.Printf("[MODEL] TAX ITEM(%d) SUM(%d:%f)", item.ID, total.InexactFloat64())
	}
	return total
}

func (r *TaxReturn) calculate(db *gorm.DB) {
	taxYear := new(TaxYear).Get(db, r.Year)
	if taxYear == nil {
		return
	}

	r.Income = new(TaxType).Sum(db, r, TaxTypeIncome)
	// qualified dividends double-counted, so remove from Income
	qualDividends := new(TaxItem).Sum(db, r, "Qualified Dividends")
	r.Income = r.Income.Sub(qualDividends)

	r.ForAGI = new(TaxType).Sum(db, r, TaxTypeDeductionsForAGI)
	r.Credits = new(TaxType).Sum(db, r, TaxTypeCredits)
	r.Payments = new(TaxType).Sum(db, r, TaxTypePayments)
	r.OtherTax = new(TaxType).Sum(db, r, TaxTypeTax)
	r.ItemizedDeduction = new(TaxType).Sum(db, r, TaxTypeItemizedDeduction)

	if r.ItemizedDeduction.IsPositive() && taxYear.SaltMaximum > 0 {
		saltMaximum := decimal.NewFromInt32(taxYear.SaltMaximum)
		saltTotal := new(TaxItem).Sum(db, r, "State Local Income Taxes")
		saltTotal = saltTotal.Add(new(TaxItem).Sum(db, r, "Real Estate Taxes"))
		saltTotal = saltTotal.Add(new(TaxItem).Sum(db, r, "Personal Property Taxes"))
		if saltTotal.IsPositive() && saltTotal.GreaterThan(saltMaximum) {
			r.ItemizedDeduction = r.ItemizedDeduction.Sub(saltTotal)
			r.ItemizedDeduction = r.ItemizedDeduction.Add(saltMaximum)
		}
	}

	r.Exemption = decimal.NewFromInt32(r.Exemptions * taxYear.ExemptionAmount)
	r.StandardDeduction = taxYear.standardDeduction(r.FilingStatus)

	// if user provided FromAGI use it, otherwise we auto-calculate
	r.FromAGI = new(TaxType).Sum(db, r, TaxTypeDeductionFromAGI)
	if r.FromAGI.IsZero() {
		r.FromAGI = decimal.Max(r.StandardDeduction, r.ItemizedDeduction).Add(r.Exemption)
	}

	// Calculate Tax Result
	r.AgiIncome = decimal.Max(r.Income.Sub(r.ForAGI), decimal.Zero)
	r.TaxableIncome = decimal.Max(r.AgiIncome.Sub(r.FromAGI), decimal.Zero)
	r.BaseTax = taxYear.calculateTax(db, r.FilingStatus, r.TaxableIncome)
	r.OwedTax = r.BaseTax.Add(r.OtherTax).Sub(r.Credits)
	r.UnpaidTax = r.OwedTax.Sub(r.Payments)
}

func (r *TaxReturn) HaveAccessPermission(session *Session) bool {
	u := session.GetCurrentUser()
	valid := !(u == nil || u.ID != r.UserID)
	if valid {
		r.Session = session
	}
	return valid
}

func (r *TaxReturn) Get(session *Session) *TaxReturn {
	db := session.DB
	db.Table("tax_users").First(&r)
	if !r.HaveAccessPermission(session) {
		return nil
	}
	return r
}

func (r *TaxReturn) Recalculate(session *Session) error {
	r = r.Get(session)
	if r == nil {
		return errors.New("Permission Denied")
	}
	db := session.DB

	if (r.TaxRegionID == 1) {
		log.Printf("[MODEL] RECALCULATE TAX RETURN(%d: %d) REGION(%d)",
			   r.ID, r.Year, r.TaxRegionID)
		r.calculate(db)
		db.Omit(clause.Associations).Table("tax_users").Save(r)
	}
	return nil
}

func (r *TaxReturn) Delete(session *Session) error {
	r = r.Get(session)
	if r == nil {
		return errors.New("Permission Denied")
	}
	db := session.DB

	log.Printf("[MODEL] DELETE TAX RETURN(%d)", r.ID)
	db.Table("tax_users").Delete(r)
	return nil
}
