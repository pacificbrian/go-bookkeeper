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
	AccountID uint `gorm:"not null"`
	oldAccountID uint `gorm:"-:all"`
	SecurityID uint `gorm:"not null"`
	Symbol string `form:"Symbol" gorm:"-:all"`
	Date time.Time
	oldDate time.Time `gorm:"-:all"`
	TaxYear int `form:"tax_year"`
	Amount decimal.Decimal `form:"amount" gorm:"not null"`
	oldAmount decimal.Decimal `gorm:"-:all"`
	Price decimal.Decimal `form:"price"`
	Shares decimal.Decimal `form:"shares"`
	// AdjustedShares is remaining unsold shares, split adjusted
	AdjustedShares decimal.Decimal
	// Basis is accumulated (used) basis from Sells (starts at 0)
	// for Buys: remaining Basis is: Amount - Basis
	// for Sells: the Gain (Loss) then is: Amount - Basis
	Basis decimal.Decimal
	oldShares decimal.Decimal `gorm:"-:all"`
	oldBasis decimal.Decimal `gorm:"-:all"`
	Closed bool
	TradeType TradeType
	Account Account
	Security Security
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

func (t *Trade) IsSharesIn() bool {
	return TradeTypeIsSharesIn(t.TradeTypeID)
}

func (t *Trade) IsSharesOut() bool {
	return TradeTypeIsSharesOut(t.TradeTypeID)
}

func (t *Trade) IsSplit() bool {
	return TradeTypeIsSplit(t.TradeTypeID)
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
	c.oldAmount = t.oldAmount
	c.Date = t.Date
	c.applyCashFlowType()
	c.CategoryID = t.TradeTypeID
	c.PayeeID = t.SecurityID
	c.ImportID = t.ID
	return c
}

// Account Trades, Account access already verified by caller
// For Security Trades, use security.ListTrades
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
				cf.PayeeName = t.Security.Company.GetName()
				cf.CategoryName = t.TradeType.Name
				cf_entries = append(cf_entries, *cf)
			}
		}
	}
	return cf_entries
}

func (t *Trade) updateBasis(db *gorm.DB, basis decimal.Decimal, soldShares decimal.Decimal) {
	updates := make(map[string]interface{})
	if t.IsBuy() {
		if t.AdjustedShares.IsZero() {
			assert(t.Basis.IsZero(), "Trade Basis Corrupted (1)")
			t.AdjustedShares = t.Shares
		}
		t.AdjustedShares = t.AdjustedShares.Sub(soldShares)
		updates["adjusted_shares"] = t.AdjustedShares

		t.Basis = t.Basis.Add(basis)
		if t.Amount.Equal(t.Basis) {
			assert(t.AdjustedShares.IsZero(), "Trade Basis Corrupted (2)")
			updates["closed"] = 1
		}
	} else {
		t.Basis = t.Basis.Add(basis)
	}
	updates["basis"] = t.Basis
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
		buy.updateBasis(db, tg.Basis, tg.Shares)
	}

	// update Sell
	t.updateBasis(db, sellBasis, t.Shares)
}

// t is Split trade and was already tested to be Valid
func (t *Trade) recordSplit(db *gorm.DB, activeBuys []Trade) {
	// update unsold Shares in Buys that are not yet closed
	for i := 0; i < len(activeBuys); i++ {
		buy := &activeBuys[i]
		buy.AdjustedShares = buy.AdjustedShares.Mul(t.Shares)
		db.Omit(clause.Associations).Model(buy).
		   Update("AdjustedShares", buy.AdjustedShares)
	}
}

// Look up Security by symbol, creates Security if none exists
func (t *Trade) securityGetBySymbol(session *Session) *Security {
	var security *Security

	if t.Symbol != "" {
		a := &t.Account
		a.ID = t.AccountID
		// verifies Account
		security = a.securityGetBySymbol(session, t.Symbol)
		if security != nil {
			t.SecurityID = security.ID
		}
	}

	return security
}

func (t *Trade) validateInputs() error {
	// ensure monetary amounts are 2 decimal places
	t.Amount = t.Amount.Round(2)

	if t.IsSell() || t.IsBuy() {
		if t.Amount.IsZero() || t.Price.IsZero() || t.Shares.IsZero() {
			return errors.New("Invalid Trade Entered (Buy/Sell)")
		}
	} else if t.IsSharesIn() || t.IsSharesOut() {
		if t.Shares.IsZero() {  // t.Amount (is optional)
			return errors.New("Invalid Trade Entered (Shares In/Out)")
		}
	} else if t.IsSplit() {
		if !t.Amount.IsZero() || t.Shares.IsZero() {
			return errors.New("Invalid Trade Entered (Split)")
		}
	}
	return nil
}

func (t *Trade) Create(session *Session) error {
	var security *Security
	var activeBuys []Trade
	var err error
	db := session.DB

	if t.SecurityID > 0 {
		// verify access to Security
		t.Security.ID = t.SecurityID
		security = t.Security.Get(session)
	} else {
		// verifies Account, creates Security if none exists
		security = t.securityGetBySymbol(session)
	}

	if security == nil {
		return errors.New("Permission Denied")
	}
	t.AccountID = security.AccountID
	t.TaxYear = t.Date.Year()
	if t.IsBuy() {
		t.AdjustedShares = t.Shares
	}
	spewModel(t)

	err = t.validateInputs()
	if err == nil && (t.IsSell() || t.IsSplit()) {
		activeBuys, err = security.validateTrade(db, t)
	}
	if err != nil {
		return err
	}

	result := db.Omit(clause.Associations).Create(t)
	log.Printf("[MODEL] CREATE TRADE(%d) SECURITY(%d) ACCOUNT(%d) TYPE(%d)",
		   t.ID, t.SecurityID, t.AccountID, t.TradeTypeID)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	if t.IsSell() {
		t.recordGain(db, activeBuys)
	} else if t.IsSplit() {
		t.recordSplit(db, activeBuys)
	}
	security.addTrade(db, t)
	c := t.toCashFlow()
	if c != nil {
		security.Account.updateBalance(db, c)
	}
	return nil
}

// t.Account must be preloaded
func (t *Trade) HaveAccessPermission(session *Session) bool {
	u := session.GetCurrentUser()
	t.Account.Verified = !(u == nil || t.Account.ID == 0 || u.ID != t.Account.UserID)
	if t.Account.Verified {
		t.Account.User = *u
		t.Account.Session = session
	}
	return t.Account.Verified
}

func (t *Trade) postQueryInit() {
	t.oldDate = t.Date
	t.oldAmount = t.Amount
	t.oldBasis = t.Basis
	t.oldShares = t.Shares
}

// Edit, Delete, Update use Get
func (t *Trade) Get(session *Session) *Trade {
	db := session.DB
	db.Preload("TradeType").Preload("Account").
	   Preload("Security.Company").Joins("Security").
	   First(&t)

	// Verify we have access to Trade
	if !t.HaveAccessPermission(session) {
		return nil
	}

	// for Update, store old values before overwritten
	t.postQueryInit()
	return t
}

func (t *Trade) Delete(session *Session) error {
	db := session.DB
	// Verify we have access to Trade
	t = t.Get(session)
	if t != nil {
		spewModel(t)
		db.Delete(t)
		return nil
	}
	return errors.New("Permission Denied")
}

// TODO
func (t *Trade) reverseGain(db *gorm.DB) ([]Trade, error) {
	return nil, nil
}

// TODO
func (t *Trade) reverseSplit(db *gorm.DB) ([]Trade, error) {
	err := t.Security.validateSplit(db, t)
	return nil, err
}

// Trade access already verified with Get
func (t *Trade) Update(session *Session) error {
	var activeBuys []Trade
	var err error
	db := session.DB

	spewModel(t)
	err = t.validateInputs()
	if err != nil {
		return err
	}

	if t.IsSell() {
		activeBuys, err = t.reverseGain(db)
		err = errors.New("Don't yet support Updating of Sell Trades!")
	} else if t.IsBuy() && !t.oldBasis.IsZero() {
		err = errors.New("Don't yet support Updating of Partially Sold Buy Trades!")
	} else if t.IsBuy() && !t.oldShares.Equal(t.AdjustedShares) {
		err = errors.New("Don't yet support Updating of Buy Trades affected by Splits!")
	} else if t.IsSplit() {
		activeBuys, err = t.reverseSplit(db)
		err = errors.New("Don't yet support Updating of Splits!")
	} else if t.IsBuy() {
		// this becomes more complicated when/if removing above error cases
		t.AdjustedShares = t.Shares
	}
	if err != nil {
		log.Printf("[MODEL] UPDATE TRADE(%d) UNSUPPORTED: %v", t.ID, err)
		return err
	}

	result := db.Omit(clause.Associations).Save(t)
	err = result.Error
	if err == nil {
		if t.IsSell() && activeBuys != nil {
			t.recordGain(db, activeBuys)
		} else if t.IsSplit() && activeBuys != nil {
			t.recordSplit(db, activeBuys)
		}

		t.Security.updateTrade(db, t)
		c := t.toCashFlow()
		if c != nil {
			t.Account.updateBalance(db, c)
		}

		log.Printf("[MODEL] UPDATE TRADE(%d) SECURITY(%d) ACCOUNT(%d) TYPE(%d)",
			   t.ID, t.SecurityID, t.AccountID, t.TradeTypeID)
	}

	return err
}
