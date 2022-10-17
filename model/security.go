/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"errors"
	"log"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SecurityValue struct {
	Basis decimal.Decimal
	Shares decimal.Decimal
	Value decimal.Decimal
}

type Security struct {
	Model
	Company Company
	CompanyID uint `gorm:"not null"`
	SecurityBasisTypeID uint `form:"security_basis_type_id"`
	SecurityTypeID uint `form:"security_type_id"`
	AccountID uint `gorm:"not null"`
	Account Account
	SecurityValue
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

func (s *Security) addTrade(db *gorm.DB, trade *Trade) {
	updates := make(map[string]interface{})
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
		updates["shares"] = s.Shares
	} else if !trade.Price.IsPositive() {
		return
	}
	if trade.Price.IsPositive() || s.Shares.IsZero() {
		s.Value = s.Shares.Mul(trade.Price).Round(2)
		updates["value"] = s.Value
	}
	db.Omit(clause.Associations).Model(s).Updates(updates)
	log.Printf("[MODEL] SECURITY(%d) ADD TRADE (%d) TYPE(%d)",
		   s.ID, trade.ID, trade.TradeTypeID)
}

// with account argument, Account access already verified by caller
func (s *Security) List(db *gorm.DB, account *Account, openPositions bool) []Security {
	entries := []Security{}

	if account == nil {
		// Verify we have access to Account
		s.Account.ID = s.AccountID
		account = s.Account.Get(db, false)
	}
	if account != nil && account.Verified && account.IsInvestment() {
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
		log.Printf("[MODEL] LIST SECURITIES ACCOUNT(%d:%d)", account.ID, len(entries))
	}
	return entries
}

// Security access already verified by caller
func (s *Security) ListTrades(db *gorm.DB, openOnly bool) []Trade {
	entries := []Trade{}
	if s.Account.Verified {
		dbQuery := db.Order("date asc")
		if openOnly {
			dbQuery = dbQuery.Where("closed == 0").
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

func (s *Security) validateTrade(db *gorm.DB, trade *Trade) ([]Trade, error) {
	if trade.IsSell() {
		return s.validateSell(db, trade)
	} else if trade.IsSplit() {
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

func (s *Security) Create(db *gorm.DB) error {
	// Verify we have access to Account
	s.Account.ID = s.AccountID
	s.init()
	account := s.Account.Get(db, false)
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
func (s *Security) HaveAccessPermission() bool {
	u := GetCurrentUser()
	s.Account.Verified = !(u == nil || s.Account.ID == 0 || u.ID != s.Account.UserID)
	if s.Account.Verified {
		s.Account.User = *u
	}
	return s.Account.Verified
}

// Edit, Delete, Update use Get
func (s *Security) Get(db *gorm.DB) *Security {
	db.Preload("Account").First(&s)
	// Verify we have access to Security
	if !s.HaveAccessPermission() {
		return nil
	}
	return s
}

func (s *Security) Delete(db *gorm.DB) error {
	// Verify we have access to Security
	s = s.Get(db)
	if s != nil {
		spewModel(s)
		db.Delete(s)
		return nil
	}
	return errors.New("Permission Denied")
}

// Security access already verified with Get
func (s *Security) Update(db *gorm.DB) error {
	spewModel(s)
	result := db.Save(s)
	return result.Error
}
