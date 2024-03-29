/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"errors"
	"fmt"
	"log"
	"gorm.io/gorm/clause"
)

type Payee struct {
	Model
	UserID uint `gorm:"not null"`
	CategoryID uint `form:"payee.category_id"`
	Name string `form:"payee.Name"`
	Address string `form:"payee.Address"`
	SkipOnImport bool `form:"payee.SkipOnImport"`
	Verified bool `gorm:"-:all"`
	User User
	Category Category
}

func (p *Payee) sanitizeInputs() {
	sanitizeString(&p.Name)
	sanitizeString(&p.Address)
}

// for Bind() and setting from input/checkboxes */
func (p *Payee) ClearBooleans() {
	p.SkipOnImport = false
}

func (p *Payee) InUse() bool {
	return p.countCashFlows(nil) > 0
}

func (p Payee) UseByAccount(a *Account) uint {
	return p.countCashFlows(a)
}

func (p Payee) UseCount() uint {
	return p.countCashFlows(nil)
}

func (p Payee) CategoryName() string {
	if p.CategoryID == 1 {
		return ""
	}
	return p.Category.Name
}

func (*Payee) List(session *Session, account *Account) []Payee {
	u := session.GetUser()
	entries := []Payee{}
	if u == nil {
		return entries
	}
	db := session.DB

	if account != nil && account.ID > 0 {
		var payee_ids []uint

		// Find Payees for CurrentUser() used with Account
		db.Model(&CashFlow{}).
		   Where("(type != ? OR type IS NULL)", "RCashFlow"). // not Repeats
		   Where("NOT (split_from > 0 AND split = 0)"). // not HasSplits
		   Where("transfer = ?", false).
		   Where("account_id = ?", account.ID).
		   Distinct().Pluck("payee_id", &payee_ids)

		db.Order("name asc").Preload("Category").
		   Find(&entries, payee_ids)
	} else {
		// Find Payees for CurrentUser()
		db.Order("name asc").Preload("Category").
		   Where(&Payee{UserID: u.ID}).Find(&entries)
	}
	return entries
}

// used with InUse counts, so should include ScheduledCashFlows
func (p *Payee) countCashFlows(account *Account) uint {
	var count int64 = 0

	db := getDbManager()
	query := map[string]interface{}{"payee_id": p.ID, "transfer": false}
	if account != nil && account.ID > 0 {
		query["account_id"] = account.ID
	}
	db.Model(&CashFlow{}).
	   Where("NOT (split_from > 0 AND split = 0)"). // not HasSplits
	   Where("user_id = ?", p.UserID).Where(query).
	   Joins("Account").Count(&count)
	log.Printf("[MODEL] COUNT CASHFLOWS PAYEE(%d:%d)", p.ID, count)

	// TODO need to cache this result
	return uint(count)
}

func (p *Payee) listCashFlows(account *Account, getSplits bool) []CashFlow {
	var entries []CashFlow

	if !p.Verified {
		return entries
	}

	db := getDbManager()
	query := map[string]interface{}{"payee_id": p.ID, "transfer": false}
	if account != nil && account.ID > 0 {
		query["account_id"] = account.ID
	}

	dbQuery := db.Order("date desc").Preload("Payee").Preload("Category")
	if getSplits {
		dbQuery = dbQuery.
		   Where("(type != ? OR type IS NULL)", "RCashFlow"). // not Repeats
		   Where("NOT (split_from > 0 AND split = 0)") // not HasSplits
	} else {
		dbQuery = dbQuery.Where("type IS NULL") // no Repeats or Splits
	}
	dbQuery.Where("user_id = ?", p.UserID).
		Joins("Account").Find(&entries, query)
	log.Printf("[MODEL] LIST CASHFLOWS PAYEE(%d:%d)", p.ID, len(entries))

	for i := 0; i < len(entries); i++ {
		entries[i].Preload(db)
	}
	return entries
}

func (p *Payee) ListCashFlows(account *Account) []CashFlow {
	return p.listCashFlows(account, true)
}

func (p *Payee) getByName(session *Session, importing bool) (error, *Payee) {
	u := session.GetUser()
	if u == nil {
		return errors.New("Permission Denied"), nil
	}
	db := session.DB
	created := false

	payee := new(Payee)
	payee.Name = p.Name
	payee.UserID = u.ID
	// need Where because these are not primary keys
	db.Where(&payee).First(&payee)

	if importing && payee.SkipOnImport {
		log.Printf("[MODEL] GET PAYEE(%d) BY NAME(%s) SKIP(1)",
			   payee.ID, payee.Name)
		return errors.New("Payee has SkipOnImport"), nil
	} else if payee.ID == 0 {
		payee.Address = p.Address
		db.Omit(clause.Associations).Create(payee)
		spewModel(payee)
		created = true
	}
	log.Printf("[MODEL] GET PAYEE(%d) BY NAME(%s) NEW(%t)",
		   payee.ID, payee.Name, created)

	return nil, payee
}

func (p *Payee) GetDuplicates() []Payee {
	entries := []Payee{}

	if !p.Verified {
		return entries
	}

	db := getDbManager()
	wildcardName := fmt.Sprintf("%s%%", p.Name)
	db.Order("name asc").Where("user_id = ?", p.UserID).
	   Where("name LIKE ?", wildcardName).
	   Where("id != ?", p.ID).
	   Find(&entries)
	return entries
}

func (p *Payee) SetCategory(account *Account, categoryID uint, toAll bool) error {
	cat := new(Category)
	cat.ID = categoryID
	if !cat.IsValid() {
		return errors.New("Invalid Category")
	}

	cashflows := p.ListCashFlows(account)
	for i := 0; i < len(cashflows); i++ {
		c := &cashflows[i]
		if toAll || (c.CategoryID == 1) {
			c.postQueryInit(true)
			c.CategoryID = categoryID
			// ensure c.Update uses c.PayeeID (no lookup)
			c.PayeeName = ""
			c.PayeeID = p.ID
			c.Account.setSession(p.User.Session)
			c.Update()
		}
	}
	return nil
}

func (p *Payee) Merge(toMerge *Payee, account *Account) error {
	if toMerge.ID != 0 {
		log.Printf("[MODEL] MERGE PAYEE(%d) WITH(%d)",
			   p.ID, toMerge.ID)
		// get CashFlows (no Scheduled or Splits)
		cashflows := toMerge.listCashFlows(account, false)
		for i := 0; i < len(cashflows); i++ {
			c := &cashflows[i]
			c.postQueryInit(true)
			// ensure c.Update uses c.PayeeID (no lookup)
			c.PayeeName = ""
			c.PayeeID = p.ID
			c.Account.setSession(p.User.Session)
			c.Update()
		}
		// will verify Payee is unused
		toMerge.Delete(p.User.Session)
	} else {
		mergeList := p.GetDuplicates()
		log.Printf("[MODEL] MERGE PAYEE(%d) MANY(%d)",
			   p.ID, len(mergeList))
		for _, m := range mergeList {
			m.Verified = p.Verified
			p.Merge(&m, account)
		}
	}
	return nil
}

func (p *Payee) Create(session *Session) error {
	db := session.DB
	u := session.GetUser()
	if u == nil {
		return errors.New("Permission Denied")
	}

	p.sanitizeInputs()
	// Payee.User is set to CurrentUser()
	p.UserID = u.ID
	spewModel(p)
	result := db.Omit(clause.Associations).Create(p)
	return result.Error
}

func (p *Payee) HaveAccessPermission(session *Session) bool {
	u := session.GetUser()
	p.Verified = !(u == nil || u.ID != p.UserID)
	if p.Verified {
		p.User = *u
	}
	return p.Verified
}

// Edit, Delete, Update use Get
func (p *Payee) Get(session *Session) *Payee {
	db := session.DB
	if p.ID > 0 {
		db.First(&p)
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
	count := p.countCashFlows(nil)
	log.Printf("[MODEL] DELETE PAYEE(%d) IF COUNT(%d == 0)", p.ID, count)
	if count == 0 {
		db.Delete(p)
	}
	return nil
}

// Payee access already verified with Get
func (p *Payee) Update() error {
	db := getDbManager()
	p.sanitizeInputs()
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
