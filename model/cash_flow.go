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
	AccountID uint
	Account Account
	Date time.Time `form:"date"`
	Amount decimal.Decimal `form:"amount"`
	Balance decimal.Decimal `gorm:"-:all"`
	Split bool
	Transfer bool
	Transnum string `form:"transnum"`
	Memo string `form:"memo"`
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
