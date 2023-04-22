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
	Address string `form:"payee.Address"`
	ImportName string `form:"payee.ImportName"`
	SkipOnImport bool `form:"payee.SkipOnImport"`
	Verified bool `gorm:"-:all"`
	User User
	Category Category
}

// for Bind() and setting from input/checkboxes */
func (p *Payee) ClearBooleans() {
	p.SkipOnImport = false
}

func (p *Payee) InUse() bool {
	return p.countCashFlows() > 0
}

func (p Payee) UseCount() uint {
	return p.countCashFlows()
}

func (p Payee) CategoryName() string {
	if p.CategoryID == 1 {
		return ""
	}
	return p.Category.Name
}

func (*Payee) List(session *Session) []Payee {
	u := session.GetUser()
	entries := []Payee{}
	if u == nil {
		return entries
	}
	db := session.DB

	// Find Payees for CurrentUser()
	db.Preload("Category").
	   Where(&Payee{UserID: u.ID}).Find(&entries)
	return entries
}

func (p *Payee) countCashFlows() uint {
	var count int64 = 0

	db := getDbManager()
	query := map[string]interface{}{"payee_id": p.ID, "transfer": false}
	db.Model(&CashFlow{}).
	   Where("(type != ? OR type IS NULL)", "RCashFlow"). // not Repeats
	   Where("NOT (split_from > 0 AND split = 0)"). // not HasSplits
	   Where("user_id = ?", p.UserID).Where(query).
	   Joins("Account").Count(&count)
	log.Printf("[MODEL] COUNT CASHFLOWS PAYEE(%d:%d)", p.ID, count)

	// TODO need to cache this result
	return uint(count)
}

func (p *Payee) ListCashFlows() []CashFlow {
	var entries []CashFlow

	if !p.Verified {
		return entries
	}

	db := getDbManager()
	query := map[string]interface{}{"payee_id": p.ID, "transfer": false}
	db.Order("date desc").Preload("Payee").Preload("Category").
	   Where("(type != ? OR type IS NULL)", "RCashFlow"). // not Repeats
	   Where("NOT (split_from > 0 AND split = 0)"). // not HasSplits
	   Where("user_id = ?", p.UserID).
	   Joins("Account").Find(&entries, query)
	log.Printf("[MODEL] LIST CASHFLOWS PAYEE(%d:%d)", p.ID, len(entries))

	for i := 0; i < len(entries); i++ {
		entries[i].Preload(db)
	}
	return entries
}

func payeeGetByName(session *Session, name string, importing bool) (error, *Payee) {
	u := session.GetUser()
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

	if importing && payee.SkipOnImport {
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
	u := session.GetUser()
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
	u := session.GetUser()
	p.Verified = !(u == nil || u.ID != p.UserID)
	return p.Verified
}

// Edit, Delete, Update use Get
func (p *Payee) Get(session *Session) *Payee {
	db := session.DB
	if p.ID > 0 {
		db.Preload("User").First(&p)
	}
	// Verify we have access to Payee
	if !p.HaveAccessPermission(session) {
		return nil
	}
	return p
}

func (p *Payee) Delete(session *Session) error {
	// Verify we have access to Payee
	p = p.Get(session)
	if p == nil {
		return errors.New("Permission Denied")
	}
	db := session.DB

	spewModel(p)
	db.Delete(p)
	return nil
}

// Payee access already verified with Get
func (p *Payee) Update() error {
	db := getDbManager()
	spewModel(p)
	result := db.Omit(clause.Associations).Save(p)
	return result.Error
}


// Find() for use with rails/ruby like REPL console (gomacro);
// controllers should not expose this as are no access controls
func (*Payee) Find(ID uint) *Payee {
	db := getDbManager()
	p := new(Payee)
	db.First(&p, ID)
	return p
}

func (p *Payee) Print() {
	forceSpewModel(p, 0)
}
