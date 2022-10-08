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
)

var FilingStatusLabels = [5]string{"","S","MFJ","MFS","HH"}

type TaxFilingStatus struct {
	Model
	Name string
	Label string
}

type TaxItem struct {
	Model
	TaxTypeID uint `form:"tax_item.tax_type_id"`
	TaxCategoryID uint `form:"tax_item.tax_category_id"`
	Name string `form:"tax_item.Name"`
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
	TaxItem TaxItem
	TaxRegion TaxRegion
	TaxType TaxType
	User User
	Amount decimal.Decimal `form:"amount" gorm:"not null"`
	Memo string `form:"memo"`
}

//db.Table("tax_users")
type TaxReturn struct {
	Model
	FilingStatus uint `form:"tax_filing_status"`
	TaxRegionID uint `form:"tax_region_id"`
	UserID uint
	TaxRegion TaxRegion
	User User
	Year int `form:"year"`
	Exemptions int `form:"exemptions"`
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

func (*TaxEntry) List(db *gorm.DB) []TaxEntry {
	u := GetCurrentUser()
	entries := []TaxEntry{}
	db.Preload("TaxRegion").
	   Preload("TaxType").
	   Preload("TaxItem").
	   Table("taxes").
	   Where(&TaxEntry{UserID: u.ID}).Find(&entries)

	log.Printf("[MODEL] LIST TAX ENTRIES(%d)", len(entries))
	return entries
}

func (t *TaxEntry) Create(db *gorm.DB) error {
	u := GetCurrentUser()
	if u != nil {
		t.UserID = u.ID
		t.Year = time.Date(t.DateYear, 1, 1, 0, 0, 0, 0, time.UTC)
		spewModel(t)
		result := db.Table("taxes").Create(t)
		return result.Error
	}
	return errors.New("Permission Denied")
}

func (*TaxReturn) List(db *gorm.DB) []TaxReturn {
	u := GetCurrentUser()
	entries := []TaxReturn{}
	db.Preload("TaxRegion").
	   Table("tax_users").
	   Where(&TaxReturn{UserID: u.ID}).Find(&entries)

	log.Printf("[MODEL] LIST TAX RETURNS(%d)", len(entries))
	return entries
}

func (t *TaxReturn) Create(db *gorm.DB) error {
	u := GetCurrentUser()
	if u != nil {
		t.UserID = u.ID
		spewModel(t)
		result := db.Table("tax_users").Create(t)
		if result.Error != nil {
			log.Panic(result.Error)
		}
		return result.Error
	}
	return errors.New("Permission Denied")
}
