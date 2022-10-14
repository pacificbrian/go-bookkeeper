/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (c CashFlow) GetTransnum() string {
	if len(c.Transnum) > 8 {
		return ""
	}
	return c.Transnum
}

func (c *CashFlow) ParentID() uint {
	if !c.Split {
		return 0
	}
	return c.SplitFrom
}

func (c *CashFlow) RepeatParentID() uint {
	if c.IsScheduled() {
		return 0
	}
	return c.RepeatInterval.CashFlowID
}

func (c *CashFlow) CanSplit() bool {
	return !(c.Transfer || c.Split)
}

func (c *CashFlow) setSplit(SplitFrom uint) {
	if SplitFrom > 0 {
		if !c.IsScheduled() {
			c.Type = "SplitCashFlow"
		}
		c.Split = true
	}
	c.SplitFrom = SplitFrom
}

func (c *CashFlow) IsScheduled() bool {
	return c.Type == "RCashFlow"
}

func (c *CashFlow) IsScheduledParent() bool {
	return c.IsScheduled() && !c.Split
}

func (c *CashFlow) IsScheduledEnterable() bool {
	return (c.IsScheduledParent() && c.RepeatInterval.HasRepeatsLeft())
}

func (c *CashFlow) IsTrade() bool {
	return c.Type == "TradeCashFlow"
}

func (c *CashFlow) mustUpdateBalance() bool {
	// aka Base Type (!Split and !Repeat)
	return (c.Type ==  "" || c.IsTrade())
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

func (c *CashFlow) PreloadRepeat(db *gorm.DB) {
	if c.RepeatIntervalID > 0 {
		c.RepeatInterval.ID = c.RepeatIntervalID
		c.RepeatInterval.Preload(db)
	}
}

func (c *CashFlow) Preload(db *gorm.DB) {
	if c.IsTrade() {
		return
	}

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

	c.PreloadRepeat(db)
}

func mergeCashFlows(db *gorm.DB, A []CashFlow, B []CashFlow,
		    balance decimal.Decimal, limit int) []CashFlow {
	totalEntries := len(A) + len(B)
	var mergedEntries []CashFlow
	var a, b, c *CashFlow

	entries := &A
	if len(A) == 0 || len(B) == 0 {
		if len(A) == 0 {
			entries = &B
		}

		for i := 0; i < totalEntries; i++ {
			c = &(*entries)[i]
			c.Balance = balance
			balance = balance.Sub(c.Amount)
			c.Preload(db)
		}
	} else {
		aIdx := 0
		bIdx := 0
		// merge the 2 arrays together, keeping sorted by date
		mergedEntries = make([]CashFlow, totalEntries)
		for i := 0; i < totalEntries; i++ {
			a = nil
			b = nil
			if aIdx < len(A) {
				a = &A[aIdx]
			}
			if bIdx < len(B) {
				b = &B[bIdx]
			}

			if b == nil || (a != nil && a.Date.After(b.Date)) {
				c = a
				aIdx += 1
			} else {
				c = b
				bIdx += 1
			}

			c.Balance = balance
			balance = balance.Sub(c.Amount)
			c.Preload(db)
			mergedEntries[i] = *c
		}
		entries = &mergedEntries
	}

	if limit <= 0 || limit > totalEntries {
		limit = totalEntries
	}
	return (*entries)[0:limit]
}

func (c *CashFlow) Count(db *gorm.DB, account *Account) int64 {
	var count int64

	db.Model(c).Where(&CashFlow{AccountID: account.ID}).Count(&count)
	return count
}

// Account access already verified by caller
func (*CashFlow) ListMergeByDate(db *gorm.DB, account *Account, other []CashFlow,
				 date *time.Time) []CashFlow {
	entries := []CashFlow{}
	if !account.Verified {
		return entries
	}

	limit := account.User.UserSettings.CashFlowLimit

	// sort by Date
	// db.Order("date desc").Find(&entries, &CashFlow{AccountID: account.IDl})
	// use map to support NULL string
	query := map[string]interface{}{"account_id": account.ID, "type": nil}
	queryPrefix := db.Order("date desc")
	if date != nil {
		queryPrefix = queryPrefix.Where("date >= ?", date)
		limit = -1
	} else if limit > 0 {
		queryPrefix = queryPrefix.Limit(int(limit))
	}
	queryPrefix.Find(&entries, query)

	// merge if multiple CashFlow sets, update Balances
	entries = mergeCashFlows(db, entries, other, account.Balance, limit)

	log.Printf("[MODEL] LIST CASHFLOWS ACCOUNT(%d:%d)", account.ID, len(entries))
	return entries
}

func (*CashFlow) List(db *gorm.DB, account *Account) []CashFlow {
	return new(CashFlow).ListMergeByDate(db, account, nil, nil)
}

func (*CashFlow) ListByDate(db *gorm.DB, account *Account, date *time.Time) []CashFlow {
	return new(CashFlow).ListMergeByDate(db, account, nil, date)
}

func (*CashFlow) ListMerge(db *gorm.DB, account *Account, other []CashFlow) []CashFlow {
	return new(CashFlow).ListMergeByDate(db, account, other, nil)
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
	if c.Account.Verified {
		c.Account.User = *u
	}
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
	c.RepeatIntervalID = src.RepeatIntervalID
	c.Date = src.Date
	c.TaxYear = c.Date.Year()
	c.Memo = src.Memo
	c.Transnum = src.Transnum
	c.AccountID = src.AccountID
	c.Account.cloneVerified(&src.Account)
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
			oldPair.deletePair(db)
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
		result := db.Omit(clause.Associations).Create(c)
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
		db.Omit(clause.Associations).Model(parent).Update("split_from", gorm.Expr("split_from + ?", 1))
	} else {
		log.Printf("[MODEL] CREATE %s CASHFLOW(%d)", c.Type, c.ID)
		spewModel(c)
		c.Account.updateBalance(db, c)
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
		db.Omit(clause.Associations).Create(pair)
		c.CategoryID = pair.ID
		db.Model(c).Update("CategoryID", pair.ID)
		log.Printf("[MODEL] CREATE PAIR CASHFLOW(%d)", pair.ID)

		pair.Account.ID = pair.AccountID
		pair.Account.updateBalance(db, pair)
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
				db.Omit(clause.Associations).Model(split).
				   Updates(transferUpdates)
			} else {
				db.Omit(clause.Associations).Model(split).
				   Updates(updates)
			}
		}
	}
}

func (repeat *CashFlow) applyRate(db *gorm.DB) bool {
	repeat.Category.ID = repeat.CategoryID
	if !repeat.Category.IsInterestIncome() {
		return false
	}

	repeat.PreloadRepeat(db)
	if repeat.RepeatInterval.RepeatIntervalType.Days != 30 {
		return false
	}

	rate := &repeat.RepeatInterval.Rate
	if rate.Equal(decimal.Zero) {
		return false
	}

	monthlyRate := rate.Div(decimal.NewFromInt32(12))
	averageDailyBalance := repeat.Account.averageDailyBalance(db, repeat.Date)
	repeat.Amount = averageDailyBalance.Mul(monthlyRate).RoundBank(2)
	return true
}

// returns true if advanced date is still less than time.Now
func (repeat *CashFlow) advance(db *gorm.DB, updateAmount bool) bool {
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
	if updateAmount {
		updates["amount"] = repeat.Amount
	}
	db.Omit(clause.Associations).Model(repeat).Updates(updates)
	log.Printf("[MODEL] ADVANCE SCHEDULED CASHFLOW(%d) to %s", repeat.ID,
		   repeat.Date.Format("2006-01-02"))
	repeat.updateSplits(db, updates)

	return time.Now().After(repeat.Date)
}

func (repeat *CashFlow) tryInsertRepeatCashFlow(db *gorm.DB) error {
	var err error
	c := new(CashFlow)
	for {
		var newAmount bool
		if !c.Split {
			// no need to extend for Splits,
			// this uses repeat.Account.Balance
			newAmount = repeat.applyRate(db)
		}
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
			split.Account.cloneVerified(&repeat.Account)
			split.tryInsertRepeatCashFlow(db)
		}

		canRepeat := repeat.advance(db, newAmount)
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
		db.Omit(clause.Associations).Model(c).
		   Update("RepeatIntervalID", c.RepeatIntervalID)

		// mark Account as having ScheduledCashFlows
		c.Account.addScheduled(db)
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

	c.determineCashFlowType() // Edit only?
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
		if c.IsScheduled() {
			c.PreloadRepeat(db)
		}
	}

	return c
}

func (c *CashFlow) deletePair(db *gorm.DB) {
	// Clear Transfer flag so Pairs don't loop deleting each other
	c.Transfer = false
	c.delete(db)
}

func (c *CashFlow) deleteTransfer(db *gorm.DB) {
	if c.Transfer {
		pair := new(CashFlow)
		pair.pairFrom(c)
		pair.deletePair(db)
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
		c.deleteTransfer(db)
	} else {
		log.Printf("[MODEL] DELETE CASHFLOW(%d)", c.ID)
		if c.HasSplits() {
			splits, _ := c.ListSplit(db)
			for i := 0; i < len(splits); i++ {
				split := splits[i]
				split.delete(db)
			}
		}

		db.Delete(c)
		c.deleteTransfer(db)

		c.Account.ID = c.AccountID
		c.Amount = decimal.Zero
		// UpdateBalance will subtract c.oldAmount
		c.Account.updateBalance(db, c)
	}
}

func (c *CashFlow) Delete(db *gorm.DB) error {
	// Verify we have access to CashFlow
	c = c.Get(db, false)
	if c == nil {
		return errors.New("Permission Denied")
	}

	c.delete(db)
	return nil
}

func (c *CashFlow) Put(db *gorm.DB, request map[string]interface{}) error {
	// Verify we have access to CashFlow
	c = c.Get(db, false)
	if c == nil {
		return errors.New("Permission Denied")
	}

	jrequest, _ := json.Marshal(request)
	log.Printf("[MODEL] PUT CASHFLOW(%d) %s", c.ID, jrequest)

	if request["apply"] != nil {
		delete(request, "apply")
		if c.IsScheduledEnterable() {
			return c.tryInsertRepeatCashFlow(db)
		}
	}

	// special case c.Amount
	// need better way if expanded with more fields/types
	if request["amount"] != nil {
		newAmount, _ := strconv.ParseFloat(request["amount"].(string), 2)
		c.Amount = decimal.NewFromFloat(newAmount)
		if c.Amount.Equal(c.oldAmount) {
			// ignore non-update
			delete(request, "amount")
		} else {
			c.Account.updateBalance(db, c)
			// change type in map for db.Update to succeed
			request["amount"] = c.Amount
		}
	}

	if len(request) > 0 {
		db.Omit(clause.Associations).Model(c).Updates(request)
	}
	return nil
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
		result := db.Omit(clause.Associations).Save(c)
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
			c.Account.updateBalance(db, c)
			if c.HasSplits() {
				// TODO use BeforeUpdate hook to test if these fields changed
				c.updateSplits(db, c.splitUpdateMap())
			}
			if c.IsScheduled() {
				c.RepeatInterval.Update(db, c)
			}
		}

		// Create or save pair CashFlow if have one (Transfers)
		// Note, either side might be a Split
		if pair != nil {
			if pair.ID == 0 {
				db.Omit(clause.Associations).Create(pair)
				c.CategoryID = pair.ID
				db.Omit(clause.Associations).Model(c).
				   Update("CategoryID", pair.ID)
				log.Printf("[MODEL] CREATE PAIR CASHFLOW(%d)", pair.ID)
			} else {
				db.Omit(clause.Associations).Save(pair)
				log.Printf("[MODEL] UPDATE PAIR CASHFLOW(%d)", pair.ID)
			}

			if pair.mustUpdateBalance() {
				// if pair.Account changed, need two updates
				if pair.oldAccountID > 0 &&
				   pair.oldAccountID != pair.AccountID {
					newAccountUpdateAmount := pair.Amount
					pair.Amount = decimal.Zero
					pair.Account.ID = pair.oldAccountID
					pair.Account.updateBalance(db, pair)

					pair.oldAmount = decimal.Zero
					pair.Amount = newAccountUpdateAmount
				}
				pair.Account.ID = pair.AccountID
				pair.Account.updateBalance(db, pair)
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
