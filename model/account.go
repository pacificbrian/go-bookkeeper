/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"errors"
	"log"
	"strconv"
	"time"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Account struct {
	gorm.Model
	UserID uint `gorm:"not null"`
	AccountTypeID uint `form:"account.account_type_id"`
	CurrencyTypeID uint `form:"account.currency_type_id"`
	AverageBalance decimal.Decimal `gorm:"-:all"`
	Balance decimal.Decimal
	Portfolio SecurityValue `gorm:"-:all"`
	Routing int `form:"account.Routing"`
	Name string `form:"account.Name"`
	Number string `form:"account.Number"`
	HasScheduled bool
	Hidden bool `form:"account.Hidden"`
	Taxable bool `form:"account.Taxable"`
	Verified bool `gorm:"-:all"`
	Session *Session `gorm:"-:all"`
	AccountType AccountType
	CurrencyType CurrencyType
	User User
	CashFlows []CashFlow
}

func (Account) Currency(value decimal.Decimal) string {
	return "$" + value.StringFixedBank(2)
}

func (a *Account) GetRouting() string {
	if a.Routing == 0 {
		return ""
	}
	return strconv.Itoa(a.Routing)
}

func (a *Account) IsInvestment() bool {
	return a.AccountType.isType("Investment")
}

func (a *Account) TotalPortfolio(securities []Security) {
	for i := 0; i < len(securities); i++ {
		security := &securities[i]
		a.Portfolio.Basis = a.Portfolio.Basis.Add(security.Basis)
		a.Portfolio.Value = a.Portfolio.Value.Add(security.Value)
	}
}

func cacheAccounts(u *User, accounts []Account) {
	for i := 0; i < len(accounts); i++ {
		u.cacheAccount(&accounts[i])
	}
}

func ListAccounts(session *Session, all bool) []Account {
	db := session.DB
	u := session.GetCurrentUser()
	entries := []Account{}
	hidden_clause := ""
	if u == nil {
		return entries
	}

	if !all {
		hidden_clause = "(hidden = 0 OR hidden IS NULL)"
	}

	// Find Accounts for CurrentUser()
	db.Preload("AccountType").
	   Order("account_type_id").Order("Name").
	   Where(hidden_clause).
	   Where(&Account{UserID: u.ID}).Find(&entries)
	log.Printf("[MODEL] LIST ACCOUNTS(%d)", len(entries))

	cacheAccounts(u, entries)
	return entries
}

func (*Account) List(session *Session, all bool) []Account {
	return ListAccounts(session, all)
}

func (account *Account) ListImports(session *Session) []Import {
	entries := []Import{}
	if !account.Verified {
		account.Get(session, false)
	}
	return entries
}

func (account *Account) ListScheduled(session *Session, canRecordOnly bool) []CashFlow {
	entries := []CashFlow{}
	if !account.Verified {
		account.Get(session, false)
	}

	if account.Verified {
		if !account.HasScheduled {
			return entries
		}

		db := session.DB
		query := map[string]interface{}{"account_id": account.ID,
					        "type": "RCashFlow", "split": false}
		if canRecordOnly {
			db.Order("date asc").Preload("RepeatInterval.RepeatIntervalType").
					     Where("date <= ?", time.Now()).
					     Where("repeats_left > 0 OR repeats_left IS NULL").
					     Where("cash_flow_id > 0").
					     Where("repeat_interval_id > 0").
					     Joins("RepeatInterval").Find(&entries, query)
		} else {
			db.Order("date asc").Preload("RepeatInterval.RepeatIntervalType").
					     Preload("Payee").
					     Where("repeat_interval_id > 0").
					     Joins("RepeatInterval").Find(&entries, query)
			for i := 0; i < len(entries); i++ {
				repeat := &entries[i]
				// for Preload access to Account.User.Cache
				repeat.Account.cloneVerified(account)
				// for #Show
				repeat.Preload(db)
			}
		}
		log.Printf("[MODEL] LIST SCHEDULED ACCOUNT(%d:%d) (%t)",
			   account.ID, len(entries), canRecordOnly)
	}
	return entries
}

func accountGetByName(session *Session, name string) *Account {
	db := session.DB
	u := session.GetCurrentUser()
	if u == nil {
		return nil
	}

	a := new(Account)
	a.Name = name
	a.UserID = u.ID
	// need Where because these are not primary keys
	db.Where(&a).First(&a)

	if a.ID == 0 || !a.HaveAccessPermission(session) {
		return nil
	}
	return a
}

func (a *Account) securityGetBySymbol(session *Session, symbol string) *Security {
	security := new(Security)
	db := session.DB
	c := companyGetBySymbol(db, symbol)
	security.CompanyID = c.ID
	security.AccountID = a.ID
	// need Where because these are not primary keys
	db.Where(&security).First(&security)
	log.Printf("[MODEL] ACCOUNT GET SECURITY for (%s:%d)", symbol, security.ID)

	if security.ID > 0 {
		// verify Account
		security.Account.ID = security.AccountID
		account := security.Account.Get(session, false)
		if account == nil {
			return nil
		}
	} else { // security.ID == 0
		err := security.Create(session)
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

// Average Balance for last 30 days prior to end date; uses/requires a.Balance.
// If Account is less than 30 days old, this will add Zeros for those days which
// gives the correct behavior for Interest calculations.
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
			days = durationDays(lastTime.Sub(cf.Date))
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
		balanceBeforeFirstCashFlow := lastBalance.Mul(decimal.NewFromInt32(daysLeft))
		log.Printf("[MODEL] ACCOUNT(%d) BALANCE ACTIVE ($%f) IDLE ($%f DAYS:%d)",
			   a.ID, total.InexactFloat64(),
			   balanceBeforeFirstCashFlow.InexactFloat64(), daysLeft)
		total = total.Add(balanceBeforeFirstCashFlow)
	}

	balance := total.DivRound(decimal.NewFromInt32(30), 2)
	log.Printf("[MODEL] ACCOUNT(%d) 30-DAY AVERAGE BALANCE ($%f from %d/%d entries)",
		   a.ID, balance.InexactFloat64(), validEntries, len(entries))
	return balance
}

func (a Account) HasAverageDailyBalance() bool {
	return !a.AverageBalance.IsZero()
}

func (a *Account) SetAverageDailyBalance(session *Session) {
	a.AverageBalance = a.averageDailyBalance(session.DB, time.Now())
}

func (a *Account) addScheduled(db *gorm.DB) {
	db.Omit(clause.Associations).Model(a).Update("HasScheduled", 1)
}

func (a *Account) updateBalance(db *gorm.DB, c *CashFlow) {
	if !c.mustUpdateBalance() {
		return
	}

	if c.oldAmount.IsZero() || a.Balance.IsZero() {
		// Create, Scheduled CashFlows
		// Put Update (Transfer) (don't know Account Balance)
		adjustAmount := c.Amount.Sub(c.oldAmount)
		log.Printf("[MODEL] UPDATE BALANCE ACCOUNT(%d:%d): +%f",
			   a.ID, c.ID, adjustAmount.InexactFloat64())
		db.Omit(clause.Associations).Model(a).
		   Update("Balance", gorm.Expr("balance + ?", adjustAmount))
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

func (a *Account) Create(session *Session) error {
	db := session.DB
	u := session.GetCurrentUser()
	if u != nil {
		// Account.User is set to CurrentUser()
		a.UserID = u.ID
		spewModel(a)
		result := db.Omit(clause.Associations).Create(a)
		return result.Error
	}
	return errors.New("Permission Denied")
}

func (a *Account) cloneVerified(src *Account) {
	a.ID = src.ID
	a.User.ID = src.User.ID
	a.User.Cache = src.User.Cache
	a.User.UserSettings = src.User.UserSettings
	a.Session = src.Session
	a.Balance = src.Balance
	a.Verified = src.Verified
}

func (a *Account) HaveAccessPermission(session *Session) bool {
	u := session.GetCurrentUser()
	// store in a.Verified if this Account is trusted
	a.Verified = !(u == nil || u.ID != a.UserID)
	if a.Verified {
		a.User = *u
		a.Session = session
	}
	return a.Verified
}

// Show, Edit, Delete, Update use Get
// a.UserID unset, need to load
func (a *Account) Get(session *Session, preload bool) *Account {
	db := session.DB
	enableScheduledCashFlow := true

	// Load and Verify we have access to Account
	if preload {
		// Get (Show)
		db.Preload("AccountType").First(&a)
	} else {
		// Edit, Delete, Update
		db.First(&a)
	}
	if !a.HaveAccessPermission(session) {
		return nil
	}

	if preload && enableScheduledCashFlow {
		spewModel(a)

		// test if any ScheduledCashFlows need to post
		scheduled := a.ListScheduled(session, true)
		for i := 0; i < len(scheduled); i++ {
			var repeat *CashFlow = &scheduled[i]
			repeat.Account.cloneVerified(a)
			repeat.tryInsertRepeatCashFlow(db)
		}
	}
	return a
}

func (a *Account) Delete(session *Session) error {
	// Verify we have access to Account
	a = a.Get(session, false)
	if a == nil {
		return errors.New("Permission Denied")
	}
	db := session.DB

	// on first delete, we only make Hidden
	if !a.Hidden {
		a.Hidden = true
		db.Omit(clause.Associations).Save(a)
	} else {
		count := new(CashFlow).Count(db, a)
		log.Printf("[MODEL] DELETE ACCOUNT(%d) IF (%d == 0)", a.ID, count)
		if count == 0 {
			db.Delete(a)
		}
	}
	spewModel(a)
	return nil
}

// Account access already verified with Get
func (a *Account) Update(session *Session) error {
	db := session.DB
	spewModel(a)
	result := db.Omit(clause.Associations).Save(a)
	return result.Error
}
