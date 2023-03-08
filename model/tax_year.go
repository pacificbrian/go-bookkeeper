/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"log"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TaxConstant struct {
	Model
	TaxTableMax int32
}

type TaxYear struct {
	Model
	Year int
	ExemptionAmount int32
	SaltMaximum int32
	StandardDeductionS int32
	StandardDeductionMFJ int32
	StandardDeductionMFS int32
	StandardDeductionHH int32
	TaxIncomeL1S int32
	TaxIncomeL1MFJ int32
	TaxIncomeL1MFS int32
	TaxIncomeL1HH int32
	TaxIncomeL2S int32
	TaxIncomeL2MFJ int32
	TaxIncomeL2MFS int32
	TaxIncomeL2HH int32
	TaxIncomeL3S int32
	TaxIncomeL3MFJ int32
	TaxIncomeL3MFS int32
	TaxIncomeL3HH int32
	TaxIncomeL4S int32
	TaxIncomeL4MFJ int32
	TaxIncomeL4MFS int32
	TaxIncomeL4HH int32
	TaxIncomeL5S int32
	TaxIncomeL5MFJ int32
	TaxIncomeL5MFS int32
	TaxIncomeL5HH int32
	TaxIncomeL6S int32
	TaxIncomeL6MFJ int32
	TaxIncomeL6MFS int32
	TaxIncomeL6HH int32
	TaxL1Rate decimal.Decimal
	TaxL2Rate decimal.Decimal
	TaxL3Rate decimal.Decimal
	TaxL4Rate decimal.Decimal
	TaxL5Rate decimal.Decimal
	TaxL6Rate decimal.Decimal
	TaxL7Rate decimal.Decimal
}

func (y *TaxYear) standardDeduction(filingStatus uint) decimal.Decimal {
	var deduction int32

	switch filingStatus {
	case Single:
		deduction = y.StandardDeductionS
	case MarriedJointly:
		deduction = y.StandardDeductionMFJ
	case MarriedSeparately:
		deduction = y.StandardDeductionMFS
	case HeadOfHousehold:
		deduction = y.StandardDeductionHH
	}
	return decimal.NewFromInt32(deduction)
}

func (y *TaxYear) taxIncomeL1(filingStatus uint) decimal.Decimal {
	var limit int32

	switch filingStatus {
	case Single:
		limit = y.TaxIncomeL1S
	case MarriedJointly:
		limit = y.TaxIncomeL1MFJ
	case MarriedSeparately:
		limit = y.TaxIncomeL1MFS
	case HeadOfHousehold:
		limit = y.TaxIncomeL1HH
	}
	return decimal.NewFromInt32(limit)
}

func (y *TaxYear) taxIncomeL2(filingStatus uint) decimal.Decimal {
	var limit int32

	switch filingStatus {
	case Single:
		limit = y.TaxIncomeL2S
	case MarriedJointly:
		limit = y.TaxIncomeL2MFJ
	case MarriedSeparately:
		limit = y.TaxIncomeL2MFS
	case HeadOfHousehold:
		limit = y.TaxIncomeL2HH
	}
	return decimal.NewFromInt32(limit)
}

func (y *TaxYear) taxIncomeL3(filingStatus uint) decimal.Decimal {
	var limit int32

	switch filingStatus {
	case Single:
		limit = y.TaxIncomeL3S
	case MarriedJointly:
		limit = y.TaxIncomeL3MFJ
	case MarriedSeparately:
		limit = y.TaxIncomeL3MFS
	case HeadOfHousehold:
		limit = y.TaxIncomeL3HH
	}
	return decimal.NewFromInt32(limit)
}

func (y *TaxYear) taxIncomeL4(filingStatus uint) decimal.Decimal {
	var limit int32

	switch filingStatus {
	case Single:
		limit = y.TaxIncomeL4S
	case MarriedJointly:
		limit = y.TaxIncomeL4MFJ
	case MarriedSeparately:
		limit = y.TaxIncomeL4MFS
	case HeadOfHousehold:
		limit = y.TaxIncomeL4HH
	}
	return decimal.NewFromInt32(limit)
}

func (y *TaxYear) taxIncomeL5(filingStatus uint) decimal.Decimal {
	var limit int32

	switch filingStatus {
	case Single:
		limit = y.TaxIncomeL5S
	case MarriedJointly:
		limit = y.TaxIncomeL5MFJ
	case MarriedSeparately:
		limit = y.TaxIncomeL5MFS
	case HeadOfHousehold:
		limit = y.TaxIncomeL5HH
	}
	return decimal.NewFromInt32(limit)
}

func (y *TaxYear) taxIncomeL6(filingStatus uint) decimal.Decimal {
	var limit int32

	switch filingStatus {
	case Single:
		limit = y.TaxIncomeL6S
	case MarriedJointly:
		limit = y.TaxIncomeL6MFJ
	case MarriedSeparately:
		limit = y.TaxIncomeL6MFS
	case HeadOfHousehold:
		limit = y.TaxIncomeL6HH
	}
	return decimal.NewFromInt32(limit)
}

func (c *TaxConstant) Get(db *gorm.DB) *TaxConstant {
	db.First(&c, 1)
	return c
}

func (y *TaxYear) Get(db *gorm.DB, year int) *TaxYear {
	y.Year = year
	if year != 0 {
		db.Where(&y).First(&y)
	}
	return y
}

func calculateBracket(income decimal.Decimal, incomeLimit decimal.Decimal,
		      lastBracket decimal.Decimal, taxRate decimal.Decimal) decimal.Decimal {
	if !income.GreaterThan(lastBracket) {
		// Skip if income level already reached with prior Bracket
		return decimal.Zero
	}

	if !incomeLimit.IsZero() {
		return decimal.Min(income, incomeLimit).Sub(lastBracket).Mul(taxRate)
	} else {
		// This catches if Tax Bracket not valid for this TaxYear
		if lastBracket.IsZero() {
			return decimal.Zero
		}
		// This is last Tax Bracket (no upper limit)
		return income.Sub(lastBracket).Mul(taxRate)
	}
}

func (y *TaxYear) calculateTax(db *gorm.DB, filingStatus uint,
			       income decimal.Decimal) decimal.Decimal {
	var tax decimal.Decimal

	constants := new(TaxConstant).Get(db)
	tableMax := decimal.NewFromInt32(constants.TaxTableMax)

	if income.LessThan(tableMax) {
		income50 := income.Mod(decimal.NewFromInt32(50))
		if income50.IsPositive() {
			income = income.Sub(income50)
			income = income.Add(decimal.NewFromInt32(25))
		}
	}

	if income.IsPositive() {
		taxIncomeL1 := y.taxIncomeL1(filingStatus)
		taxIncomeL2 := y.taxIncomeL2(filingStatus)
		taxIncomeL3 := y.taxIncomeL3(filingStatus)
		taxIncomeL4 := y.taxIncomeL4(filingStatus)
		taxIncomeL5 := y.taxIncomeL5(filingStatus)
		taxIncomeL6 := y.taxIncomeL6(filingStatus)

		tax = tax.Add(calculateBracket(income, taxIncomeL1, decimal.Zero, y.TaxL1Rate))
		tax = tax.Add(calculateBracket(income, taxIncomeL2, taxIncomeL1, y.TaxL2Rate))
		tax = tax.Add(calculateBracket(income, taxIncomeL3, taxIncomeL2, y.TaxL3Rate))
		tax = tax.Add(calculateBracket(income, taxIncomeL4, taxIncomeL3, y.TaxL4Rate))
		tax = tax.Add(calculateBracket(income, taxIncomeL5, taxIncomeL4, y.TaxL5Rate))
		tax = tax.Add(calculateBracket(income, taxIncomeL6, taxIncomeL5, y.TaxL6Rate))
		tax = tax.Add(calculateBracket(income, decimal.Zero, taxIncomeL6, y.TaxL7Rate))

		if income.LessThan(tableMax) {
			tax = tax.Round(0)
		} else {
			tax = tax.Round(2)
		}
	}

	log.Printf("[MODEL] CALCULATE TAX (%f) on INCOME (%f)",
		   tax.InexactFloat64(), income.InexactFloat64())
	return tax
}
