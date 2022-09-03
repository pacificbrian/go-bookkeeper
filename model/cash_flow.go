/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"time"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type CashFlow struct {
	gorm.Model
	CashFlowTypeID uint `form:"cash_flow_type_id" gorm:"-:all"`
	AccountID uint `gorm:"not null"`
	Account Account
	Date time.Time
	Amount decimal.Decimal `form:"amount" gorm:"not null"`
	Balance decimal.Decimal `gorm:"-:all"`
	Split bool
	Transfer bool
	Transnum string `form:"transnum"`
	Memo string `form:"memo"`
	PayeeID uint `gorm:"not null"`
	Payee Payee
	PayeeName string `form:"payee_name" gorm:"-:all"`
	CategoryID uint `form:"category_id"`
	Category Category
}

func (CashFlow) Currency(value decimal.Decimal) string {
	return  "$" + value.StringFixedBank(2)
}

func (*CashFlow) List(db *gorm.DB, account *Account) []CashFlow {
	// Verify account is for current User
	entries := []CashFlow{}
	db.Find(&entries, &CashFlow{AccountID: account.ID})
	return entries
}
