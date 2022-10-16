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

type Trade struct {
	gorm.Model
	TradeTypeID uint `form:"trade_type_id"`
	TradeType TradeType
	AccountID uint `gorm:"not null"`
	oldAccountID uint `gorm:"-:all"`
	Account Account
	SecurityID uint `gorm:"not null"`
	Security Security
	Symbol string `form:"Symbol" gorm:"-:all"`
	Date time.Time
	TaxYear int `form:"tax_year"`
	Amount decimal.Decimal `form:"amount" gorm:"not null"`
	oldAmount decimal.Decimal `gorm:"-:all"`
	Price decimal.Decimal `form:"price"`
	Shares decimal.Decimal `form:"shares"`
	AdjustedShares decimal.Decimal
	// Basis is accumulated (used) basis from Sells (starts at 0)
	// for Buys: remaining Basis is: Amount - Basis
	// for Sells: the Gain (Loss) then is: Amount - Basis
	Basis decimal.Decimal
	Closed bool
}

func (Trade) Currency(value decimal.Decimal) string {
	return currency(value)
}

func (t *Trade) IsBuy() bool {
	return TradeTypeIsBuy(t.TradeTypeID)
}

func (t *Trade) IsSell() bool {
	return TradeTypeIsSell(t.TradeTypeID)
}

func (t Trade) GetBasis() string {
	if t.IsSell() {
		return "$" + t.Basis.StringFixed(2)
	} else if t.IsBuy() {
		return "$" + t.Amount.Sub(t.Basis).StringFixed(2)
	} else {
		return ""
	}
}

func (t *Trade) getCashFlowType() uint {
	var cType uint

	switch t.TradeTypeID {
	case Buy:
		cType = Debit
	case Sell:
	case Dividend:
	case Distribution:
		cType = Credit
	default:
		cType = 0
	}

	return cType
}

func (t *Trade) toCashFlow() *CashFlow {
	cType := t.getCashFlowType()
	if cType == 0 {
		return nil
	}

	c := new(CashFlow)
	c.Type = "TradeCashFlow"
	c.AccountID = t.AccountID
	c.CashFlowTypeID = cType
	c.Amount = t.Amount
	c.Date = t.Date
	c.applyCashFlowType()
	c.CategoryID = t.TradeTypeID
	c.PayeeID = t.SecurityID
	return c
}

// Account access already verified by caller
func (*Trade) List(db *gorm.DB, account *Account) []Trade {
	entries := []Trade{}
	if account.Verified {
		// Find Trades for Account
		db.Preload("TradeType").
		   Order("date asc").
		   Where(&Trade{AccountID: account.ID}).Find(&entries)
		log.Printf("[MODEL] LIST TRADES ACCOUNT(%d:%d)", account.ID, len(entries))
	}
	return entries
}

// Account access already verified by caller
func (*Trade) ListCashFlows(db *gorm.DB, account *Account) []CashFlow {
	entries := []Trade{}
	cf_entries := []CashFlow{}

	if account.Verified {
		// Find Trades for Account
		db.Preload("TradeType").
		   Preload("Security.Company").
		   Order("date desc").
		   Joins("Security").
		   Where(TradeTypeCashFlowsQuery).
		   Where(&Trade{AccountID: account.ID}).Find(&entries)
		log.Printf("[MODEL] LIST TRADES ACCOUNT(%d:%d)", account.ID, len(entries))

		for i := 0; i < len(entries); i++ {
			t := entries[i]
			cf := t.toCashFlow()
			if cf != nil {
				cf.PayeeName = t.Security.Company.CompanyName()
				cf.CategoryName = t.TradeType.Name
				cf_entries = append(cf_entries, *cf)
			}
		}
	}
	return cf_entries
}

func (t *Trade) updateBasis(db *gorm.DB, basis decimal.Decimal) {
	t.Basis = t.Basis.Add(basis)
	updates := map[string]interface{}{"basis": t.Basis}
	if t.IsBuy() && t.Amount.Equal(t.Basis) {
		updates["closed"] = 1
	}
	db.Omit(clause.Associations).Model(t).Updates(updates)
}

// t is Sell trade and was already tested to be Valid
func (t *Trade) recordGain(db *gorm.DB, activeBuys []Trade) {
	var sellBasis decimal.Decimal
	sharesRemain := t.Shares

	tg := new(TradeGain)
	for i := 0; sharesRemain.IsPositive(); i++ {
		buy := &activeBuys[i]
		tg.ID = 0
		tg.recordGain(db, t, buy, sharesRemain)
		sharesRemain = sharesRemain.Sub(tg.Shares)
		sellBasis = sellBasis.Add(tg.Basis)

		// update Basis in Buy
		buy.updateBasis(db, tg.Basis)
	}

	// update Sell
	t.updateBasis(db, sellBasis)
}

// Look up Security by symbol, creates Security if none exists
func (t *Trade) securityGetBySymbol(db *gorm.DB) *Security {
	var security *Security

	if t.Symbol != "" {
		a := &t.Account
		a.ID = t.AccountID
		// verifies Account
		security = a.securityGetBySymbol(db, t.Symbol)
		if security != nil {
			t.SecurityID = security.ID
		}
	}

	return security
}

func (t *Trade) Create(db *gorm.DB) error {
	var security *Security
	var activeBuys []Trade
	var err error

	if t.SecurityID > 0 {
		// verify access to Security
		t.Security.ID = t.SecurityID
		security = t.Security.Get(db)
	} else {
		// verifies Account, creates Security if none exists
		security = t.securityGetBySymbol(db)
	}

	if security == nil {
		return errors.New("Permission Denied")
	}
	t.AccountID = security.AccountID
	spewModel(t)

	if t.IsSell() {
		activeBuys, err = security.validateSell(db, t)
		if err != nil {
			return err
		}
	}

	result := db.Omit(clause.Associations).Create(t)
	log.Printf("[MODEL] CREATE TRADE(%d) SECURITY(%d) ACCOUNT(%d) TYPE(%d)",
		   t.ID, t.SecurityID, t.AccountID, t.TradeTypeID)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	if t.IsSell() {
		t.recordGain(db, activeBuys)
	}
	security.addTrade(db, t)
	c := t.toCashFlow()
	if c != nil {
		security.Account.updateBalance(db, c)
	}
	return nil
}

// t.Account must be preloaded
func (t *Trade) HaveAccessPermission() bool {
	u := GetCurrentUser()
	t.Account.Verified = !(u == nil || t.Account.ID == 0 || u.ID != t.Account.UserID)
	if t.Account.Verified {
		t.Account.User = *u
	}
	return t.Account.Verified
}

// Edit, Delete, Update use Get
func (t *Trade) Get(db *gorm.DB) *Trade {
	db.Preload("Account").First(&t)
	// Verify we have access to Trade
	if !t.HaveAccessPermission() {
		return nil
	}
	return t
}

func (t *Trade) Delete(db *gorm.DB) error {
	// Verify we have access to Trade
	t = t.Get(db)
	if t != nil {
		spewModel(t)
		db.Delete(t)
		return nil
	}
	return errors.New("Permission Denied")
}

// Trade access already verified with Get
func (t *Trade) Update(db *gorm.DB) error {
	spewModel(t)
	result := db.Omit(clause.Associations).Save(t)
	return result.Error
}
