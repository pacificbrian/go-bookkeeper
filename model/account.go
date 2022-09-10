/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"errors"
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

func accountGetByName(db *gorm.DB, name string) *Account {
	u := GetCurrentUser()
	if u == nil {
		return nil
	}

	a := new(Account)
	a.Name = name
	a.UserID = u.ID
	// need Where because these are not primary keys
	db.Where(&a).First(&a)

	if a.ID == 0 {
		return nil
	}
	return a
}

func (a *Account) Init() *Account {
	a.Taxable = true
	// a.UserID unset (not needed for New)
	return a
}

func (a *Account) UpdateBalance(db *gorm.DB, c *CashFlow) {
	a.Balance = (a.Balance.Sub(c.oldAmount)).Add(c.Amount)
	db.Model(a).Update("Balance", a.Balance)
}

func (a *Account) Create(db *gorm.DB) error {
	u := GetCurrentUser()
	if u != nil {
		// Account.User is set to CurrentUser()
		a.UserID = u.ID
		spew.Dump(a)
		result := db.Create(a)
		return result.Error
	}
	return errors.New("Permission Denied")
}

func (a *Account) HaveAccessPermission() bool {
	u := GetCurrentUser()
	// store in a.Verified if this Account is trusted
	a.Verified = !(u == nil || u.ID != a.UserID)
	return a.Verified
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

	if preload {
		spew.Dump(a)
	}
	return a
}

func (a *Account) Delete(db *gorm.DB) error {
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
		return nil
	}
	return errors.New("Permission Denied")
}

// Account access already verified with Get
func (a *Account) Update(db *gorm.DB) error {
	spew.Dump(a)
	result := db.Save(a)
	return result.Error
}
