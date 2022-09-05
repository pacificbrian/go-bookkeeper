/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"errors"
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

func (p *Payee) Create(db *gorm.DB) error {
	u := GetCurrentUser()
	if u != nil {
		// Payee.User is set to CurrentUser()
		p.UserID = u.ID
		spew.Dump(p)
		result := db.Create(p)
		return result.Error
	}
	return errors.New("Permission Denied")
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

func (p *Payee) Delete(db *gorm.DB) error {
	// Verify we have access to Payee
	p = p.Get(db)
	if p != nil {
		spew.Dump(p)
		db.Delete(p)
		return nil
	}
	return errors.New("Permission Denied")
}

// Payee access already verified with Get
func (p *Payee) Update(db *gorm.DB) error {
	spew.Dump(p)
	result := db.Save(p)
	return result.Error
}
