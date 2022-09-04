/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"github.com/shopspring/decimal"
	"github.com/davecgh/go-spew/spew"
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
	Verified bool `gorm:"-:all"`
	CashFlows []CashFlow
}

func (Account) Currency(value decimal.Decimal) string {
	return "$" + value.StringFixedBank(2)
}

func ListAccounts(db *gorm.DB) []Account {
	u := GetCurrentUser()
	entries := []Account{}
	if u == nil {
		return entries
	}

	// Find Accounts for CurrentUser()
	db.Where(&Account{UserID: u.ID}).Find(&entries)
	return entries
}

func (*Account) List(db *gorm.DB) []Account {
	return ListAccounts(db)
}

func (a *Account) HaveAccessPermission() bool {
	u := GetCurrentUser()
	return !(u == nil || u.ID != a.UserID)
}

func (a *Account) Init() *Account {
	a.Taxable = true
	// a.UserID unset (not needed for New)
	return a
}

func (a *Account) Create(db *gorm.DB) {
	u := GetCurrentUser()
	if u != nil {
		// Account.User is set to CurrentUser()
		a.UserID = u.ID
		spew.Dump(a)
		db.Create(a)
	}
}

// Show, Edit, Delete, Update use Get
// a.UserID unset, need to load
func (a *Account) Get(db *gorm.DB, preload bool) *Account {
	// Load and Verify we have access to Account
	if preload {
		// Get (Show)
		db.Preload("AccountType").First(&a)
	} else {
		// Edit, Delete, Update
		db.First(&a)
	}
	if !a.HaveAccessPermission() {
		return nil
	}

	// Set verified so this Account is trusted
	a.Verified = true
	if preload {
		spew.Dump(a)
	}
	return a
}

func (a *Account) Delete(db *gorm.DB) {
	// Verify we have access to Account
	a = a.Get(db, false)
	if a != nil {
		// on first delete, we only make Hidden
		if !a.Hidden {
			a.Hidden = true
			db.Save(a)
		} else {
			db.Delete(a)
		}
		spew.Dump(a)
	}
}

// Account access already verified with Get
func (a *Account) Update(db *gorm.DB) *Account {
	spew.Dump(a)
	db.Save(a)

	return a
}
