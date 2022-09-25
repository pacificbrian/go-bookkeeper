/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"errors"
	"log"
	"gorm.io/gorm"
)

type Security struct {
	Model
	SecurityTypeID uint `form:"security_type_id" gorm:"-:all"`
	AccountID uint `gorm:"not null"`
	Account Account
	Symbol string `form:"Name"`
}

func (s *Security) List(db *gorm.DB) []Security {
	entries := []Security{}
	// Verify we have access to Account
	s.Account.ID = s.AccountID
	account := s.Account.Get(db, false)
	if account != nil {
		// Find Securities for Account()
		db.Where(&Security{AccountID: account.ID}).Find(&entries)
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
		result := db.Create(s)
		log.Printf("[MODEL] CREATE SECURITY(%d)", s.ID)
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