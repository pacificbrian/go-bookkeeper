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
	Portfolio SecurityValue `gorm:"-:all"`
}

func (Account) Currency(value decimal.Decimal) string {
	return "$" + value.StringFixedBank(2)
}

func (a *Account) IsInvestment() bool {
	return a.AccountType.isType("Investment")
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
		query := map[string]interface{}{"account_id": account.ID,
					        "type": "Repeat", "split": false}
		if canRecordOnly {
			db.Order("date asc").Where("date <= ?", time.Now()).
					     Where("repeats_left > 0 OR repeats_left IS NULL").
					     Where("repeat_interval_id > 0").
					     Joins("RepeatInterval").Find(&entries, query)
		} else {
			db.Order("date asc").Where("repeat_interval_id > 0").
					     Find(&entries, query)
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

func (a *Account) securityGetBySymbol(db *gorm.DB, symbol string) *Security {
	security := new(Security)
	c := companyGetBySymbol(db, symbol)
	security.CompanyID = c.ID
	security.AccountID = a.ID
	// need Where because these are not primary keys
	db.Where(&security).First(&security)
	log.Printf("[MODEL] ACCOUNT GET SECURITY for (%s:%d)", symbol, security.ID)

	if security.ID > 0 {
		// verify Account
		security.Account.ID = security.AccountID
		account := security.Account.Get(db, false)
		if account == nil {
			return nil
		}
	} else { // security.ID == 0
		err := security.Create(db)
		if err != nil {
			return nil
		}
	}

	return security
}

func (a *Account) Init() *Account {
	a.Taxable = true
	// a.UserID unset (not needed for New)
	return a
}

// Average Balance for last 30 days prior to end date.
// We need to handle case where Account age < 30 days, but currently cannot.
func (a *Account) averageDailyBalance(db *gorm.DB, endDate time.Time) decimal.Decimal {
	var total decimal.Decimal
	var daysLeft int32 = 30
	var days int32
	var validEntries uint

	lastBalance := a.Balance
	lastTime := endDate
	thirtyDaysAgo := lastTime.AddDate(0, 0, int(-daysLeft))

	// entries will be descending order
	entries := new(CashFlow).ListByDate(db, a, &thirtyDaysAgo)

	for i := 0; i < len(entries); i++ {
		if daysLeft <= 0 {
			break
		}

		cf := &entries[i]
		lastBalance = cf.Balance
		if !(cf.Date.After(endDate)) {
			days = int32(lastTime.Sub(cf.Date).Hours()) / 24
			if days > 0 {
				if days > daysLeft {
					days = daysLeft
				}
				total = total.Add(lastBalance.Mul(decimal.NewFromInt32(days)))
				daysLeft -= days
			}
			lastTime = cf.Date
			validEntries += 1
		}
		// if no more entries[], but still days_left, this is correct
		// Balance that was in account for remaining days left, which
		// is handled outside loop (below)
		lastBalance = lastBalance.Sub(cf.Amount)
	}

	if daysLeft > 0 {
		total = total.Add(lastBalance.Mul(decimal.NewFromInt32(daysLeft)))
	}

	balance := total.DivRound(decimal.NewFromInt32(30), 2)
	log.Printf("[MODEL] ACCOUNT 30-DAY AVERAGE BALANCE (%d: $%f from %d/%d entries)",
		   a.ID, balance.InexactFloat64(), validEntries, len(entries))
	return balance
}

func (a *Account) UpdateBalance(db *gorm.DB, c *CashFlow) {
	if !c.mustUpdateBalance() {
		return
	}

	if c.oldAmount.Equal(decimal.Zero) {
		// Create, Scheduled CashFlows
		log.Printf("[MODEL] UPDATE BALANCE ACCOUNT(%d:%d): +%f",
			   a.ID, c.ID, c.Amount.InexactFloat64())
		db.Omit(clause.Associations).Model(a).
		   Update("Balance", gorm.Expr("balance + ?", c.Amount))
	} else {
		// Update
		newBalance := (a.Balance.Sub(c.oldAmount)).Add(c.Amount)
		if !(a.Balance.Equal(newBalance)) {
			log.Printf("[MODEL] UPDATE BALANCE ACCOUNT(%d:%d): %f -> %f",
				   a.ID, c.ID, a.Balance.InexactFloat64(),
				   newBalance.InexactFloat64())
			db.Omit(clause.Associations).Model(a).
			   Update("Balance", newBalance)
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
			repeat.Account.Balance = a.Balance
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
