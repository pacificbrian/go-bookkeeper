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
	Basis decimal.Decimal
	Closed bool
}

func (Trade) Currency(value decimal.Decimal) string {
	return currency(value)
}

// Account access already verified by caller
func (*Trade) List(db *gorm.DB, account *Account) []Trade {
	entries := []Trade{}
	if account.Verified {
		// Find Trades for Account()
		db.Preload("TradeType").
		   Order("date asc").
		   Where(&Trade{AccountID: account.ID}).Find(&entries)
		log.Printf("[MODEL] LIST TRADES ACCOUNT(%d:%d)", account.ID, len(entries))
	}
	return entries
}

// Look up Security by symbol, creates Security if none exists
func (t *Trade) securityGetBySymbol(db *gorm.DB) *Security {
	var security *Security

	if t.Symbol != "" {
		log.Printf("[MODEL] GET SECURITY for '%s'", t.Symbol)
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

	if t.SecurityID > 0 {
		// verify access to Security
		t.Security.ID = t.SecurityID
		security = t.Security.Get(db)
	} else {
		// verifies Account, creates Security if none exists
		security = t.securityGetBySymbol(db)
	}

	if security != nil {
		spewModel(t)
		result := db.Create(t)
		log.Printf("[MODEL] CREATE %s TRADE(%d)", t.TradeType.Name, t.ID)
		if result.Error != nil {
			log.Fatal(result.Error)
		}
		return result.Error
	}
	return errors.New("Permission Denied")
}

// t.Account must be preloaded
func (t *Trade) HaveAccessPermission() bool {
	u := GetCurrentUser()
	t.Account.Verified = !(u == nil || t.Account.ID == 0 || u.ID != t.Account.UserID)
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
	result := db.Save(t)
	return result.Error
}
