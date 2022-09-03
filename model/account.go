/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Account struct {
	gorm.Model
	UserID uint `gorm:"not null"`
	User User
	AccountTypeID uint `form:"account.account_type_id"`
	AccountType AccountType
	CurrencyTypeID uint `form:"account.currency_type_id"`
	CurrencyType CurrencyType
	Name string `form:"account.Name"`
	Number string `form:"account.Number"`
	Routing int `form:"account.Routing"`
	Balance decimal.Decimal
	Taxable bool `form:"account.Taxable"`
	Hidden bool `form:"account.Hidden"`
	CashFlows []CashFlow
}

func (Account) Currency(value decimal.Decimal) string {
	return "$" + value.StringFixedBank(2)
}

func ListAccounts(db *gorm.DB) []Account {
	entries := []Account{}
	db.Find(&entries)
	return entries
}

func (*Account) List(db *gorm.DB) []Account {
	return ListAccounts(db)
}

func (a *Account) Delete(db *gorm.DB) {
	if a.Hidden {
		db.Delete(a)
	} else {
		a.Hidden = true
		db.Save(a)
	}
}
