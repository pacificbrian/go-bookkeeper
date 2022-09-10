/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"errors"
	"time"
	"github.com/davecgh/go-spew/spew"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	UndefinedType uint = iota
	Debit
	Credit
	DebitTransfer
	CreditTransfer
)

type CashFlow struct {
	gorm.Model
	CashFlowTypeID uint `form:"cash_flow_type_id" gorm:"-:all"`
	AccountID uint `gorm:"not null"`
	oldAccountID uint `gorm:"-:all"`
	Account Account
	Date time.Time
	TaxYear int `form:"tax_year"`
	Amount decimal.Decimal `form:"amount" gorm:"not null"`
	oldAmount decimal.Decimal `gorm:"-:all"`
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
	CategoryName string `gorm:"-:all"`
}

func (CashFlow) Currency(value decimal.Decimal) string {
	return  "$" + value.StringFixedBank(2)
}

func (CashFlow) ParentID() any {
	return nil
}

func (c *CashFlow) Preload(db *gorm.DB) {
	if c.Transfer {
		a := new(Account)
		db.First(&a, c.PayeeID)
		c.PayeeName = a.Name
		c.CategoryName = "Transfer"
	} else {
		c.Payee.ID = c.PayeeID
		db.First(&c.Payee)
		c.PayeeName = c.Payee.Name

		if c.CategoryID > 0 {
			c.Category.ID = c.CategoryID
			db.First(&c.Category)
			c.CategoryName = c.Category.Name
		}
	}
}

// Account access already verified by caller
func (*CashFlow) List(db *gorm.DB, account *Account) []CashFlow {
	entries := []CashFlow{}
	if account.Verified {
		// sort by Date
		db.Order("date desc").Find(&entries, &CashFlow{AccountID: account.ID})

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

func (c *CashFlow) applyCashFlowType() {
	switch c.CashFlowTypeID {
	case Debit:
		c.Amount = c.Amount.Neg()
	case DebitTransfer:
		c.Amount = c.Amount.Neg()
		c.Transfer = true
	case CreditTransfer:
		c.Transfer = true
	}
}

func (c *CashFlow) cloneTransfer(src *CashFlow) {
	c.Transfer = true
	c.Date = src.Date
	c.TaxYear = src.TaxYear
	c.Memo = src.Memo
	c.Transnum = src.Transnum
	c.oldAccountID = c.AccountID // used if Update
	c.AccountID = src.PayeeID
	c.PayeeID = src.AccountID
	c.oldAmount = c.Amount // used if Update
	c.Amount = src.Amount.Neg()
}

func (c *CashFlow) prepareInsertCashFlow(db *gorm.DB) (error, *CashFlow) {
	var pair *CashFlow = nil // Transfer Pair

	c.applyCashFlowType()
	if c.Transfer {
		if c.PayeeName != "" {
			a := accountGetByName(db, c.PayeeName)
			if a == nil {
				return errors.New("Account.Name Invalid"), nil
			}
			// store pair account.ID in PayeeID (aka TransferAccountID)
			c.PayeeID = a.ID

			// create pair CashFlow
			pair = new(CashFlow)
			pair.cloneTransfer(c)
		}
	} else {
		if c.PayeeName != "" {
			// creates Payee if none exists
			p := payeeGetByName(db, c.PayeeName)
			c.PayeeID = p.ID
		}
	}

	return nil, pair
}

func (c *CashFlow) Create(db *gorm.DB) error {
	// Verify we have access to Account
	c.Account.ID = c.AccountID
	account := c.Account.Get(db, false)
	if account != nil {
		// defaults for DB fields not set during Create (Edit only)
		c.TaxYear = c.Date.Year()

		err, pair := c.prepareInsertCashFlow(db)
		if err == nil {
			spew.Dump(c)
			result := db.Create(c)
			err = result.Error
		}
		if err == nil {
			account.UpdateBalance(db, c)

			// create pair CashFlow if have one (Transfers)
			if pair != nil {
				// categoryID stores paired CashFlow.ID
				pair.CategoryID = c.ID
				db.Create(pair)
				c.CategoryID = pair.ID
				db.Model(c).Update("CategoryID", pair.ID)

				pair.Account.ID = pair.AccountID
				pair.Account.UpdateBalance(db, pair)
			}
		}
		return err
	}
	return errors.New("Permission Denied")
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
	c.oldAmount = c.Amount
	return c
}

func (c *CashFlow) Delete(db *gorm.DB) error {
	// Verify we have access to CashFlow
	c = c.Get(db, false)
	if c != nil {
		spew.Dump(c)
		db.Delete(c)

		account := new(Account)
		account.ID = c.AccountID
		c.Amount = decimal.Zero
		account.UpdateBalance(db, c)

		return nil
	}
	return errors.New("Permission Denied")
}

// CashFlow access already verified with Get
func (c *CashFlow) Update(db *gorm.DB) error {
	spew.Dump(c)
	result := db.Save(c)
	if result.Error == nil {
		account := new(Account)
		account.ID = c.AccountID
		account.UpdateBalance(db, c)
	}
	return result.Error
}

// Debug routines -

func CashFlowFind(db *gorm.DB, id uint) *CashFlow {
	c := new(CashFlow)
	db.First(&c, id)
	return c
}
