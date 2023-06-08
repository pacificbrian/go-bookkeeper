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

type CashFlow struct {
	gorm.Model
	CashFlowTypeID uint `form:"cash_flow_type_id" gorm:"-:all"`
	AccountID uint `gorm:"not null"`
	oldAccountID uint `gorm:"-:all"`
	TaxYear int `form:"tax_year"`
	Date time.Time
	oldDate time.Time
	Amount decimal.Decimal `gorm:"not null"`
	oldAmount decimal.Decimal `gorm:"-:all"`
	Balance decimal.Decimal `gorm:"-:all"`
	PayeeID uint `gorm:"not null"` // also serves as Pair.AccountID (Transfers)
	CategoryID uint `form:"category_id"` // also serves as Pair.ID (Transfers)
	oldPayeeID uint `gorm:"-:all"`
	PairID uint `gorm:"-:all"`
	ImportID uint
	RepeatIntervalID uint
	SplitFrom uint
	Split bool
	Transfer bool
	PayeeName string `form:"payee_name" gorm:"-:all"`
	CategoryName string `gorm:"-:all"`
	Memo string `form:"memo"`
	Transnum string `form:"transnum"`
	Type string `gorm:"default:NULL"`
	Account Account
	Category Category
	Payee Payee
	RepeatInterval RepeatInterval
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

func (c *CashFlow) IsCredit() bool {
	return CashFlowTypeIsCredit(c.CashFlowTypeID)
}

func (c *CashFlow) IsDebit() bool {
	return CashFlowTypeIsDebit(c.CashFlowTypeID)
}

// if SplitCashFlow, get parent CashFlow.ID
func (c *CashFlow) ParentID() uint {
	if !c.Split {
		return 0
	}
	return c.SplitFrom
}

// get ScheduledCashFlow.ID that is assocated with object
func (c *CashFlow) RepeatParentID() uint {
	if c.IsScheduled() {
		return 0
	}

	// For applied CashFlows, RepeatIntervalID is ID of the
	// origin ScheduledCashFlow; it is not a RepeatInterval.ID.
	// See cloneScheduled().
	return c.RepeatIntervalID
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

func (c *CashFlow) IsScheduledEnterable(allowFutureNoRepeat bool) bool {
	if allowFutureNoRepeat {
		return c.IsScheduledParent() && c.Date.After(time.Now())
	}
	return c.IsScheduledParent() && c.RepeatInterval.HasRepeatsLeft()
}

func (c *CashFlow) IsTrade() bool {
	return c.Type == "TradeCashFlow"
}

func (c CashFlow) IsSellTrade() bool {
	return c.Type == "TradeCashFlow" &&
	       TradeTypeIsSell(c.CategoryID)
}

func (c CashFlow) ShowTradeLinks() bool {
	return c.IsTrade()
}

func (c *CashFlow) mustUpdateBalance() bool {
	// aka Base Type (!Split and !Repeat)
	return (c.Type ==  "" || c.IsTrade())
}

func (c *CashFlow) getSession() *Session {
	if !c.Account.Verified {
		return nil
	}
	return c.Account.Session
}

// Used with CreateSplitCashFlow. Controller calls to get common CashFlow
// fields first, and before Bind (which can/will override other fields).
func NewSplitCashFlow(session *Session, SplitFrom uint) (*CashFlow, int) {
	c := new(CashFlow)
	c.ID = SplitFrom
	// Get will set Date, PayeeID, Type = RCashFlow
	c = c.Get(session, false)
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

func (c *CashFlow) setCategoryName(db *gorm.DB) {
	if c.HasSplits() {
		c.CategoryName = "Split"
	} else if c.CategoryID > 0 {
		c.CategoryName = c.Category.Name
		if c.CategoryName == "" && c.Account.Verified {
			c.CategoryName = c.Account.User.lookupCategoryName(c.CategoryID)
		}
		if c.CategoryName == "" {
			c.Category.ID = c.CategoryID
			db.First(&c.Category)
			if c.Account.Verified {
				c.Account.User.cacheCategoryName(&c.Category)
			}
			c.CategoryName = c.Category.Name
		}
	}
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
		assert(c.PayeeID > 0, "CashFlow.Preload: (Transfer) no PayeeID!")
		if c.Account.Verified {
			c.PayeeName = c.Account.User.lookupAccountName(c.PayeeID)
		}
		if c.PayeeName == "" {
			a := new(Account)
			db.First(&a, c.PayeeID)
			c.PayeeName = a.Name
			if c.Account.Verified && a.ID > 0 {
				c.Account.User.cacheAccountName(a)
			}
		}
		c.CategoryName = "Transfer"
	} else {
		c.PayeeName = c.Payee.Name
		if c.PayeeName == "" && c.PayeeID > 0 {
			db.First(&c.Payee, c.PayeeID)
			c.PayeeName = c.Payee.Name
		}
		c.setCategoryName(db)
	}

	if c.IsScheduledParent() {
		c.PreloadRepeat(db)
	}
}

// Merges two sorted CashFlow arrays into one, with computed Balances.
// Can be sorted ascending (default) or descending.
// Account is needed if setting preload boolean, or if setting descending
// boolean, but otherwise does not need to be valid Account.
func (ca *Account) mergeCashFlows(db *gorm.DB, A []CashFlow, B []CashFlow,
				  limit int, descending bool, preload bool) []CashFlow {
	totalEntries := len(A) + len(B)
	balance := decimal.Zero
	logTime := false
	var mergedEntries []CashFlow
	var a, b, c *CashFlow

	if logTime {
		log.Printf("[MODEL] ACCOUNT(%d) CASHFLOW MERGE/PRELOAD START", ca.ID)
	}

	if descending {
		balance = ca.CashBalance
	}

	entries := &A
	if len(A) == 0 || len(B) == 0 {
		if len(A) == 0 {
			entries = &B
		}

		for i := 0; i < totalEntries; i++ {
			c = &(*entries)[i]
			if descending {
				c.Balance = balance
				balance = balance.Sub(c.Amount)
			} else {
				balance = balance.Add(c.Amount)
				c.Balance = balance
			}
			if preload {
				c.Account.cloneVerified(ca)
				c.Preload(db)
			}
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

			if b == nil || (a != nil && dateFirst(&a.Date, &b.Date, descending)) {
				c = a
				aIdx += 1
			} else {
				c = b
				bIdx += 1
			}

			if descending {
				c.Balance = balance
				balance = balance.Sub(c.Amount)
			} else {
				balance = balance.Add(c.Amount)
				c.Balance = balance
			}
			if preload {
				c.Account.cloneVerified(ca)
				c.Preload(db)
			}
			mergedEntries[i] = *c
		}
		entries = &mergedEntries
	}
	if logTime {
		log.Printf("[MODEL] ACCOUNT(%d) CASHFLOW MERGE/PRELOAD STOP", ca.ID)
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
	var entries []CashFlow
	if !account.Verified {
		return entries
	}

	limit := account.User.UserSettings.CashFlowLimit

	// sort by Date
	// db.Order("date desc").Find(&entries, &CashFlow{AccountID: account.IDl})
	// use map to support NULL string
	query := map[string]interface{}{"account_id": account.ID, "type": nil}
	queryPrefix := db.Order("date desc").Preload("Payee").Preload("Category")
	if date != nil {
		queryPrefix = queryPrefix.Where("date >= ?", date)
		limit = -1
	} else if limit > 0 {
		queryPrefix = queryPrefix.Limit(int(limit))
	}
	queryPrefix.Find(&entries, query)

	// merge if multiple CashFlow sets, update Balances
	entries = account.mergeCashFlows(db, entries, other, limit, true, true)

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
			// sets CashFlowType, old values, PairID (Transfers)
			split.postQueryInit()
			split.Account.cloneVerified(&c.Account)
			split.Preload(db)
			total = total.Add(split.Amount)
		}
	}
	return entries, currency(total)
}

func (u *User) listTaxCategory(db *gorm.DB, year int, taxCat *TaxCategory,
			       wantEntries bool) ([]CashFlow, decimal.Decimal) {
	var total decimal.Decimal
	var entries []CashFlow

	if taxCat.CategoryID > 0 {
		db.Preload("CashFlows",
			   func(db *gorm.DB) *gorm.DB {
				return db.Order("date asc").
					  Where("(type != ? OR type IS NULL) AND tax_year = ? AND category_id = ?",
						"RCashFlow", year, taxCat.CategoryID)
			   }).
		   Find(&u.Accounts, &Account{UserID: u.ID, Taxable: true})
		for i := 0; i < len(u.Accounts); i++ {
			a := &u.Accounts[i]
			for j := 0; j < len(a.CashFlows); j++ {
				c := &a.CashFlows[j]
				c.determineCashFlowType()
				total = total.Add(c.Amount)
			}

			if (wantEntries) {
				//entries = append(entries, *c)
				a.setSession(u.Session)
				entries = a.mergeCashFlows(db, entries, a.CashFlows,
                                                           0, false, true)
			}
		}
	}
	return entries, total
}

func (u *User) ListTaxCategory(db *gorm.DB, year int, taxCat *TaxCategory) ([]CashFlow, decimal.Decimal) {
	return u.listTaxCategory(db, year, taxCat, true)
}

func (u *User) ListTaxCategoryTotal(db *gorm.DB, year int, taxCat *TaxCategory) decimal.Decimal {
	_, total  := u.listTaxCategory(db, year, taxCat, false)
	return total
}

// c.Account must be preloaded
func (c *CashFlow) HaveAccessPermission(session *Session) bool {
	u := session.GetUser()
	c.Account.Verified = !(u == nil || c.Account.ID == 0 || u.ID != c.Account.UserID)
	if c.Account.Verified {
		c.Account.User = *u
		c.Account.Session = session
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
	// ensure monetary amounts are 2 decimal places
	c.Amount = c.Amount.Round(2)
}

// set CashFlowType to be correct, and backup some values used to
// optimize out unneeded database updates
func (c *CashFlow) postQueryInit() {
	c.oldAmount = c.Amount
	c.oldDate = c.Date
	c.oldPayeeID = c.PayeeID
	c.determineCashFlowType()
	if c.Transfer {
		// backup CategoryID as cleared by Bind (during Update)
		// and needed to find Pair during updateSplits
		// ScheduledCashFlows don't have Pairs, so unused in that case
		c.PairID = c.CategoryID // Peer Cashflow (Transfers)
		c.CategoryID = 0
	}
}

func (c *CashFlow) cloneScheduled(src *CashFlow) {
	c.Transfer = src.Transfer
	if src.Split {
		c.setSplit(src.SplitFrom)
	}
	c.RepeatIntervalID = src.ID
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

// Using src CashFlow, construct the Pair (other side of a Transfer).
// This is used during Create or Update.
func (c *CashFlow) cloneTransfer(src *CashFlow) {
	c.Transfer = true
	c.Date = src.Date
	c.TaxYear = src.TaxYear
	c.Memo = src.Memo
	c.Transnum = src.Transnum
	c.Account.cloneVerifiedFrom(&src.Account)
	c.oldAccountID = c.AccountID // used if Update
	c.AccountID = src.PayeeID
	c.PayeeID = src.AccountID
	c.CategoryID = src.ID
	c.oldAmount = c.Amount // used if Update
	c.Amount = src.Amount.Neg()
}

// Using src CashFlow, construct the Pair (other side of a Transfer).
// This is used during Update, Put, or Delete and only fill in minimum.
// Update will call cloneTransfer to complete Pair CashFlow.
// We can even reconstruct Splits (careful: old transactions in DB
// don't store this!)
func (c *CashFlow) pairFrom(src *CashFlow) {
	c.Transfer = true
	c.ID = src.PairID
	// keep split details accurate, and decrement SplitCount in Parent (Delete)
	c.setSplit(src.SplitFrom)
	c.Account.cloneVerifiedFrom(&src.Account)
	c.AccountID = src.PayeeID
	c.Amount = src.oldAmount.Neg()
	c.oldAmount = c.Amount // used if Delete or Put
}

// prepare CashFlow to write to DB (used by both Create and Update)
//   - update Amount and Transfer based on CashFlowType
//   -- (above done earlier (moved to Create/Update functions))
//   - create Payee if needed
//   - lookup Account (error if not found/accessible)
//   - return Pair cashflow (for other Account) if this is a Transfer
//   - UPDATEs are allowed to change to/from Transfer type and change Peer Account
func (c *CashFlow) prepareInsertCashFlow(db *gorm.DB, importing bool) (error, *CashFlow) {
	var pair *CashFlow = nil // Transfer Pair

	if c.Transfer {
		var a *Account

		if c.PayeeName != "" {
			a = accountGetByName(c.getSession(), c.PayeeName)
		} else if c.PayeeID > 0 {
			// retrieve Account for scheduled CashFlows
			a = new(Account)
			a.ID = c.PayeeID
			a = a.Get(c.getSession(), false)
		}

		if a == nil {
			return errors.New("Account.Name Invalid"), nil
		} else {
			// store pair account.ID in PayeeID (aka TransferAccountID)
			c.PayeeID = a.ID
			c.CategoryID = 0
		}
		if !c.IsScheduled() {
			// create pair CashFlow
			pair = new(CashFlow)
			if c.PairID > 0 {
				// #UPDATE: use existing pair CashFlow
				c.CategoryID = c.PairID
				pair.pairFrom(c)
			}

			// fill in pair CashFlow with remaining details
			pair.cloneTransfer(c)
			// #UPDATE: if pair.AccountID changed, this is handled in caller
		}
	} else {
		// #UPDATE: if Transfer type True->False, delete pair CashFlow
		if c.PairID > 0 {
			oldPair := new(CashFlow)
			oldPair.pairFrom(c)
			oldPair.deletePair(db)
		}

		if !c.Split && c.PayeeName != "" {
			// creates Payee if none exists
			err, p := payeeGetByName(c.getSession(), c.PayeeName, importing)
			if err != nil {
				return err, nil
			}
			c.PayeeID = p.ID
		}
	}

	return nil, pair
}

// c.Account access must be verified
func (c *CashFlow) insertCashFlow(db *gorm.DB, importing bool) error {
	if !c.Account.Verified {
		log.Printf("[MODEL] INSERT CASHFLOW PERMISSION DENIED")
		return errors.New("Permission Denied")
	}
	err, pair := c.prepareInsertCashFlow(db, importing)
	if err == nil {
		result := db.Omit(clause.Associations).Create(c)
		err = result.Error
	}
	if err != nil {
		log.Printf("[MODEL] INSERT CASHFLOW ERROR: %s", err)
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
		db.Omit(clause.Associations).Model(c).Update("CategoryID", pair.ID)
		log.Printf("[MODEL] CREATE PAIR CASHFLOW(%d)", pair.ID)

		pair.Account.ID = pair.AccountID
		pair.Account.updateBalance(db, pair)
	}

	return err
}

func (c *CashFlow) repeatUpdateMap() map[string]interface{} {
	updates:= map[string]interface{}{"date": c.Date, "tax_year": c.TaxYear}
	// prune date if not changed
	if compareDates(&c.oldDate, &c.Date) {
		delete(updates, "date")
		delete(updates, "tax_year")
	}

	// this is to repair old database
	if len(updates) > 0 && c.IsScheduledParent() {
		updates["type"] = "RCashFlow"
	}

	log.Printf("[MODEL] CASHFLOW(%d) REPEAT UPDATES MAP LEN(%d)", c.ID, len(updates))
	return updates
}

// map of fields that must be equivalent in Split/Parent
// when applied to split.Transfer, payee_id is pruned out later
func (c *CashFlow) splitUpdateMap() map[string]interface{} {
	updates := map[string]interface{}{"date": c.Date, "tax_year": c.TaxYear,
				          "payee_id": c.PayeeID}
	// prune if not changed
	if compareDates(&c.oldDate, &c.Date) {
		delete(updates, "date")
		delete(updates, "tax_year")
	}
	if c.oldPayeeID == c.PayeeID {
		delete(updates, "payee_id")
	}

	// this is to repair old database
	if len(updates) > 0 && c.IsScheduledParent() {
		updates["type"] = "RCashFlow"
	}

	log.Printf("[MODEL] CASHFLOW(%d) SPLIT UPDATES MAP LEN(%d)", c.ID, len(updates))
	return updates
}

// update only selected fields in Splits from the given map
func updateSplits(db *gorm.DB, splits []CashFlow, updates map[string]interface{},
		  testAmount bool) {
	// for Transfers, copy map and remove payee_id
	// wish there was cleaner way
	transferUpdates := make(map[string]interface{})
	for k,v := range updates {
		transferUpdates[k] = v
	}
	delete(transferUpdates, "payee_id")

	for i := 0; i < len(splits); i++ {
		split := &splits[i]

		if testAmount {
			if !split.oldAmount.Equal(split.Amount) {
				log.Printf("[MODEL] UPDATE SPLIT (%d) AMOUNT (%f) (%f)",
					   split.ID, split.oldAmount.InexactFloat64(),
					   split.Amount.InexactFloat64())
				updates["amount"] = split.Amount
				transferUpdates["amount"] = split.Amount
				split.oldAmount = split.Amount
			} else {
				delete(updates, "amount")
				delete(transferUpdates, "amount")
			}
		}
		if split.Transfer && len(transferUpdates) > 0 {
			db.Omit(clause.Associations).Model(split).
			   Updates(transferUpdates)
			if transferUpdates["date"] != nil && !split.IsScheduled() &&
			   transferUpdates["type"] == nil {
				assert(transferUpdates["amount"] == nil,
				       "updateSplits: update split.Amount unexpected")
				// update Pair.Date
				pair := new(CashFlow)
				pair.pairFrom(split)
				db.Omit(clause.Associations).Model(pair).
				   Updates(transferUpdates)
			}
		} else if len(updates) > 0 {
			db.Omit(clause.Associations).Model(split).
			   Updates(updates)
		}
	}
}

func (c *CashFlow) updateSplits(db *gorm.DB, updates map[string]interface{}) {
	if c.HasSplits() && len(updates) > 0 {
		splits, _ := c.ListSplit(db)
		updateSplits(db, splits, updates, false)
	}
}

func (repeat *CashFlow) getMonthlyRate(db *gorm.DB) decimal.Decimal {
	repeat.PreloadRepeat(db)
	if repeat.RepeatInterval.RepeatIntervalType.Days != 30 {
		return decimal.Zero
	}

	// returns decimal.Zero if not set
	rate := repeat.RepeatInterval.Rate
	rate = rate.Div(decimal.NewFromInt32(100))
	return rate.Div(decimal.NewFromInt32(12))
}

func (repeat *CashFlow) applyRate(db *gorm.DB) bool {
	repeat.Category.ID = repeat.CategoryID
	if !repeat.Category.IsInterestIncome() {
		return false
	}

	monthlyRate := repeat.getMonthlyRate(db)
	if !monthlyRate.IsPositive() { // positive Rates only
		return false
	}

	averageDailyBalance := repeat.Account.averageDailyBalance(db, repeat.Date)
	repeat.Amount = averageDailyBalance.Mul(monthlyRate).RoundBank(2)
	return true
}

func (repeat *CashFlow) calculateLoanPI(db *gorm.DB) ([]CashFlow, bool) {
	var paymentCF *CashFlow
	var principleCF *CashFlow
	var interestCF *CashFlow
	var splits []CashFlow
	var fees decimal.Decimal
	debugPI := false
	matched := 0

	if !repeat.HasSplits() {
		return splits, false
	}

	monthlyRate := repeat.getMonthlyRate(db)
	if !monthlyRate.IsPositive() { // positive Rates only
		return splits, false
	}

	// iterate over Splits to find P and I payments
	updateAmounts := true
	splits, _ = repeat.ListSplit(db)
	for i := 0; i < len(splits); i++ {
		var split *CashFlow = &splits[i]
		split.Category.ID = split.CategoryID
		if split.Transfer {
			if split.IsCredit() {
				/* both are Credits */
				paymentCF = split
				principleCF = repeat
			} else {
				/* both are Debits, flip to Credit (reversed below) */
				paymentCF = repeat
				principleCF = split
				paymentCF.Amount = paymentCF.Amount.Neg()
				principleCF.Amount = principleCF.Amount.Neg()
			}
			assert(principleCF.Amount.IsPositive(), "LoanPI: bad Principle Amount")
			assert(paymentCF.Amount.IsPositive(), "LoanPI: bad Payment Amount")
			matched += 1
		} else if split.Category.LoanPI() {
			interestCF = split
			matched += 1
		} else {
			// other fixed fees
			fees = fees.Add(split.Amount)
		}
	}

	if (matched != 2 || principleCF == nil || interestCF == nil) {
		updateAmounts = false
		// cannot return as may have to flip Credits back to Debits
	}

	// we have valid ScheduledCashFlow for determining P and I
	for updateAmounts {
		averageDailyBalance := interestCF.Account.averageDailyBalance(db, repeat.Date)
		// This is loan, so we should return if loan amount is somehow non-negative
		if !averageDailyBalance.IsNegative() {
			updateAmounts = false
			break
		}

		// record new Amounts in returned SplitCashFlows,
		// takes affect as applied next in caller's main logic
		interestCF.Amount = averageDailyBalance.Mul(monthlyRate).RoundBank(2)
		assert(interestCF.Amount.IsNegative(), "LoanPI: bad Interest Amount")
		principleCF.Amount = paymentCF.Amount.Add(fees).Add(interestCF.Amount)
		if debugPI {
			log.Printf("[MODEL] CPI INTEREST CASHFLOW(%d) (%f)", interestCF.ID,
				   interestCF.Amount.InexactFloat64())
			log.Printf("[MODEL] CPI PRINCIPLE CASHFLOW(%d) (%f)", principleCF.ID,
				   principleCF.Amount.InexactFloat64())
		}
		break
	}

	// ensure Amount is positive (Credits) or negative (Debits)
	if principleCF != nil {
		paymentCF.applyCashFlowType()
		principleCF.applyCashFlowType()
	}
	return splits, updateAmounts
}

// returns true if advanced date is still less than time.Now
func (repeat *CashFlow) advance(db *gorm.DB, updateDB bool) (bool, int) {
	days := repeat.RepeatInterval.advance(db)
	if days == 0 {
		return false, days
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

	if updateDB {
		updates := repeat.repeatUpdateMap()
		if !repeat.oldAmount.Equal(repeat.Amount) {
			updates["amount"] = repeat.Amount
		}
		db.Omit(clause.Associations).Model(repeat).Updates(updates)
	}
	log.Printf("[MODEL] ADVANCE SCHEDULED CASHFLOW(%d) to %s", repeat.ID,
		   repeat.Date.Format("2006-01-02"))

	return time.Now().After(repeat.Date), days
}

func (c CashFlow) insertRepeatSplits(splits []CashFlow, ch chan uint) {
	count := <- ch
	log.Printf("[MODEL] APPLYING REPEAT SPLITS (%d)", count)
	for i := 0; i < len(splits); i++ {
		split := &splits[i]
		split.SplitFrom = c.ID
		split.tryInsertRepeatCashFlow()
	}
	ch <- count + 1
}

func (repeat *CashFlow) tryInsertRepeatCashFlow() (decimal.Decimal, error) {
	var amountAdded decimal.Decimal
	var splits []CashFlow
	var err error
	db := getDbManager()
	updateDB := true

	// channel so background Splits processed in order
	splitsChan := make(chan uint, 1)
	chanSignaled := false
	chanPending := 0

	c := new(CashFlow)
	for {
		var newSplitAmounts bool

		// Below handles when ScheduledCashFlow has RepeatInterval.Rate
		// no need to extend use of Rate for Splits
		if !repeat.Split {
			// Backup repeat.Date, repeat.Amount and call each iteration
			// thru this loop in case Date, Amount changed.
			// For Splits, this happens in ListSplit and we don't want
			// to call again here as would reset modified split.Amount
			// before we call updateSplits.
			repeat.postQueryInit()

			// logic here requires repeat.Account.Balance
			applied := repeat.applyRate(db)
			if applied {
				log.Printf("[MODEL] INTEREST RATE APPLIED")
			} else {
				splits, newSplitAmounts = repeat.calculateLoanPI(db)
				if newSplitAmounts {
					log.Printf("[MODEL] CALCULATE PI APPLIED (%d)",
						   len(splits))
				}
			}
		}
		c.cloneScheduled(repeat)

		if updateDB {
			// add scheduled CashFlow
			err = c.insertCashFlow(db, false)
			if err != nil || c.Split {
				break
			}
			amountAdded = amountAdded.Add(c.Amount)

			// reuse Splits array if queried above
			if len(splits) == 0 {
				splits, _ = repeat.ListSplit(db)
			}

			// now add Repeat's SplitCashFlows (in background)
			if len(splits) > 0 {
				var splitsCopy = make([]CashFlow, len(splits))

				if chanPending> 0 && !chanSignaled {
					// If multiple goroutines, it's nice
					// if they run in order, so signal to
					// start executing the first goroutine
					// here.
					splitsChan <- 1
					chanSignaled = true
				}
				// If single goroutine, it is started later
				// before we return, as we'd like to defer
				// associated database operations.
				chanPending += 1
				copy(splitsCopy, splits)
				go c.insertRepeatSplits(splitsCopy, splitsChan)
			}
		}

		// advance Date in Repeat CashFlow and Splits, but reuse
		// array of Splits we already queried
		canRepeat,_ := repeat.advance(db, updateDB)
		if updateDB && len(splits) > 0 {
			updateSplits(db, splits, repeat.repeatUpdateMap(),
				     newSplitAmounts)
		}
		if !canRepeat {
			break
		}
		c.ID = 0
	}

	// signal to start addng SplitCashFlows if not started above
	if !chanSignaled {
		splitsChan <- 1
	}

	return amountAdded, err
}

// defaults for DB fields not set during Create (are Edit only)
func (c *CashFlow) setDefaults() {
	c.TaxYear = c.Date.Year()
	if c.Transfer {
		// set via PayeeName, clear for NewSplit+Create
		c.PayeeID = 0
	}
}

func (c *CashFlow) Create(session *Session) error {
	db := session.DebugDB
	// Verify we have access to Account
	if !c.Account.Verified {
		c.Account.ID = c.AccountID
		account := c.Account.Get(session, false)
		if account == nil {
			return errors.New("Permission Denied")
		}
	}

	c.applyCashFlowType()
	// defaults for DB fields not set during Create (are Edit only)
	c.setDefaults()

	err := c.insertCashFlow(db, false)
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

// Edit, Delete, Update, NewSplit use Get
// c.Account needs to be preloaded
func (c *CashFlow) Get(session *Session, edit bool) *CashFlow {
	db := session.DB
	if c.ID > 0 {
		if edit {
			db.Preload("Payee").Preload("Category").
			   Preload("Account").First(&c)
		} else {
			db.Preload("Account").First(&c)
		}
	}

	// Verify we have access to CashFlow
	if !c.HaveAccessPermission(session) {
		return nil
	}

	// sets CashFlowType, old values, PairID (Transfers)
	c.postQueryInit()

	if edit {
		// some Preloads done above at start of Get()
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
	if c.ID > 0 {
		c.delete(db)
	}
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
		db.Omit(clause.Associations).Model(parent).
		   Update("split_from", gorm.Expr("split_from - ?", 1))

		db.Delete(c)
		c.deleteTransfer(db)
	} else {
		log.Printf("[MODEL] DELETE CASHFLOW(%d)", c.ID)
		if c.HasSplits() {
			splits, _ := c.ListSplit(db)
			for i := 0; i < len(splits); i++ {
				split := &splits[i]
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

func (c *CashFlow) Delete(session *Session) error {
	db := session.DB
	// Verify we have access to CashFlow
	c = c.Get(session, false)
	if c == nil {
		return errors.New("Permission Denied")
	}

	c.delete(db)
	return nil
}

func (c *CashFlow) Put(session *Session, request map[string]interface{}) error {
	db := session.DB
	// Verify we have access to CashFlow
	c = c.Get(session, false)
	if c == nil {
		return errors.New("Permission Denied")
	}

	jrequest, _ := json.Marshal(request)
	log.Printf("[MODEL] PUT CASHFLOW(%d) %s", c.ID, jrequest)

	if request["apply"] != nil {
		delete(request, "apply")
		if c.IsScheduledEnterable(true) {
			_,err := c.tryInsertRepeatCashFlow()
			return err
		}
	}

	// special case c.Amount
	// need better way if expanded with more fields/types
	if request["amount"] != nil {
		newAmount, _ := strconv.ParseFloat(request["amount"].(string), 2)
		c.Amount = decimal.NewFromFloatWithExponent(newAmount, -2)
		if c.Amount.Equal(c.oldAmount) {
			// ignore non-update
			delete(request, "amount")
		} else {
			if c.Transfer {
				pair := new(CashFlow)
				pair.pairFrom(c)
				pair.Amount = c.Amount.Neg()
				pair.Account.ID = pair.AccountID
				pair.Account.updateBalance(db, pair)

				// change type in map for db.Update to succeed
				request["amount"] = pair.Amount
				db.Omit(clause.Associations).Model(pair).Updates(request)
			}

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
func (c *CashFlow) Update() error {
	db := getDbManager()

	if !c.Account.Verified {
		return errors.New("!Account.Verified")
	}

	c.applyCashFlowType()
	if c.Split {
		// don't let Splits mess with date
		c.Date = c.oldDate
	}

	err, pair := c.prepareInsertCashFlow(db, false)
	if err == nil {
		result := db.Omit(clause.Associations, "type").Save(c)
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
				c.updateSplits(db, c.splitUpdateMap())
				// above also updates Pair.Date (Transfers)
			}
			if c.IsScheduled() {
				c.RepeatInterval.StartDay = c.Date.Day()
				c.RepeatInterval.Update()
			}
		}

		// Create or save pair CashFlow if have one (Transfers)
		// Note, either side might be a Split
		if pair != nil {
			if pair.ID == 0 {
				pair.CategoryID = c.ID
				db.Omit(clause.Associations).Create(pair)
				c.CategoryID = pair.ID
				db.Omit(clause.Associations).Model(c).
				   Update("CategoryID", pair.ID)
				log.Printf("[MODEL] CREATE PAIR CASHFLOW(%d)", pair.ID)
			} else {
				db.Omit(clause.Associations, "type").Save(pair)
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

// Find() for use with rails/ruby like REPL console (gomacro);
// controllers should not expose this as are no access controls
func (*CashFlow) Find(ID uint) *CashFlow {
	db := getDbManager()
	c := new(CashFlow)
	db.First(&c, ID)
	c.postQueryInit()
	return c
}

func (c *CashFlow) Print() {
	forceSpewModel(c.Model, 0)
	forceSpewModel(c, 1)
}

func (c *CashFlow) PrintAll() {
	forceSpewModel(c, 0)
}
