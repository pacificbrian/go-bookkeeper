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

func (account *Account) ListScheduled(db *gorm.DB, canRecordOnly bool) []CashFlow {
	entries := []CashFlow{}
	if !account.Verified {
		account.Get(db, false)
	}
	if account.Verified {
		// &CashFlow{AccountID: account.ID, Type: "Repeat", Split: false})
		query := map[string]interface{}{"account_id": account.ID, "type": "Repeat", "split": false}
		if canRecordOnly {
			db.Order("date asc").Where("date <= ? AND repeat_interval_id > ?", time.Now(), 0).Find(&entries, query)
		} else {
			db.Order("date asc").Find(&entries, query)
			for i := 0; i < len(entries); i++ {
				repeat := &entries[i]
				// for #Show
				repeat.Preload(db)
			}
		}
		log.Printf("[MODEL] LIST SCHEDULED ACCOUNT(%d:%d)", account.ID, len(entries))
	}
	return entries
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
	if !c.mustUpdateBalance() {
		return
	}

	if c.oldAmount.Equal(decimal.Zero) {
		// Create, Scheduled CashFlows
		log.Printf("[MODEL] UPDATE BALANCE ACCOUNT(%d:%d): +%f",
			   a.ID, c.ID, c.Amount.InexactFloat64())
		db.Model(a).Update("Balance", gorm.Expr("balance + ?", c.Amount))
	} else {
		// Update
		newBalance := (a.Balance.Sub(c.oldAmount)).Add(c.Amount)
		if !(a.Balance.Equal(newBalance)) {
			log.Printf("[MODEL] UPDATE BALANCE ACCOUNT(%d:%d): %f -> %f",
				   a.ID, c.ID, a.Balance.InexactFloat64(),
				   newBalance.InexactFloat64())
			db.Model(a).Update("Balance", newBalance)
			a.Balance = newBalance
		}
	}
}

func (a *Account) Create(db *gorm.DB) error {
	u := GetCurrentUser()
	if u != nil {
		// Account.User is set to CurrentUser()
		a.UserID = u.ID
		spewModel(a)
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
		spewModel(a)

		// test if any ScheduledCashFlows need to post
		scheduled := a.ListScheduled(db, true)
		for i := 0; i < len(scheduled); i++ {
			repeat := &scheduled[i]
			repeat.Account.ID = a.ID
			repeat.Account.Verified = a.Verified
			repeat.tryInsertRepeatCashFlow(db)
		}
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
		spewModel(a)
		return nil
	}
	return errors.New("Permission Denied")
}

// Account access already verified with Get
func (a *Account) Update(db *gorm.DB) error {
	spewModel(a)
	result := db.Save(a)
	return result.Error
}
