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
	TradeTypeID uint `form:"trade_type_id" gorm:"-:all"`
	TradeType TradeType
	AccountID uint `gorm:"not null"`
	oldAccountID uint `gorm:"-:all"`
	Account Account
	SecurityID uint `gorm:"not null"`
	Security Security
	Date time.Time
	TaxYear int `form:"tax_year"`
	Amount decimal.Decimal `form:"amount" gorm:"not null"`
	oldAmount decimal.Decimal `gorm:"-:all"`
	Closed bool
}

func (*Trade) List(db *gorm.DB, account *Account) []Trade {
	entries := []Trade{}
	if account.Verified {
		// Find Trades for Account()
		db.Where(&Trade{AccountID: account.ID}).Find(&entries)
		log.Printf("[MODEL] LIST TRADES ACCOUNT(%d:%d)", account.ID, len(entries))
	}
	return entries
}

func (t *Trade) Create(db *gorm.DB) error {
	// Verify we have access to Account
	t.Account.ID = t.AccountID
	account := t.Account.Get(db, false)
	if account != nil {
		spewModel(t)
		result := db.Create(t)
		log.Printf("[MODEL] CREATE %s TRADE(%d)", t.TradeType.Name, t.ID)
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
