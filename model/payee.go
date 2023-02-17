/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"errors"
	"log"
	"gorm.io/gorm/clause"
)

type Payee struct {
	Model
	UserID uint `gorm:"not null"`
	CategoryID uint `form:"payee.category_id"`
	Name string `form:"payee.Name"`
	Address string
	ImportName string `form:"payee.ImportName"`
	SkipOnImport bool `form:"payee.SkipOnImport"`
	User User
	Category Category
}

// for Bind() and setting from input/checkboxes */
func (p *Payee) ClearBooleans() {
	p.SkipOnImport = false
}

func (*Payee) List(session *Session) []Payee {
	u := session.GetCurrentUser()
	entries := []Payee{}
	if u == nil {
		return entries
	}
	db := session.DB

	// Find Payees for CurrentUser()
	db.Where(&Payee{UserID: u.ID}).Find(&entries)
	return entries
}

func payeeGetByName(session *Session, name string, importing bool) (error, *Payee) {
	u := session.GetCurrentUser()
	if u == nil {
		return errors.New("Permission Denied"), nil
	}
	db := session.DB
	created := false

	payee := new(Payee)
	payee.Name = name
	payee.UserID = u.ID
	// need Where because these are not primary keys
	db.Where(&payee).First(&payee)

	if payee.SkipOnImport {
		log.Printf("[MODEL] GET PAYEE(%d) BY NAME(%s) SKIP(1)",
			   payee.ID, name)
		return errors.New("Payee has SkipOnImport"), nil
	} else if payee.ID == 0 {
		db.Omit(clause.Associations).Create(payee)
		spewModel(payee)
		created = true
	}
	log.Printf("[MODEL] GET PAYEE(%d) BY NAME(%s) NEW(%t)",
		   payee.ID, name, created)

	return nil, payee
}

func (p *Payee) Create(session *Session) error {
	db := session.DB
	u := session.GetCurrentUser()
	if u != nil {
		// Payee.User is set to CurrentUser()
		p.UserID = u.ID
		spewModel(p)
		result := db.Omit(clause.Associations).Create(p)
		return result.Error
	}
	return errors.New("Permission Denied")
}

func (p *Payee) HaveAccessPermission(session *Session) bool {
	u := session.GetCurrentUser()
	return !(u == nil || u.ID != p.UserID)
}

// Edit, Delete, Update use Get
func (p *Payee) Get(session *Session) *Payee {
	db := session.DB
	db.Preload("User").First(&p)
	// Verify we have access to Payee
	if !p.HaveAccessPermission(session) {
		return nil
	}
	return p
}

func (p *Payee) Delete(session *Session) error {
	db := session.DB
	// Verify we have access to Payee
	p = p.Get(session)
	if p != nil {
		spewModel(p)
		db.Delete(p)
		return nil
	}
	return errors.New("Permission Denied")
}

// Payee access already verified with Get
func (p *Payee) Update() error {
	db := getDbManager()
	spewModel(p)
	result := db.Omit(clause.Associations).Save(p)
	return result.Error
}
