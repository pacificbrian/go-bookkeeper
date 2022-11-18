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

type SecurityValue struct {
	Basis decimal.Decimal `form:"Basis"`
	Shares decimal.Decimal
	Value decimal.Decimal
}

type Security struct {
	Model
	CompanyID uint `gorm:"not null"`
	SecurityBasisTypeID uint `form:"security_basis_type_id"`
	SecurityTypeID uint `form:"security_type_id"`
	AccountID uint `gorm:"not null"`
	ImportName string `form:"ImportName"`
	SecurityValue
	lastQuoteUpdate time.Time
	Account Account
	Company Company
	SecurityType SecurityType
}

func (Security) Currency(value decimal.Decimal) string {
	return currency(value)
}

func (s Security) Price() decimal.Decimal {
	if s.Shares.Equal(decimal.Zero) {
		return decimal.Zero
	} else {
		return s.Value.DivRound(s.Shares, 2)
	}
}

func (s Security) BasisPrice() decimal.Decimal {
	if s.Shares.Equal(decimal.Zero) {
		return decimal.Zero
	} else {
		return s.Basis.DivRound(s.Shares, 2)
	}
}

func (s Security) TotalReturn() decimal.Decimal {
	if s.Basis.IsZero() {
		return decimal.Zero
	}
	simpleReturn := s.Value.Sub(s.Basis).DivRound(s.Basis, 4)
	return decimalToPercentage(simpleReturn)
}

func (s *Security) setValue(price decimal.Decimal) decimal.Decimal {
	s.Value = s.Shares.Mul(price).Round(2)
	return s.Value
}

func (s *Security) addTrade(db *gorm.DB, trade *Trade) {
	updates := make(map[string]interface{})
	price := s.Price()

	// if trade.Date is newer than last time we pushed a Quote to database,
	// use given Price, else use Price computed above
	if trade.Price.IsPositive() && trade.Date.After(s.lastQuoteUpdate) {
		price = trade.Price
	}

	if trade.IsSell() {
		s.Basis = s.Basis.Sub(trade.Basis)
		s.Shares = s.Shares.Sub(trade.Shares)
		updates["basis"] = s.Basis
		updates["shares"] = s.Shares
	} else if trade.IsBuy() {
		s.Basis = s.Basis.Add(trade.Amount)
		s.Shares = s.Shares.Add(trade.Shares)
		updates["basis"] = s.Basis
		updates["shares"] = s.Shares
	} else if trade.IsSharesIn() {
		s.Shares = s.Shares.Add(trade.Shares)
		updates["shares"] = s.Shares
	} else if trade.IsSharesOut() {
		s.Shares = s.Shares.Sub(trade.Shares)
		updates["shares"] = s.Shares
	} else if trade.IsSplit() {
		s.Shares = s.Shares.Mul(trade.Shares)
		// value doesn't change for Split
		price = decimal.Zero
		updates["shares"] = s.Shares
	} else if !trade.Price.IsPositive() {
		return
	}

	// update Security Value as Shares changed;
	// if we sold all Shares, will update Value to Zero
	if !price.IsZero() {
		updates["value"] = s.setValue(price)
	}
	db.Omit(clause.Associations).Model(s).Updates(updates)
	log.Printf("[MODEL] SECURITY(%d) ADD TRADE (%d) TYPE(%d)",
		   s.ID, trade.ID, trade.TradeTypeID)
}

func (s *Security) updateTrade(db *gorm.DB, trade *Trade) {
	updates := make(map[string]interface{})
	price := s.Price()

	if trade.IsBuy() {
		s.Basis = s.Basis.Sub(trade.oldAmount)
		s.Basis = s.Basis.Add(trade.Amount)
		s.Shares = s.Shares.Sub(trade.oldShares)
		s.Shares = s.Shares.Add(trade.Shares)
		updates["basis"] = s.Basis
		updates["shares"] = s.Shares
	} else if trade.IsSharesIn() {
		s.Shares = s.Shares.Sub(trade.oldShares)
		s.Shares = s.Shares.Add(trade.Shares)
		updates["shares"] = s.Shares
	} else if trade.IsSharesOut() {
		s.Shares = s.Shares.Add(trade.oldShares)
		s.Shares = s.Shares.Sub(trade.Shares)
		updates["shares"] = s.Shares
	} else if trade.IsSplit() {
		s.Shares = s.Shares.Div(trade.oldShares)
		s.Shares = s.Shares.Mul(trade.Shares)
		// value doesn't change for Split
		price = decimal.Zero
		updates["shares"] = s.Shares
	} else {
		return
	}

	// update Security Value as Shares changed;
	// if we sold all Shares, will update Value to Zero
	if !price.IsZero() {
		updates["value"] = s.setValue(price)
	}
	db.Omit(clause.Associations).Model(s).Updates(updates)
	log.Printf("[MODEL] SECURITY(%d) UPDATE TRADE (%d) TYPE(%d)",
		   s.ID, trade.ID, trade.TradeTypeID)
}

// goroutine: this fetches latest Price and updates cached Quotes.
// It should not access the database.
func updateSecurities(securities []Security) {
	for i := 0; i < len(securities); i++ {
		if securities[i].Shares.IsPositive() {
			securities[i].fetchPrice(false)
		}
	}
}

// with account argument, Account access already verified by caller
func (s *Security) List(session *Session, account *Account, openPositions bool) []Security {
	entries := []Security{}
	db := session.DB

	if account == nil {
		// Verify we have access to Account
		s.Account.ID = s.AccountID
		account = s.Account.Get(session, false)
	}
	if account == nil || !account.Verified || !account.IsInvestment() {
		return entries
	}

	// Find Securities for Account
	if (openPositions) {
		db.Preload("Company").
		   Where("shares > 0 AND account_id = ?", account.ID).
		   Find(&entries)
	} else {
		db.Preload("Company").
		   Where(&Security{AccountID: account.ID}).
		   Find(&entries)
	}

	// initiate fetching of Security Quotes
	go updateSecurities(entries)

	log.Printf("[MODEL] LIST SECURITIES ACCOUNT(%d:%d)", account.ID, len(entries))
	return entries
}

// Security access already verified by caller
func (s *Security) ListTrades(db *gorm.DB, openOnly bool) []Trade {
	entries := []Trade{}
	if s.Account.Verified {
		dbQuery := db.Order("date asc").Preload("TradeType")
		if openOnly {
			dbQuery = dbQuery.Where("closed = 0").
					  Where(TradeTypeQueries[Buy])
		}
		// Find Trades for Security
		dbQuery.Where(&Trade{SecurityID: s.ID}).
			Find(&entries)
	}
	log.Printf("[MODEL] LIST TRADES SECURITY(%d:%d)", s.ID, len(entries))
	return entries
}

func (s *Security) validateSell(db *gorm.DB, trade *Trade) ([]Trade, error) {
	var buyShares decimal.Decimal

	activeBuys := s.ListTrades(db, true)
	if len(activeBuys) == 0 {
		return nil, errors.New("Invalid Sell Trade (No Shares)")
	}

	for i := 0; i < len(activeBuys); i++ {
		buy := &activeBuys[i]
		if !buy.Date.After(trade.Date) {
			buyShares = buyShares.Add(activeBuys[i].Shares)
		}
	}
	if buyShares.LessThan(trade.Shares) {
		return nil, errors.New("Invalid Sell Trade (Insufficient Shares)")
	}

	return activeBuys, nil
}

// TODO validate there are no Sells or Splits on/after trade.Date
func (s *Security) validateSplit(db *gorm.DB, trade *Trade) error {
	return nil
}

func (s *Security) validateTrade(db *gorm.DB, trade *Trade) ([]Trade, error) {
	if trade.IsSell() {
		return s.validateSell(db, trade)
	} else if trade.IsSplit() {
		err := s.validateSplit(db, trade)
		if err != nil {
			return nil, err
		}

		activeBuys := s.ListTrades(db, true)
		if len(activeBuys) == 0 {
			return nil, errors.New("Ignoring Split (No Shares)")
		}
		return activeBuys, nil
	}
	return nil, nil
}

func (s *Security) init() {
	s.SecurityTypeID = 1 // Default is Stock
	s.SecurityBasisTypeID = 1 // Default is FIFO
}

func (s *Security) Create(session *Session) error {
	db := session.DB
	// Verify we have access to Account
	s.Account.ID = s.AccountID
	s.init()
	account := s.Account.Get(session, false)
	if account != nil {
		spewModel(s)
		result := db.Omit(clause.Associations).Create(s)
		log.Printf("[MODEL] CREATE SECURITY(%d) ACCOUNT(%d)", s.ID, s.AccountID)
		if result.Error != nil {
			log.Fatal(result.Error)
		}
		return result.Error
	}
	return errors.New("Permission Denied")
}

// s.Account must be preloaded
func (s *Security) HaveAccessPermission(session *Session) bool {
	u := session.GetCurrentUser()
	s.Account.Verified = !(u == nil || s.Account.ID == 0 || u.ID != s.Account.UserID)
	if s.Account.Verified {
		s.Account.User = *u
		s.Account.Session = session
	}
	return s.Account.Verified
}

func (s *Security) updateValue() {
	// don't update when no Shares
	if s.Company.Symbol == "" || s.Shares.IsZero() ||
	   GetQuoteCache() == nil {
		return
	}

	quote := GetQuoteCache().Get(s.Company.Symbol)
	if quote.Price.IsPositive() {
		s.setValue(quote.Price)
	}
	if false {
		log.Printf("[MODEL] SECURITY(%d:%s) UPDATE VALUE(%f) (%f)",
			   s.ID, s.Company.Symbol,
			   s.Value.InexactFloat64(), quote.Price.InexactFloat64())
	}
}

// controllers(Get, Edit, Delete, Update) use Get
func (s *Security) Get(session *Session) *Security {
	db := session.DB
	db.Preload("SecurityType").Preload("Company").Preload("Account").First(&s)
	// Verify we have access to Security
	if !s.HaveAccessPermission(session) {
		return nil
	}

	s.Company.oldSymbol = s.Company.Symbol
	s.Company.oldName = s.Company.Name
	// updates s.Value (if have Shares) from latest Quote
	s.updateValue()

	log.Printf("[MODEL] GET SECURITY(%d:%s)", s.ID, s.Company.Symbol)
	return s
}

func (s *Security) Delete(session *Session) error {
	db := session.DB
	// Verify we have access to Security
	s = s.Get(session)
	if s != nil {
		spewModel(s)
		db.Delete(s)
		return nil
	}
	return errors.New("Permission Denied")
}

// Security access already verified with Get
func (s *Security) Update(session *Session) error {
	db := session.DB
	spewModel(s)

	// test if Symbol changed, and must update CompanyID
	updatedSymbol := s.Company.update(db)
	if updatedSymbol {
		s.CompanyID = s.Company.ID
	}
	result := db.Omit(clause.Associations).Save(s)
	if result.Error != nil {
		return result.Error
	}
	if !updatedSymbol {
		// test if just Company.Name changed
		s.Company.updateName(db)
	}

	log.Printf("[MODEL] UPDATE SECURITY(%d:%s)", s.ID, s.Company.Symbol)
	return nil
}
