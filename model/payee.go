/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"github.com/davecgh/go-spew/spew"
	"gorm.io/gorm"
)

type Payee struct {
	Model
	UserID uint `gorm:"not null"`
	User User
	CategoryID uint `form:"payee.category_id"`
	Category Category
	Name string `form:"payee.Name"`
	Address string
	SkipOnImport bool `form:"payee.SkipOnImport"`
}

func (*Payee) List(db *gorm.DB) []Payee {
	u := GetCurrentUser()
	entries := []Payee{}
	if u == nil {
		return entries
	}

	// Find Payees for CurrentUser()
	db.Where(&Payee{UserID: u.ID}).Find(&entries)
	return entries
}

func payeeGetByName(db *gorm.DB, name string) *Payee {
	u := GetCurrentUser()
	if u == nil {
		return nil
	}

	payee := new(Payee)
	payee.Name = name
	payee.UserID = u.ID
	// need Where because these are not primary keys
	db.Where(&payee).First(&payee)

	if payee.ID == 0 {
		db.Create(payee)
		spew.Dump(payee)
	}

	return payee
}

func (p *Payee) HaveAccessPermission() bool {
	u := GetCurrentUser()
	return !(u == nil || u.ID != p.UserID)
}

// Edit, Delete, Update use Get
func (p *Payee) Get(db *gorm.DB) *Payee {
	db.Preload("User").First(&p)
	// Verify we have access to Payee
	if !p.HaveAccessPermission() {
		return nil
	}
	return p
}