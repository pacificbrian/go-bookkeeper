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
		return s.Value.Div(s.Shares).RoundBank(2)
	}
}

func (s Security) BasisPrice() decimal.Decimal {
	if s.Shares.Equal(decimal.Zero) {
		return decimal.Zero
	} else {
		return s.Basis.Div(s.Shares).RoundBank(2)
	}
}

// with account argument, Account access already verified by caller
func (s *Security) List(db *gorm.DB, account *Account) []Security {
	entries := []Security{}

	if account == nil {
		// Verify we have access to Account
		s.Account.ID = s.AccountID
		account = s.Account.Get(db, false)
	}
	if account != nil && account.Verified && account.IsInvestment() {
		// Find Securities for Account()
		db.Preload("Company").
		   Where(&Security{AccountID: account.ID}).Find(&entries)
		log.Printf("[MODEL] LIST SECURITIES ACCOUNT(%d:%d)", account.ID, len(entries))
	}
	return entries
}

func (s *Security) Create(db *gorm.DB) error {
	// Verify we have access to Account
	s.Account.ID = s.AccountID
	account := s.Account.Get(db, false)
	if account != nil {
		spewModel(s)
		result := db.Omit(clause.Associations).Create(s)
		log.Printf("[MODEL] CREATE SECURITY(%d)", s.ID)
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
