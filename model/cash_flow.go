/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"time"
	"github.com/davecgh/go-spew/spew"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type CashFlow struct {
	gorm.Model
	CashFlowTypeID uint `form:"cash_flow_type_id" gorm:"-:all"`
	AccountID uint `gorm:"not null"`
	Account Account
	Date time.Time
	TaxYear int `form:"tax_year"`
	Amount decimal.Decimal `form:"amount" gorm:"not null"`
	Balance decimal.Decimal `gorm:"-:all"`
	SplitFrom uint `form:"split_from"`
	Split bool `form:"split"`
	Transfer bool
	Transnum string `form:"transnum"`
	Memo string `form:"memo"`
	PayeeID uint `gorm:"not null"`
	Payee Payee
	PayeeName string `form:"payee_name" gorm:"-:all"`
	CategoryID uint `form:"category_id"`
	Category Category
}

func (CashFlow) Currency(value decimal.Decimal) string {
	return  "$" + value.StringFixedBank(2)
}

func (CashFlow) ParentID() any {
	return nil
}

func (c *CashFlow) Preload(db *gorm.DB) {
	if c.Transfer {
		c.PayeeName = "Transfer"
	} else {
		c.Payee.ID = c.PayeeID
		db.First(&c.Payee)
		c.PayeeName = c.Payee.Name
	}
}

// Account access already verified by caller
func (*CashFlow) List(db *gorm.DB, account *Account) []CashFlow {
	entries := []CashFlow{}
	if account.Verified {
		db.Find(&entries, &CashFlow{AccountID: account.ID})
		// need to sort by Date

		// update Balances
		balance := account.Balance
		for i := 0; i < len(entries); i++ {
			c := &entries[i]
			c.Balance = balance
			balance = balance.Sub(c.Amount)
			c.Preload(db)
		}
	}
	return entries
}

// c.Account must be preloaded
func (c *CashFlow) HaveAccessPermission() bool {
	u := GetCurrentUser()
	return !(u == nil || u.ID != c.Account.UserID)
}

func (c *CashFlow) Create(db *gorm.DB) {
	// Verify we have access to Account
	c.Account.ID = c.AccountID
	account := c.Account.Get(db, false)
	if account != nil {
		c.TaxYear = c.Date.Year()
		if c.PayeeName != "" {
			p := payeeGetByName(db, c.PayeeName)
			c.PayeeID = p.ID
		}
		spew.Dump(c)
		db.Create(c)
	}
}

// Edit, Delete, Update use Get
// c.Account needs to be preloaded
func (c *CashFlow) Get(db *gorm.DB, preload bool) *CashFlow {
	db.Preload("Account").First(&c)
	// Verify we have access to CashFlow
	if !c.HaveAccessPermission() {
		return nil
	}
	if preload {
		c.Preload(db)
	}
	return c
}

func (c *CashFlow) Delete(db *gorm.DB) {
	// Verify we have access to CashFlow
	c = c.Get(db, false)
	if c != nil {
		spew.Dump(c)
		db.Delete(c)
	}
}

// CashFlow access already verified with Get
func (c *CashFlow) Update(db *gorm.DB) *CashFlow {
	spew.Dump(c)
	db.Save(c)
	return c
}
