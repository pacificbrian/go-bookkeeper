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
	AccountTypeID uint `form:"account.account_type_id"`
	AccountType AccountType
	CurrencyTypeID uint `form:"account.currency_type_id"`
	CurrencyType CurrencyType
	Name string `form:"account.Name"`
	Number string `form:"account.Number"`
	Routing int `form:"account.Routing"`
	Balance decimal.Decimal
	Taxable bool `form:"account.Taxable"`
	CashFlows []CashFlow
	//Categories []Category
	//Payees []Payee
}

func ListAccounts(db *gorm.DB) []Account {
	entries := []Account{}
	db.Find(&entries)

	return entries
}
