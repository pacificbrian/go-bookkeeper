/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"errors"
	"log"
	"time"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Trade struct {
	gorm.Model
	TradeTypeID uint `form:"trade_type_id"`
	oldTradeTypeID uint `gorm:"-:all"`
	AccountID uint `gorm:"not null"`
	oldAccountID uint `gorm:"-:all"`
	SecurityID uint `gorm:"not null"`
	ImportID uint
	Symbol string `form:"Symbol" gorm:"-:all"`
	Date time.Time
	oldDate time.Time `gorm:"-:all"`
	TaxYear int `form:"tax_year"`
	Amount decimal.Decimal `gorm:"not null"`
	oldAmount decimal.Decimal `gorm:"-:all"`
	Price decimal.Decimal
	Shares decimal.Decimal
	SharesSum decimal.Decimal `gorm:"-:all"`
	// AdjustedShares is remaining unsold shares, split adjusted
	AdjustedShares decimal.Decimal
	// Basis is accumulated (used) basis from Sells (starts at 0)
	// for Buys: remaining Basis is: Amount - Basis
	// (for AVGB, we still increment Buy specific (FIFO basis) in
	//  the buy.Basis; for actual used (average) basis, this is in
	//  the associated TradeGain.)
	// for Sells: the Gain (Loss) then is: Amount - Basis
	Basis decimal.Decimal
	BasisPS decimal.Decimal `gorm:"-:all"`
	Gain decimal.Decimal `gorm:"-:all"`
	oldGain decimal.Decimal `gorm:"-:all"`
	GainPS decimal.Decimal `gorm:"-:all"`
	oldShares decimal.Decimal `gorm:"-:all"`
	oldBasis decimal.Decimal `gorm:"-:all"`
	Tainted bool
	Closed bool
	TradeType TradeType
	Account Account
	Security Security
	TradeGains []TradeGain `gorm:"foreignKey:SellID"`
}

func (t *Trade) sanitizeInputs() {
	sanitizeString(&t.Symbol)
}

func (Trade) Currency(value decimal.Decimal) string {
	return currency(value)
}

func (t *Trade) IsBuy() bool {
	return (TradeTypeIsBuy(t.TradeTypeID) ||
	        TradeTypeIsReinvest(t.TradeTypeID))
}

func (t *Trade) IsCredit() bool {
	return (TradeTypeIsDividend(t.TradeTypeID) ||
	        TradeTypeIsDistribution(t.TradeTypeID))
}

func (t *Trade) IsReinvest() bool {
	return TradeTypeIsReinvest(t.TradeTypeID)
}

func (t *Trade) WasReinvest() bool {
	return TradeTypeIsReinvest(t.oldTradeTypeID)
}

func (t *Trade) IsSell() bool {
	return TradeTypeIsSell(t.TradeTypeID)
}

func (t *Trade) IsSharesIn() bool {
	return TradeTypeIsSharesIn(t.TradeTypeID)
}

func (t *Trade) IsSharesOut() bool {
	return TradeTypeIsSharesOut(t.TradeTypeID)
}

func (t *Trade) IsSplit() bool {
	return TradeTypeIsSplit(t.TradeTypeID)
}

func (t *Trade) IsAverageCost() bool {
	return SecurityBasisTypeIsAverage(t.Security.SecurityBasisTypeID)
}

func (t Trade) ViewIsSell() bool {
	return TradeTypeIsSell(t.TradeTypeID)
}

func (t Trade) GetBasis() string {
	if t.IsSell() {
		return "$" + t.Basis.StringFixed(2)
	} else if t.IsBuy() {
		return "$" + t.Amount.Sub(t.Basis).StringFixed(2)
	} else {
		return ""
	}
}

func (t *Trade) SharesRemaining() decimal.Decimal {
	if t.AdjustedShares.IsPositive() {
		return t.AdjustedShares
	}
	return t.Shares
}

func (t *Trade) getCashFlowType() uint {
	return TradeTypeToCashFlowType(t.TradeTypeID)
}

func (t *Trade) oldCashFlowType() uint {
	return TradeTypeToCashFlowType(t.oldTradeTypeID)
}

func (t *Trade) toCashFlow(preload bool) *CashFlow {
	cType := t.getCashFlowType()
	if cType == 0 {
		return nil
	}

	c := new(CashFlow)
	c.Type = "TradeCashFlow"
	c.AccountID = t.AccountID
	c.CashFlowTypeID = cType
	if !t.IsReinvest() {
		c.Amount = t.Amount
	}
	if !t.WasReinvest() {
		c.oldAmount = t.oldAmount
		// handle here unless CashFlow.oldCashFlowTypeID is added
		// and then can decide to move into applyCashFlowType()
		if t.oldCashFlowType() == Debit {
			c.oldAmount = c.oldAmount.Neg()
		}
	}
	c.Date = t.Date
	c.applyCashFlowType()
	c.CategoryID = t.TradeTypeID
	c.PayeeID = t.SecurityID
	c.ImportID = t.ID
	if preload {
		c.PayeeName = t.Security.Company.GetName()
		c.CategoryName = t.TradeType.Name
	}

	return c
}

func (t *Trade) Count(account *Account) int64 {
	db := getDbManager()
	var count int64

	db.Model(t).Where(&Trade{AccountID: account.ID}).Count(&count)
	return count
}

// Account Trades, Account access already verified by caller
// For Security Trades, use security.ListTrades
func (*Trade) List(db *gorm.DB, account *Account) []Trade {
	entries := []Trade{}
	if account.Verified {
		// Find Trades for Account
		db.Preload("TradeType").
		   Order("date asc").
		   Where(&Trade{AccountID: account.ID}).Find(&entries)
		log.Printf("[MODEL] LIST ACCOUNT(%d) TRADES(%d)", account.ID, len(entries))
	}
	return entries
}

func (t *Trade) totalGains(daysHeld uint) decimal.Decimal {
	if len(t.TradeGains) == 0 {
		return  t.Gain
	}

	gain := decimal.Zero
	for i := 0; i < len(t.TradeGains); i++ {
		tg := &t.TradeGains[i]
		tg.postQueryInit(t)
		if uint(tg.DaysHeld) >= daysHeld {
			log.Printf("[MODEL] TOTAL GAIN(%d) TRADE(%d) AMOUNT(%f) BASIS(%f)",
				   tg.ID, t.ID, tg.Amount.InexactFloat64(),
				   tg.Basis.InexactFloat64())
			gain = gain.Add(tg.Gain)
		}
	}
	return gain
}

// Filtered Account or User Trades for just single Year (t.Date.Year)
func (t *Trade) ListByType(session *Session, tradeType uint, daysHeld uint) ([]Trade, [2]decimal.Decimal) {
	var gain [2]decimal.Decimal
	entries := []Trade{}
	year := t.Date.Year()
	db := session.DB

	if year == 0 || TradeTypeQueries[tradeType] == "" {
		return entries, gain
	}

	if !t.Account.Verified && t.AccountID > 0 {
		// verify account against Session
		t.Account.ID = t.AccountID
		account := t.Account.Get(session, false)
		if account == nil {
			return entries, gain
		}
	}

	if t.AccountID > 0 {
		db.Preload("TradeType").Preload("Security.Company").
		   Order("date asc").
		   Where("date >= ? AND date < ?", t.Date, yearToDate(year+1)).
		   Where(TradeTypeQueries[tradeType]).
		   Where(&Trade{AccountID: t.AccountID}).Find(&entries)
	} else {
		dbPreload := db.Preload("TradeType").Preload("Security.Company")
		if daysHeld > 0 {
			dbPreload = dbPreload.Preload("TradeGains")
		}
		dbPreload.
		   Order("Account.Name").
		   Order("date asc").
		   Where("date >= ? AND date < ?", t.Date, yearToDate(year+1)).
		   Where(TradeTypeQueries[tradeType]).
		   Where("user_id = ?", session.GetUser().ID).
		   Joins("Account").Find(&entries)
	}

	for i := 0; i < len(entries); i++ {
		entry := &entries[i]
		entry.postQueryInit()
		gain[0] = gain[0].Add(entry.Gain)
		if entry.Account.Taxable {
			capGain := entry.totalGains(daysHeld)
			gain[1] = gain[1].Add(capGain)
		}
	}

	log.Printf("[MODEL] LIST ACCOUNT(%d) TRADES(%d:%d)",
		   t.AccountID, tradeType, len(entries))
	return entries, gain
}

func (t *Trade) ListByTypeTotal(session *Session, tradeType uint, daysHeld uint) [2]decimal.Decimal {
	_,total := t.ListByType(session, tradeType, daysHeld)
	return total
}

// Usage is interested in Trade Gains (Taxable), so here in returned CashFlows,
// we switch cf.Amount with the trade.Gain.
func (t *Trade) ListCashFlowByType(session *Session, tradeType uint) ([]CashFlow, decimal.Decimal) {
	entries := []CashFlow{}

	trades, total := t.ListByType(session, tradeType, 0)
	for i := 0; i < len(trades); i++ {
		t := trades[i]
		if !t.Account.Taxable {
			continue
		}
		cf := t.toCashFlow(true)
		if cf != nil {
			cf.Amount = t.Gain
			entries = append(entries, *cf)
		}
	}

	return entries, total[1]
}

// Account access already verified by caller
func (*Trade) listCashFlows(db *gorm.DB, account *Account, importID uint) []CashFlow {
	trades := []Trade{}
	entries := []CashFlow{}

	if account.Verified {
		// Find Trades for Account
		dbQuery := db.Preload("TradeType").
			      Preload("Security.Company").
			      Joins("Security").
			      Where(TradeTypeCashFlowsQuery)
		if (importID > 0) {
			dbQuery.Where(&Trade{AccountID: account.ID, ImportID: importID}).
			        Order("date asc").Find(&trades)
		} else {
			dbQuery.Where(&Trade{AccountID: account.ID}).
			        Order("date desc").Find(&trades)
		}

		log.Printf("[MODEL] LIST TRADES ACCOUNT(%d:%d)", account.ID, len(trades))

		for i := 0; i < len(trades); i++ {
			t := trades[i]
			cf := t.toCashFlow(true)
			if cf != nil {
				entries = append(entries, *cf)
			}
		}
	}
	return entries
}

func (t *Trade) ListCashFlows(db *gorm.DB, account *Account) []CashFlow {
	return t.listCashFlows(db, account, 0)
}

func (t *Trade) ListImportedCashFlows(im *Import) []CashFlow {
	db := getDbManager()
	return t.listCashFlows(db, &im.Account, im.ID)
}

func (t *Trade) gainBasisFIFO(soldShares decimal.Decimal) decimal.Decimal {
	sharesRemain := t.SharesRemaining()
	basis := t.Amount.Sub(t.Basis)
	if !sharesRemain.Equal(soldShares) {
		// must calculate using Basis per share
		basis = basis.Div(sharesRemain).Mul(soldShares).Round(2)
	}
	return basis
}

func (t *Trade) gainBasis(soldShares decimal.Decimal) decimal.Decimal {
	if t.IsAverageCost() {
		return t.Security.gainBasis(soldShares)
	} else {
		return t.gainBasisFIFO(soldShares)
	}
}

func (t *Trade) revertBasis(basis decimal.Decimal, soldShares decimal.Decimal) {
	db := getDbManager()
	updates := make(map[string]interface{})
	if t.IsBuy() {
		t.AdjustedShares = t.AdjustedShares.Add(soldShares)
		updates["adjusted_shares"] = t.AdjustedShares
		updates["closed"] = 0
	}

	if basis.IsPositive() {
		assert(t.Basis.GreaterThanOrEqual(basis),
		       "Revert Trade Basis Corruption (1)")
		t.Basis = t.Basis.Sub(basis)
		updates["basis"] = t.Basis
	}

	db.Omit(clause.Associations).Model(t).Updates(updates)
}

func (t *Trade) updateBasis(basis decimal.Decimal, soldShares decimal.Decimal) {
	db := getDbManager()
	updates := make(map[string]interface{})
	if t.IsBuy() {
		if t.AdjustedShares.IsZero() {
			assert(t.Basis.IsZero(), "Trade Basis Corrupted (1)")
			t.AdjustedShares = t.Shares
		}
		assert(t.AdjustedShares.GreaterThanOrEqual(soldShares),
		       "Update Trade Basis Corruption (1)")
		t.AdjustedShares = t.AdjustedShares.Sub(soldShares)
		updates["adjusted_shares"] = t.AdjustedShares

		t.Basis = t.Basis.Add(basis)
		if t.Amount.Equal(t.Basis) {
			assert(t.AdjustedShares.IsZero(), "Trade Basis Corrupted (2)")
			updates["closed"] = 1
		}
	} else {
		t.Basis = t.Basis.Add(basis)
	}
	updates["basis"] = t.Basis
	db.Omit(clause.Associations).Model(t).Updates(updates)
}

// t is Sell trade and was already tested to be Valid
func (t *Trade) recordGain(activeBuys []Trade) {
	sellBasis := decimal.Zero
	sellGain := decimal.Zero
	sharesRemain := t.Shares
	updateDB := true

	tg := new(TradeGain)
	for i := 0; sharesRemain.IsPositive(); i++ {
		buy := &activeBuys[i]
		buy.Security.clone(&t.Security)
		tg.ID = 0
		tg.recordGain(t, buy, sharesRemain, updateDB)
		sharesRemain = sharesRemain.Sub(tg.Shares)
		sellBasis = sellBasis.Add(tg.Basis)
		sellGain = sellGain.Add(tg.Gain)

		// update Basis in Buy
		buy.updateBasis(tg.BasisFIFO, tg.Shares)
	}

	// update Sell
	t.updateBasis(sellBasis, t.Shares)
	t.Gain = sellGain
}

// t is Split trade and was already tested to be Valid
func (t *Trade) recordSplit(activeBuys []Trade) {
	db := getDbManager()

	// update unsold Shares in Buys that are not yet closed
	for i := 0; i < len(activeBuys); i++ {
		buy := &activeBuys[i]
		buy.AdjustedShares = buy.AdjustedShares.Mul(t.Shares)
		db.Omit(clause.Associations).Model(buy).
		   Update("AdjustedShares", buy.AdjustedShares)
	}
}

// Look up Security by symbol, creates Security if none exists.
// Verifies Account Access, and returns nil if access denied.
func (t *Trade) securityGetBySymbol(session *Session) *Security {
	var security *Security

	if t.Symbol != "" {
		a := &t.Account
		a.ID = t.AccountID

		// verifies Account
		security,_ = a.GetSecurityBySymbol(session, t.Symbol)
		if security == nil {
			return nil
		} else if security.ID == 0 {
			err := security.create(session, true)
			if err != nil {
				return nil
			}
		}
		t.SecurityID = security.ID
	}

	return security
}

func (t *Trade) validateInputs() error {
	// ensure monetary amounts are 2 decimal places
	t.Amount = t.Amount.Round(2)

	if !TradeTypeIsValid(t.TradeTypeID) {
		return errors.New("Invalid Trade Type")
	} else if t.IsSell() || t.IsBuy() {
		if t.Amount.IsZero() || t.Price.IsZero() || t.Shares.IsZero() {
			return errors.New("Invalid Trade Entered (Buy/Sell)")
		}
	} else if t.IsSharesIn() || t.IsSharesOut() {
		if t.Shares.IsZero() {  // t.Amount (is optional)
			return errors.New("Invalid Trade Entered (Shares In/Out)")
		}
	} else if t.IsSplit() {
		if !t.Amount.IsZero() || t.Shares.IsZero() {
			return errors.New("Invalid Trade Entered (Split)")
		}
	}
	return nil
}

func (t *Trade) setDefaults() {
	if t.TaxYear == 0 {
		t.TaxYear = t.Date.Year()
	}
	if t.IsBuy() && t.Basis.IsZero() {
		t.AdjustedShares = t.Shares
	}
}

// security.Account access must be verified by caller,
// trade.Account should not be used here and assumed to be unset
func (t *Trade) insertTrade(db *gorm.DB, security *Security) error {
	var activeBuys []Trade
	var err error

	if !security.Account.Verified {
		log.Printf("[MODEL] INSERT TRADE PERMISSION DENIED")
		return errors.New("Permission Denied")
	}
	t.AccountID = security.AccountID

	err = t.validateInputs()
	if err == nil && (t.IsSell() || t.IsSplit()) {
		activeBuys, err = security.validateTrade(t)
	}
	if err != nil {
		return err
	}
	spewModel(t)

	result := db.Omit(clause.Associations).Create(t)
	log.Printf("[MODEL] CREATE TRADE(%d) SECURITY(%d) ACCOUNT(%d) TYPE(%d)",
		   t.ID, t.SecurityID, t.AccountID, t.TradeTypeID)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	if t.IsSell() {
		t.Security.clone(security)
		t.recordGain(activeBuys)
	} else if t.IsSplit() {
		t.recordSplit(activeBuys)
	}
	security.addTrade(t)
	c := t.toCashFlow(false)
	if c != nil {
		security.Account.updateBalance(c)
	}
	return nil
}

func (t *Trade) Create(session *Session) error {
	var security *Security
	db := session.DB

	if t.SecurityID > 0 {
		// verify access to Security
		t.Security.ID = t.SecurityID
		security = t.Security.Get(session)
	} else if t.AccountID > 0 {
		t.sanitizeInputs()
		// verifies Account, creates Security if none exists
		security = t.securityGetBySymbol(session)
	}

	if security == nil {
		return errors.New("Permission Denied")
	}
	t.ID = 0
	t.setDefaults()
	return t.insertTrade(db, security)
}

// t.Account must be preloaded
func (t *Trade) HaveAccessPermission(session *Session) bool {
	u := session.GetUser()
	t.Account.Verified = !(u == nil || t.Account.ID == 0 || u.ID != t.Account.UserID)
	if t.Account.Verified {
		t.Account.User = *u
		t.Account.Session = session
		t.Security.Account.cloneVerified(&t.Account)
	}
	return t.Account.Verified
}

func (t *Trade) postQueryInit() {
	t.oldDate = t.Date
	t.oldAmount = t.Amount
	t.oldBasis = t.Basis
	t.oldShares = t.Shares
	t.oldTradeTypeID = t.TradeTypeID
	if t.IsSell() {
		t.Gain = t.Amount.Sub(t.Basis)
		t.GainPS = t.Gain.Div(t.Shares)
		t.BasisPS = t.Basis.Div(t.Shares)
	} else if t.IsCredit() {
		t.Gain = t.Amount
	}
	t.oldGain = t.Gain
}

// Edit, Delete, Update use Get
func (t *Trade) Get(session *Session) *Trade {
	db := session.DB
	if t.ID > 0 {
		db.Preload("TradeType").Preload("Account").
		   Preload("Security.Company").Joins("Security").
		   First(&t)
	}

	// Verify we have access to Trade
	if !t.HaveAccessPermission(session) {
		return nil
	}

	// This is safe and will only set NULL fields to default values.
	// Cannot seem to use GORM hook as NULL value would already be
	// converted to 0. (This is for working with old database).
	t.setDefaults()
	// need another query if AdjustedShares is NULL and Basis is set
	//db.Where("adjusted_shares IS NOT NULL").First(&nullTest, t.ID)

	// for Update, store old values before overwritten
	t.postQueryInit()
	return t
}

func (t *Trade) reverseGain(isDelete bool) error {
	db := getDbManager()
	sellBasis := decimal.Zero
	entries := []TradeGain{}

	// validate is newest Sell Trade
	sell := t.Security.LatestTradeBy(db, Sell)
	if sell == nil || sell.ID != t.ID {
		return errors.New("Only support reversing of newest Sell Trades!")
	}

	// Find Gains for Trade, update Basis in Buys
	db.Where(&TradeGain{SellID: t.ID}).Find(&entries)
	for i := 0; i < len(entries); i++ {
		tg := &entries[i]
		sellBasis = sellBasis.Add(tg.Basis)
		tg.Delete(t.Account.Session)
	}

	// update Basis in Sell (don't bother if Trade will be deleted)
	if !isDelete {
		t.revertBasis(sellBasis, t.Shares)
	}

	log.Printf("[MODEL] REVERSED TRADE(%d) AND %d GAINS", t.ID, len(entries))
	return nil
}

// TODO
func (t *Trade) reverseSplit() ([]Trade, error) {
	err := t.Security.validateSplit(t)
	return nil, err
}

func (t *Trade) updateGains() {
	db := getDbManager()
	entries := []TradeGain{}

	duration := t.Date.Sub(t.oldDate)
	days := int32(duration.Hours()) / 24

	if !t.IsSell() || days == 0 {
		return
	}

	// Find Gains for Trade, update DaysHeld
	db.Where(&TradeGain{SellID: t.ID}).Find(&entries)
	for i := 0; i < len(entries); i++ {
		tg := &entries[i]
		tg.updateDaysHeld(days)
	}

	log.Printf("[MODEL] UPDATE TRADE(%d) DAYS(%d) FOR %d GAINS",
		   t.ID, days, len(entries))
}

func (t *Trade) Delete(session *Session) error {
	var err error
	db := getDbManager()

	// Verify we have access to Trade
	t = t.Get(session)
	if t == nil {
		return errors.New("Permission Denied")
	}

	// set these to Zero, updateTrade/Balance will reverse Trade
	t.Amount = decimal.Zero
	t.Shares = decimal.Zero

	if t.IsSell() {
		err = t.reverseGain(true)
	} else if t.IsBuy() && !t.oldBasis.IsZero() {
		err = errors.New("Don't yet support Delete of Partially Sold Buy Trades!")
	} else if t.IsSplit() {
		_, err = t.reverseSplit()
		err = errors.New("Don't yet support Delete of Splits!")
		// set Shares to 1 for reversing Split in updateTrade
		t.Shares = decimal.NewFromInt32(1)
	}
	if err != nil {
		log.Printf("[MODEL] DELETE TRADE(%d) UNSUPPORTED: %v", t.ID, err)
		return err
	}

	t.Security.updateTrade(t)
	c := t.toCashFlow(false)
	if c != nil {
		t.Account.updateBalance(c)
	}
	spewModel(t)
	db.Delete(t)
	log.Printf("[MODEL] DELETE TRADE(%d)", t.ID)
	return nil
}

func (t *Trade) isSimpleUpdate() bool {
	return t.oldAmount.Equal(t.Amount) && t.oldBasis.Equal(t.Basis) &&
	       t.oldShares.Equal(t.Shares) && t.oldTradeTypeID == t.TradeTypeID
}

// Trade access already verified with Get
func (t *Trade) Update() error {
	var activeBuys []Trade
	var err error
	db := getDbManager()
	logSimple := ""

	if !t.Account.Verified {
		return errors.New("!Account.Verified")
	}

	err = t.validateInputs()
	if err != nil {
		return err
	}
	spewModel(t)

	isSimple := t.isSimpleUpdate()
	if isSimple {
		// (Price, Date) only, no validation needed
		// Though, some risk (with Date) that Trades are now reordered
		// and Split or TradeGain is now wrong...
		logSimple = " (SIMPLE)"
		if t.IsSell() {
			// if Date changed, update TradeGains.DaysHeld
			t.updateGains()
		}
	} else if t.IsSell() {
		err = t.reverseGain(false)
		if err == nil {
			activeBuys, err = t.Security.validateTrade(t)
		}
	} else if t.IsBuy() && !t.oldBasis.IsZero() {
		err = errors.New("Don't yet support Updating of Partially Sold Buy Trades!")
	} else if t.IsBuy() && !t.oldShares.Equal(t.AdjustedShares) {
		err = errors.New("Don't yet support Updating of Buy Trades affected by Splits!")
	} else if t.IsSplit() {
		activeBuys, err = t.reverseSplit()
		err = errors.New("Don't yet support Updating of Splits!")
	} else if t.IsBuy() {
		// this becomes more complicated when/if removing above error cases
		t.AdjustedShares = t.Shares
	}
	if err != nil {
		log.Printf("[MODEL] UPDATE TRADE(%d) UNSUPPORTED: %v", t.ID, err)
		return err
	}

	result := db.Omit(clause.Associations).Save(t)
	err = result.Error
	if err == nil && !isSimple {
		if t.IsSell() && activeBuys != nil {
			t.recordGain(activeBuys)
		} else if t.IsSplit() && activeBuys != nil {
			t.recordSplit(activeBuys)
		}

		t.Security.updateTrade(t)
		c := t.toCashFlow(false)
		if c != nil {
			t.Account.updateBalance(c)
		}
	}
	if err == nil {
		log.Printf("[MODEL] UPDATE%s TRADE(%d) SECURITY(%d) ACCOUNT(%d) TYPE(%d)",
			   logSimple, t.ID, t.SecurityID, t.AccountID, t.TradeTypeID)
	}

	return err
}


// Debug routines -

// Find() for use with rails/ruby like REPL console (gomacro);
// controllers should not expose this as are no access controls
func (*Trade) Find(ID uint) *Trade {
	db := getDbManager()
	t := new(Trade)
	db.First(&t, ID)
	t.Account.Verified = true
	t.postQueryInit()
	return t
}

func (t *Trade) UpdateAdjustedShares(fShares float64) {
	soldShares := decimal.NewFromFloat(fShares)
	t.updateBasis(decimal.Zero, soldShares)
}

func (t *Trade) Save() error {
	db := getDbManager()
	result := db.Omit(clause.Associations).Save(t)
	return result.Error
}

func (t *Trade) Print() {
	forceSpewModel(t.Model, 0)
	forceSpewModel(t, 1)
}

func (t *Trade) PrintAll() {
	forceSpewModel(t, 0)
}
