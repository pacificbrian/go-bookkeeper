/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"errors"
	"log"
	"net/http"
	"time"
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
	oldDate time.Time
	TaxYear int `form:"tax_year"`
	Amount decimal.Decimal `form:"amount" gorm:"not null"`
	oldAmount decimal.Decimal `gorm:"-:all"`
	Balance decimal.Decimal `gorm:"-:all"`
	SplitFrom uint
	Split bool
	Transfer bool
	Transnum string `form:"transnum"`
	Memo string `form:"memo"`
	PayeeID uint `gorm:"not null"` // also serves as Pair.AccountID (Transfers)
	Payee Payee
	PayeeName string `form:"payee_name" gorm:"-:all"`
	CategoryID uint `form:"category_id"` // also serves as Pair.ID (Transfers)
	oldPairID uint `gorm:"-:all"`
	Category Category
	CategoryName string `gorm:"-:all"`
	RepeatIntervalID uint
	RepeatInterval RepeatInterval
	Type string `gorm:"default:NULL"`
}

func (CashFlow) Currency(value decimal.Decimal) string {
	return currency(value)
}

func (c CashFlow) ParentID() uint {
	if !c.Split {
		return 0
	}
	return c.SplitFrom
}

func (c *CashFlow) CanSplit() bool {
	return !(c.Transfer || c.Split)
}

func (c *CashFlow) setSplit(SplitFrom uint) {
	if SplitFrom > 0 {
		if c.Type != "Repeat" {
			c.Type = "Split"
		}
		c.Split = true
	}
	c.SplitFrom = SplitFrom
}

func (c *CashFlow) IsScheduled() bool {
	return c.Type == "Repeat"
}

func (c *CashFlow) IsScheduledParent() bool {
	return c.IsScheduled() && !c.Split
}

func (c *CashFlow) mustUpdateBalance() bool {
	// aka Base Type (!Split and !Repeat)
	return c.Type ==  ""
}

// Used with CreateSplitCashFlow. Controller calls to get common CashFlow
// fields first, and before Bind (which can/will override other fields).
func NewSplitCashFlow(db *gorm.DB, SplitFrom uint) (*CashFlow, int) {
	c := new(CashFlow)
	c.ID = SplitFrom
	c = c.Get(db, false)
	if c == nil {
		return nil, http.StatusUnauthorized
	}
	if !c.CanSplit() {
		return nil, http.StatusBadRequest
	}

	c.setSplit(SplitFrom)
	c.oldAmount = decimal.Zero
	c.ID = 0
	return c, 0
}

func (c *CashFlow) SplitCount() uint {
	var count uint = 0
	if !c.Transfer && !c.Split && c.SplitFrom > 0 {
		count = c.SplitFrom
	}
	return count
}

func (c *CashFlow) HasSplits() bool {
	return c.SplitCount() > 0
}

func (c *CashFlow) PreloadScheduled(db *gorm.DB) {
	if c.IsScheduled() {
		c.RepeatInterval.ID = c.RepeatIntervalID
		c.RepeatInterval.Preload(db)
	}
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

		if c.HasSplits() {
			c.CategoryName = "Split"
		} else if c.CategoryID > 0 {
			// need userCache lookup
			c.Category.ID = c.CategoryID
			db.First(&c.Category)
			c.CategoryName = c.Category.Name
		}
	}

	c.PreloadScheduled(db)
}

// Account access already verified by caller
func (*CashFlow) List(db *gorm.DB, account *Account) []CashFlow {
	entries := []CashFlow{}
	if account.Verified {
		// sort by Date
		// db.Order("date desc").Find(&entries, &CashFlow{AccountID: account.IDl})
		// use map to support NULL string
		query := map[string]interface{}{"account_id": account.ID, "type": nil}
		db.Order("date desc").Find(&entries, query)

		// update Balances
		balance := account.Balance
		for i := 0; i < len(entries); i++ {
			c := &entries[i]
			c.Balance = balance
			balance = balance.Sub(c.Amount)
			c.Preload(db)
		}

		log.Printf("[MODEL] LIST CASHFLOWS ACCOUNT(%d:%d)", account.ID, len(entries))
	}
	return entries
}

// Account access already verified by caller
func (c *CashFlow) ListSplit(db *gorm.DB) ([]CashFlow, string) {
	var total decimal.Decimal
	entries := []CashFlow{}
	if c.HasSplits() && c.Account.Verified {
		db.Find(&entries, &CashFlow{AccountID: c.AccountID, SplitFrom: c.ID, Split: true})
		for i := 0; i < len(entries); i++ {
			split := &entries[i]
			split.Preload(db)
			total = total.Add(split.Amount)
		}
	}
	return entries, currency(total)
}

// c.Account must be preloaded
func (c *CashFlow) HaveAccessPermission() bool {
	u := GetCurrentUser()
	c.Account.Verified = !(u == nil || c.Account.ID == 0 || u.ID != c.Account.UserID)
	return c.Account.Verified
}

func (c *CashFlow) determineCashFlowType() {
	if c.Amount.IsPositive() {
		c.CashFlowTypeID = Credit
	} else {
		c.CashFlowTypeID = Debit
	}
	if c.Transfer {
		c.CashFlowTypeID += 2
	}
}

func (c *CashFlow) applyCashFlowType() {
	// Transfer not set during Bind, is updated here
	switch c.CashFlowTypeID {
	case Debit:
		c.Amount = c.Amount.Neg()
		c.Transfer = false
	case Credit:
		c.Transfer = false
	case DebitTransfer:
		c.Amount = c.Amount.Neg()
		c.Transfer = true
	case CreditTransfer:
		c.Transfer = true
	}
}

func (c *CashFlow) cloneScheduled(src *CashFlow) {
	c.Transfer = src.Transfer
	if src.Split {
		c.setSplit(src.SplitFrom)
	}
	c.Date = src.Date
	c.TaxYear = c.Date.Year()
	c.Memo = src.Memo
	c.Transnum = src.Transnum
	c.AccountID = src.AccountID
	c.Account.ID = c.AccountID
	c.Account.Verified = src.Account.Verified
	c.PayeeID = src.PayeeID
	c.CategoryID = src.CategoryID
	c.Amount = src.Amount
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
	c.CategoryID = src.ID
	c.oldAmount = c.Amount // used if Update
	c.Amount = src.Amount.Neg()
}

// Using src CashFlow, construct the Pair (other side of a Transfer).
// This is used during Update or Delete.
// We can even reconstruct Splits (careful: old transactions in DB
// don't store this!)
func (c *CashFlow) pairFrom(src *CashFlow) {
	c.Transfer = true
	c.ID = src.oldPairID
	// keep split details accurate, and decrement SplitCount in Parent (Delete)
	c.setSplit(src.SplitFrom)
	c.AccountID = src.PayeeID
	c.Amount = src.oldAmount.Neg()
	c.oldAmount = c.Amount // used if Delete
}

// prepare CashFlow to write to DB (used by both Create and Update)
//   - update Amount and Transfer based on CashFlowType
//   - create Payee if needed
//   - lookup Account (error if not found/accessible)
//   - return Pair cashflow (for other Account) if this is a Transfer
//   - UPDATEs are allowed to change to/from Transfer type and change Peer Account
func (c *CashFlow) prepareInsertCashFlow(db *gorm.DB) (error, *CashFlow) {
	var pair *CashFlow = nil // Transfer Pair

	if c.Transfer {
		var a *Account

		if c.PayeeName != "" {
			a = accountGetByName(db, c.PayeeName)
			if a == nil {
				return errors.New("Account.Name Invalid"), nil
			}
		}

		if a != nil && !c.IsScheduled() {
			// create pair CashFlow
			pair = new(CashFlow)
			if c.oldPairID > 0 {
				// #UPDATE: use existing pair CashFlow
				c.CategoryID = c.oldPairID
				pair.pairFrom(c)
			}

			// store pair account.ID in PayeeID (aka TransferAccountID)
			c.PayeeID = a.ID

			// fill in pair CashFlow with remaining details
			pair.cloneTransfer(c)
			// #UPDATE: if pair.AccountID changed, this is handled in caller
		}
	} else {
		// #UPDATE: if Transfer type True->False, delete pair CashFlow
		if c.oldPairID > 0 {
			oldPair := new(CashFlow)
			oldPair.pairFrom(c)
			oldPair.delete(db)
		}

		if !c.Split && c.PayeeName != "" {
			// creates Payee if none exists
			p := payeeGetByName(db, c.PayeeName)
			c.PayeeID = p.ID
		}
	}

	return nil, pair
}

// c.Account access must be verified
func (c *CashFlow) insertCashFlow(db *gorm.DB) error {
	if !c.Account.Verified {
		log.Printf("[MODEL] INSERT CASHFLOW PERMISSION DENIED")
		return errors.New("Permission Denied")
	}
	err, pair := c.prepareInsertCashFlow(db)
	if err == nil {
		result := db.Create(c)
		err = result.Error
	}
	if err != nil {
		log.Fatalf("[MODEL] INSERT CASHFLOW ERROR: %s", err)
		return err
	}
	// insert successful, no errors after this point

	if c.Split {
		log.Printf("[MODEL] CREATE SPLIT CASHFLOW(%d) PARENT(%d)", c.ID, c.SplitFrom)
		spewModel(c)

		// increment split count in parent
		parent := new(CashFlow)
		parent.ID = c.SplitFrom
		db.Model(parent).Update("split_from", gorm.Expr("split_from + ?", 1))
	} else {
		log.Printf("[MODEL] CREATE %s CASHFLOW(%d)", c.Type, c.ID)
		spewModel(c)
		c.Account.UpdateBalance(db, c)
	}

	// Create pair CashFlow if have one (Transfers)
	// Note, impossible to be a Split
	// Create should not be able to fail as cloned from primary CashFlow
	if pair != nil {
		// mark when paired with a Split (Update restrictions)
		if c.Split {
			pair.SplitFrom = c.SplitFrom
		}
		// categoryID stores paired CashFlow.ID
		pair.CategoryID = c.ID
		db.Create(pair)
		c.CategoryID = pair.ID
		db.Model(c).Update("CategoryID", pair.ID)
		log.Printf("[MODEL] CREATE PAIR CASHFLOW(%d)", pair.ID)

		pair.Account.ID = pair.AccountID
		pair.Account.UpdateBalance(db, pair)
	}

	return err
}

func (c *CashFlow) splitUpdateMap() map[string]interface{} {
	// map of fields that must be equivalent in Split/Parent
	// if Transfer, payee_id is pruned out later
	return map[string]interface{}{"date": c.Date, "tax_year": c.TaxYear,
				      "payee_id": c.PayeeID}
}

// update only selected fields in Splits from the given map
func (repeat *CashFlow) updateSplits(db *gorm.DB, updates map[string]interface{}) {
	if repeat.HasSplits() {
		// for Transfers, copy map and remove payee_id
		// wish there was cleaner way
		transferUpdates := make(map[string]interface{})
		for k,v := range updates {
			transferUpdates[k] = v
		}
		delete(transferUpdates, "payee_id")

		splits, _ := repeat.ListSplit(db)
		for i := 0; i < len(splits); i++ {
			split := splits[i]
			if split.Transfer {
				db.Model(split).Updates(transferUpdates)
			} else {
				db.Model(split).Updates(updates)
			}
		}
	}
}

// returns true if advanced date is still less than time.Now
func (repeat *CashFlow) advance(db *gorm.DB) bool {
	repeat.RepeatInterval.ID = repeat.RepeatIntervalID
	days := repeat.RepeatInterval.Advance(db)
	if days == 0 {
		return false
	}

	day_of_month := repeat.Date.Day()
	if repeat.RepeatInterval.StartDay > 0 {
		day_of_month = repeat.RepeatInterval.StartDay
	}

	if days < 15 {
		// weekly / bi-weekly
		repeat.Date = repeat.Date.AddDate(0, 0, days)
	} else if days >= 30 {
		// monthly, quarterly, annually, etc
		months := days / 30
		adjustedDate := repeat.Date.AddDate(0, months, day_of_month - repeat.Date.Day())
		if  adjustedDate.Day() < repeat.Date.Day() {
			// we overran into next month (less than 30/31 days)
			adjustedDate = adjustedDate.AddDate(0, 0, -adjustedDate.Day())
		}
		repeat.Date = adjustedDate
	} else {
		// semi-monthly, one of two halves should use day_of_month exactly
		if repeat.Date.Day() <= 15 {
			// advance to 2nd half of month
			adjustedDate := repeat.Date.AddDate(0, 0, 15)
			if  adjustedDate.Day() < repeat.Date.Day() {
				// we overran into next month (less than 30/31 days)
				adjustedDate = adjustedDate.AddDate(0, 0, -adjustedDate.Day())
			}
			repeat.Date = adjustedDate
		} else {
			if day_of_month > 15 {
				day_of_month -= 15
			}
			// advance to next month
			repeat.Date = repeat.Date.AddDate(0, 1, day_of_month - repeat.Date.Day())
		}
	}
	repeat.TaxYear = repeat.Date.Year()

	updates := map[string]interface{}{"date": repeat.Date, "tax_year": repeat.TaxYear}
	db.Model(repeat).Updates(updates)
	log.Printf("[MODEL] ADVANCE SCHEDULED CASHFLOW(%d)", repeat.ID,
		   repeat.Date.Format("2006-01-02"))
	repeat.updateSplits(db, updates)

	return time.Now().After(repeat.Date)
}

func (repeat *CashFlow) tryInsertRepeatCashFlow(db *gorm.DB) error {
	var err error
	c := new(CashFlow)
	for {
		c.cloneScheduled(repeat)
		err = c.insertCashFlow(db)
		if err != nil || c.Split {
			break
		}

		// handle Splits
		splits, _ := repeat.ListSplit(db)
		for i := 0; i < len(splits); i++ {
			split := splits[i]
			split.SplitFrom = c.ID
			split.tryInsertRepeatCashFlow(db)
		}

		canRepeat := repeat.advance(db)
		if !canRepeat {
			break
		}
		c.ID = 0
	}
	return err
}

func (c *CashFlow) Create(db *gorm.DB) error {
	// Verify we have access to Account
	if !c.Account.Verified {
		c.Account.ID = c.AccountID
		account := c.Account.Get(db, false)
		if account == nil {
			return errors.New("Permission Denied")
		}
	}

	c.applyCashFlowType()
	// defaults for DB fields not set during Create (are Edit only)
	c.TaxYear = c.Date.Year()

	err := c.insertCashFlow(db)
	if err == nil && c.IsScheduledParent() {
		_err := c.RepeatInterval.Create(db, c)
		if _err != nil {
			log.Fatalf("INSERT REPEAT_INTERVAL ERROR: %s", _err)
		}
		c.RepeatIntervalID = c.RepeatInterval.ID
		db.Model(c).Update("RepeatIntervalID", c.RepeatIntervalID)
	}

	return err
}

// Edit, Delete, Update use Get
// c.Account needs to be preloaded
func (c *CashFlow) Get(db *gorm.DB, edit bool) *CashFlow {
	db.Preload("Account").First(&c)
	// Verify we have access to CashFlow
	if !c.HaveAccessPermission() {
		return nil
	}

	c.determineCashFlowType()
	c.oldAmount = c.Amount
	c.oldDate = c.Date
	if c.Transfer {
		// backup CategoryID as cleared by Bind
		c.oldPairID = c.CategoryID // Peer Cashflow (Transfers)
		c.CategoryID = 0
	}

	if edit {
		c.Preload(db)

		// #Edit wants Amount to be always positive; safe to
		// modify here because Delete doen't use, and Update overwrites
		c.Amount = c.Amount.Abs()
	} else {
		c.PreloadScheduled(db)
	}

	return c
}

func (c *CashFlow) deleteTransfer(db *gorm.DB) {
	if c.Transfer {
		pair := new(CashFlow)
		pair.pairFrom(c)
		pair.delete(db)
	}
}

func (c *CashFlow) delete(db *gorm.DB) {
	if c.Split {
		log.Printf("[MODEL] DELETE CASHFLOW(%d) PARENT(%d)", c.ID, c.SplitFrom)

		// decrement split count in parent
		parent := new(CashFlow)
		parent.ID = c.SplitFrom
		db.Model(parent).Update("split_from", gorm.Expr("split_from - ?", 1))

		db.Delete(c)
	} else {
		log.Printf("[MODEL] DELETE CASHFLOW(%d)", c.ID)
		if c.HasSplits() {
			splits, _ := c.ListSplit(db)
			for i := 0; i < len(splits); i++ {
				split := splits[i]
				split.delete(db)
				split.deleteTransfer(db)
			}
		}

		db.Delete(c)

		c.Account.ID = c.AccountID
		c.Amount = decimal.Zero
		// UpdateBalance will subtract c.oldAmount
		c.Account.UpdateBalance(db, c)
	}
}

func (c *CashFlow) Delete(db *gorm.DB) error {
	// Verify we have access to CashFlow
	c = c.Get(db, false)
	if c != nil {
		c.delete(db)
		c.deleteTransfer(db)
	}
	return errors.New("Permission Denied")
}

// CashFlow access already verified with Get
func (c *CashFlow) Update(db *gorm.DB) error {
	c.applyCashFlowType()
	if c.Split {
		// don't let Splits mess with date
		c.Date = c.oldDate
	}

	err, pair := c.prepareInsertCashFlow(db)
	if err == nil {
		result := db.Save(c)
		err = result.Error
	}
	if err == nil {
		c.Account.ID = c.AccountID
		if c.Split {
			log.Printf("[MODEL] UPDATE CASHFLOW(%d) PARENT(%d)", c.ID, c.SplitFrom)
			spewModel(c)
		} else {
			log.Printf("[MODEL] UPDATE CASHFLOW(%d)", c.ID)
			spewModel(c)
			c.Account.UpdateBalance(db, c)
			if c.HasSplits() {
				// TODO use BeforeUpdate hook to test if these fields changed
				c.updateSplits(db, c.splitUpdateMap())
			}
			if c.IsScheduled() {
				c.RepeatInterval.Update(db)
			}
		}

		// Create or save pair CashFlow if have one (Transfers)
		// Note, either side might be a Split
		if pair != nil {
			if pair.ID == 0 {
				db.Create(pair)
				c.CategoryID = pair.ID
				db.Model(c).Update("CategoryID", pair.ID)
				log.Printf("[MODEL] CREATE PAIR CASHFLOW(%d)", pair.ID)
			} else {
				db.Save(pair)
				log.Printf("[MODEL] UPDATE PAIR CASHFLOW(%d)", pair.ID)
			}

			if pair.mustUpdateBalance() {
				// if pair.Account changed, need two updates
				if pair.oldAccountID > 0 &&
				   pair.oldAccountID != pair.AccountID {
					newAccountUpdateAmount := pair.Amount
					pair.Amount = decimal.Zero
					pair.Account.ID = pair.oldAccountID
					pair.Account.UpdateBalance(db, pair)

					pair.oldAmount = decimal.Zero
					pair.Amount = newAccountUpdateAmount
				}
				pair.Account.ID = pair.AccountID
				pair.Account.UpdateBalance(db, pair)
			}
		}
	}
	return err
}

// Debug routines -

func CashFlowFind(db *gorm.DB, id uint) *CashFlow {
	c := new(CashFlow)
	db.First(&c, id)
	return c
}
