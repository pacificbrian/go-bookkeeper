/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"errors"
	"log"
	"strings"
	"time"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SecurityValue struct {
	Basis decimal.Decimal
	Shares decimal.Decimal
	Value decimal.Decimal
	AccumulatedBasis decimal.Decimal
	RetainedEarnings decimal.Decimal
}

type Security struct {
	Model
	CompanyID uint `gorm:"not null"`
	SecurityBasisTypeID uint `form:"security_basis_type_id"`
	oldSecurityBasisTypeID uint `gorm:"-:all"`
	SecurityTypeID uint `form:"security_type_id"`
	AccountID uint `gorm:"not null"`
	ImportName string `form:"security.ImportName"`
	BasisFromTrades decimal.Decimal `gorm:"-:all"`
	SecurityValue
	lastQuoteUpdate time.Time
	Account Account
	Company Company
	SecurityType SecurityType
}

func (s *Security) sanitizeInputs() {
	s.Company.sanitizeInputs()
	sanitizeString(&s.ImportName)
}

func (Security) Currency(value decimal.Decimal) string {
	return currency(value)
}

func (s Security) Price() decimal.Decimal {
	if s.Shares.Equal(decimal.Zero) {
		return decimal.Zero
	} else {
		return s.Value.DivRound(s.Shares, 2)
	}
}

func (s Security) BasisPrice() decimal.Decimal {
	if s.Shares.IsZero() {
		return decimal.Zero
	} else {
		return s.Basis.DivRound(s.Shares, 2)
	}
}

// if using average cost basis
func (s *Security) gainBasis(soldShares decimal.Decimal) decimal.Decimal {
	assert(s.Shares.GreaterThanOrEqual(soldShares),
	       "gainBasis: Security Shares Corrupted")
	return s.Basis.Div(s.Shares).Mul(soldShares).Round(2)
}

func (s *Security) UnrealizedGain() decimal.Decimal {
	if s.Basis.IsZero() {
		return decimal.Zero
	}
	return s.Value.Sub(s.Basis)
}

func (s Security) UnrealizedReturn() decimal.Decimal {
	paperGains := s.UnrealizedGain()
	if paperGains.IsZero() {
		return decimal.Zero
	}
	simpleReturn := paperGains.DivRound(s.Basis, 4)
	return decimalToPercentage(simpleReturn)
}

func (s Security) TotalReturn() decimal.Decimal {
	if s.RetainedEarnings.IsZero() || s.AccumulatedBasis.IsZero() {
		return s.UnrealizedReturn()
	}
	earned := s.RetainedEarnings.Add(s.UnrealizedGain())
	simpleReturn := earned.DivRound(s.AccumulatedBasis, 4)
	return decimalToPercentage(simpleReturn)
}

func (s *Security) HasFilings() bool {
	return SecurityTypeHasFilings[s.SecurityTypeID] &&
	       s.Company.HasFilings()
}

func (s *Security) clone(from *Security) {
	s.ID = from.ID
	s.Account.Verified = from.Account.Verified
	s.SecurityBasisTypeID = from.SecurityBasisTypeID
	s.SecurityValue = from.SecurityValue
}

func (*Security) sanitizeSecurityName(securityName string) string {
	subName := strings.Split(securityName, "(")[0]
	if subName != "" {
		return subName
	}
	return securityName
}

func (s *Security) setValue(price decimal.Decimal) decimal.Decimal {
	s.Value = s.Shares.Mul(price).Round(2)
	return s.Value
}

func (s *Security) addTrade(db *gorm.DB, trade *Trade) {
	updates := make(map[string]interface{})
	price := s.Price()

	// if trade.Date is newer than last time we pushed a Quote to database,
	// use given Price, else use Price computed above
	if trade.Price.IsPositive() && trade.Date.After(s.lastQuoteUpdate) {
		price = trade.Price
	}

	// just for 'old' Securities in database
	if s.AccumulatedBasis.IsZero() && !s.Basis.IsZero() {
		s.AccumulatedBasis = s.Basis
		updates["accumulated_basis"] = s.AccumulatedBasis
	}

	if trade.IsSell() {
		// don't assert s.Basis as may trip on rounding issues
		assert(s.Shares.GreaterThanOrEqual(trade.Shares),
		       "addTrade: Security Shares Corrupted")
		s.Shares = s.Shares.Sub(trade.Shares)
		if s.Shares.IsZero() {
			s.Basis = decimal.Zero
		} else {
			s.Basis = s.Basis.Sub(trade.Basis)
		}
		s.RetainedEarnings = s.RetainedEarnings.Add(trade.Gain)
		updates["basis"] = s.Basis
		updates["retained_earnings"] = s.RetainedEarnings
		updates["shares"] = s.Shares
	} else if trade.IsBuy() {
		s.AccumulatedBasis = s.AccumulatedBasis.Add(trade.Amount)
		s.Basis = s.Basis.Add(trade.Amount)
		s.Shares = s.Shares.Add(trade.Shares)
		updates["accumulated_basis"] = s.AccumulatedBasis
		updates["basis"] = s.Basis
		updates["shares"] = s.Shares
		if trade.IsReinvest() {
			s.RetainedEarnings = s.RetainedEarnings.Add(trade.Amount)
			updates["retained_earnings"] = s.RetainedEarnings
		}
	} else if trade.IsCredit() {
		s.RetainedEarnings = s.RetainedEarnings.Add(trade.Amount)
		updates["retained_earnings"] = s.RetainedEarnings
	} else if trade.IsSharesIn() {
		s.Shares = s.Shares.Add(trade.Shares)
		updates["shares"] = s.Shares
	} else if trade.IsSharesOut() {
		s.Shares = s.Shares.Sub(trade.Shares)
		updates["shares"] = s.Shares
	} else if trade.IsSplit() {
		s.Shares = s.Shares.Mul(trade.Shares)
		// value doesn't change for Split
		price = decimal.Zero
		updates["shares"] = s.Shares
	} else if !trade.Price.IsPositive() {
		return
	}

	// update Security Value as Shares changed;
	// if we sold all Shares, will update Value to Zero
	if !price.IsZero() {
		updates["value"] = s.setValue(price)
	}
	db.Omit(clause.Associations).Model(s).Updates(updates)
	log.Printf("[MODEL] SECURITY(%d) ADD TRADE (%d) TYPE(%d)",
		   s.ID, trade.ID, trade.TradeTypeID)
}

func (s *Security) updateTrade(db *gorm.DB, trade *Trade) {
	updates := make(map[string]interface{})
	price := s.Price()

	if trade.IsSell() {
		s.Basis = s.Basis.Add(trade.oldBasis)
		s.Basis = s.Basis.Sub(trade.Basis)
		s.RetainedEarnings = s.RetainedEarnings.Sub(trade.oldGain)
		s.RetainedEarnings = s.RetainedEarnings.Add(trade.Gain)
		s.Shares = s.Shares.Add(trade.oldShares)
		s.Shares = s.Shares.Sub(trade.Shares)
		updates["basis"] = s.Basis
		updates["retained_earnings"] = s.RetainedEarnings
		updates["shares"] = s.Shares
	} else if trade.IsBuy() {
		s.AccumulatedBasis = s.AccumulatedBasis.Sub(trade.oldAmount)
		s.AccumulatedBasis = s.AccumulatedBasis.Add(trade.Amount)
		s.Basis = s.Basis.Sub(trade.oldAmount)
		s.Basis = s.Basis.Add(trade.Amount)
		s.Shares = s.Shares.Sub(trade.oldShares)
		s.Shares = s.Shares.Add(trade.Shares)
		updates["accumulated_basis"] = s.AccumulatedBasis
		updates["basis"] = s.Basis
		updates["shares"] = s.Shares
		if trade.IsReinvest() {
			s.RetainedEarnings = s.RetainedEarnings.Sub(trade.oldAmount)
			s.RetainedEarnings = s.RetainedEarnings.Add(trade.Amount)
			updates["retained_earnings"] = s.RetainedEarnings
		}
	} else if trade.IsCredit() {
		s.RetainedEarnings = s.RetainedEarnings.Sub(trade.oldAmount)
		s.RetainedEarnings = s.RetainedEarnings.Add(trade.Amount)
		updates["retained_earnings"] = s.RetainedEarnings
	} else if trade.IsSharesIn() {
		s.Shares = s.Shares.Sub(trade.oldShares)
		s.Shares = s.Shares.Add(trade.Shares)
		updates["shares"] = s.Shares
	} else if trade.IsSharesOut() {
		s.Shares = s.Shares.Add(trade.oldShares)
		s.Shares = s.Shares.Sub(trade.Shares)
		updates["shares"] = s.Shares
	} else if trade.IsSplit() {
		s.Shares = s.Shares.Div(trade.oldShares)
		s.Shares = s.Shares.Mul(trade.Shares)
		// value doesn't change for Split
		price = decimal.Zero
		updates["shares"] = s.Shares
	} else {
		return
	}

	// update Security Value as Shares changed;
	// if we sold all Shares, will update Value to Zero
	if !price.IsZero() {
		updates["value"] = s.setValue(price)
	}
	db.Omit(clause.Associations).Model(s).Updates(updates)
	log.Printf("[MODEL] SECURITY(%d) UPDATE TRADE (%d) TYPE(%d)",
		   s.ID, trade.ID, trade.TradeTypeID)
}

// goroutine: this fetches latest Price and updates cached Quotes.
// It should not access the database.
func updateSecurities(securities *[]Security) {
	for i := 0; i < len(*securities); i++ {
		s := &(*securities)[i]
		if s.Shares.IsPositive() {
			s.fetchPrice(false)
		}
	}
}

// update Security.Values and update Account.Balance from cached Quotes
func (a *Account) getSecurities(openPositions bool, async bool) {
	updated := false
	db := getDbManager()
	if !a.Verified || !a.IsInvestment() {
		return
	}

	// Find Securities for Account
	if (openPositions) {
		db.Order("Company.Symbol").
		   Where("shares > 0 AND account_id = ?", a.ID).
		   Joins("Company").
		   Find(&a.Securities)
	} else {
		db.Order("Company.Symbol").
		   Where(&Security{AccountID: a.ID}).
		   Joins("Company").
		   Find(&a.Securities)
	}

	if async {
		// initiate background fetching of Security Quotes
		go updateSecurities(&a.Securities)
	} else {
		log.Printf("[MODEL] GET SECURITIES ACCOUNT(%d:%d)", a.ID,
			   len(a.Securities))
		updateSecurities(&a.Securities)
	}

	for i := 0; i < len(a.Securities); i++ {
		entry := &a.Securities[i]
		updated = entry.postQueryInit() || updated
	}
	if updated {
		a.updateValue(true)
	}
}

func (a *Account) getOpenSecurities(async bool) {
	a.getSecurities(true, async)
}

// with account argument, Account access already verified by caller
func (s *Security) List(session *Session, account *Account, openPositions bool) []Security {
	entries := []Security{}
	if account == nil {
		// Verify we have access to Account
		s.Account.ID = s.AccountID
		account = s.Account.Get(session, false)
		if account == nil  {
			return entries
		}
	}
	account.getSecurities(openPositions, true)

	log.Printf("[MODEL] LIST SECURITIES ACCOUNT(%d:%d)", account.ID,
		   len(account.Securities))
	return account.Securities
}

// Find only most recent Trade for Security
func (s *Security) LatestTradeBy(db *gorm.DB, tradeType uint) *Trade {
	entry := &Trade{}

	if s.Account.Verified {
		dbQuery := db.Order("date desc").Order("id desc")
		if tradeType > 0 {
			dbQuery = dbQuery.Where(TradeTypeQueries[tradeType])
		}

		dbQuery.Where(&Trade{SecurityID: s.ID}).
			First(&entry)
	}
	log.Printf("[MODEL] LIST SINGLE TRADE SECURITY(%d:%d)", s.ID, entry.ID)
	return entry
}

// Find Trades for Security
// Security access already verified by caller
func (s *Security) ListTradesBy(db *gorm.DB, tradeType uint, openOnly bool) []Trade {
	entries := []Trade{}

	if s.Account.Verified {
		dbQuery := db.Order("date asc").Preload("TradeType")
		if tradeType > 0 {
			if openOnly {
				dbQuery = dbQuery.Where("closed = 0")
			}
			dbQuery = dbQuery.Where(TradeTypeQueries[tradeType])
		}
		dbQuery.Where(&Trade{SecurityID: s.ID}).
			Find(&entries)

		s.fixupTrades(db, entries)
	}
	log.Printf("[MODEL] LIST TRADES SECURITY(%d:%d)", s.ID, len(entries))
	return entries
}

func (s *Security) ListTrades(db *gorm.DB) []Trade {
	return s.ListTradesBy(db, 0, false)
}

func (s *Security) computeShares(db *gorm.DB) decimal.Decimal {
	var shares decimal.Decimal
	var sharesInOut decimal.Decimal
	includeInOut := true

	activeBuys := s.ListTradesBy(db, Buy, true)
	for i := 0; i < len(activeBuys); i++ {
		shares = shares.Add(activeBuys[i].SharesRemaining())
	}

	if includeInOut {
		trades := s.ListTradesBy(db, SharesIn, true)
		for i := 0; i < len(trades); i++ {
			sharesInOut = sharesInOut.Add(trades[i].Shares)
		}
		trades = s.ListTradesBy(db, SharesOut, false)
		for i := 0; i < len(trades); i++ {
			sharesInOut = sharesInOut.Sub(trades[i].Shares)
		}
	}

	return shares.Add(sharesInOut)
}

func (s *Security) validateSell(db *gorm.DB, trade *Trade) ([]Trade, error) {
	var buyShares decimal.Decimal

	activeBuys := s.ListTradesBy(db, Buy, true)
	if len(activeBuys) == 0 {
		return nil, errors.New("Invalid Sell Trade (No Shares)")
	}

	for i := 0; i < len(activeBuys); i++ {
		buy := &activeBuys[i]
		if !buy.Date.After(trade.Date) {
			buyShares = buyShares.Add(activeBuys[i].SharesRemaining())
		}
	}
	if buyShares.LessThan(trade.Shares) {
		return nil, errors.New("Invalid Sell Trade (Insufficient Shares)")
	}

	return activeBuys, nil
}

// TODO validate there are no Sells or Splits on/after trade.Date
func (s *Security) validateSplit(db *gorm.DB, trade *Trade) error {
	return nil
}

func (s *Security) validateTrade(db *gorm.DB, trade *Trade) ([]Trade, error) {
	if trade.IsSell() {
		return s.validateSell(db, trade)
	} else if trade.IsSplit() {
		err := s.validateSplit(db, trade)
		if err != nil {
			return nil, err
		}

		activeBuys := s.ListTradesBy(db, Buy, true)
		if len(activeBuys) == 0 {
			return nil, errors.New("Ignoring Split (No Shares)")
		}
		return activeBuys, nil
	}
	return nil, nil
}

func (s *Security) validateInputs() error {
	if !SecurityBasisTypeIsValid(s.SecurityBasisTypeID) {
		return errors.New("Invalid Security Basis Type")
	}
	if !SecurityTypeIsValid(s.SecurityTypeID) {
		return errors.New("Invalid Security Type")
	}
	return nil
}

func (s *Security) init() {
	s.SecurityTypeID = 1 // Default is Stock
	s.SecurityBasisTypeID = 1 // Default is FIFO
}

func (s *Security) create(session *Session, useDefaults bool) error {
	db := session.DB
	if useDefaults {
		s.init()
	}

	// Verify we have access to Account
	s.Account.ID = s.AccountID
	account := s.Account.Get(session, false)
	if account == nil {
		return errors.New("Permission Denied")
	}
	if !account.IsInvestment() {
		return errors.New("Invalid Account")
	}

	err := s.validateInputs()
	if err != nil {
		return err
	}
	spewModel(s)

	result := db.Omit(clause.Associations).Create(s)
	log.Printf("[MODEL] CREATE SECURITY(%d) ACCOUNT(%d)", s.ID, s.AccountID)
	if result.Error != nil {
		log.Fatal(result.Error)
	}
	return result.Error
}

func (s *Security) Create(session *Session) error {
	s.sanitizeInputs()
	s.Account.ID = s.AccountID
	existing,err := s.Account.GetSecurity(session, &s.Company)
	if existing == nil {
		return err
	} else if existing.ID != 0 {
		return errors.New("Security Already Exists")
	}
	s.CompanyID = existing.CompanyID

	s.ID = 0
	return s.create(session, false)
}

// s.Account must be preloaded
func (s *Security) HaveAccessPermission(session *Session) bool {
	u := session.GetUser()
	s.Account.Verified = !(u == nil || s.Account.ID == 0 || u.ID != s.Account.UserID)
	if s.Account.Verified {
		s.Account.User = *u
		s.Account.Session = session
	}
	return s.Account.Verified
}

func (s *Security) updateValue(debugValue bool) bool {
	updated := false
	// don't update when no Shares
	if s.Company.Symbol == "" || s.Shares.IsZero() ||
	   GetQuoteCache() == nil {
		return updated
	}

	quote := GetQuoteCache().Get(s.Company.Symbol)
	if quote.Price.IsPositive() {
		s.setValue(quote.Price)
		updated = true
	}

	if debugValue {
		log.Printf("[MODEL] SECURITY(%d:%s) UPDATE VALUE(%t:%f) (%f)",
			   s.ID, s.Company.Symbol, updated,
			   s.Value.InexactFloat64(), quote.Price.InexactFloat64())
	}
	return updated
}

func (s *Security) postQueryInit() bool {
	debugValue := false

	s.Company.oldSymbol = s.Company.Symbol
	s.Company.oldName = s.Company.Name
	s.oldSecurityBasisTypeID = s.SecurityBasisTypeID

	// update s.Account.Balance to cached value
	if s.Account.ID > 0 {
		s.Account.postQueryInit()
	}

	// updates s.Value (if have Shares) from latest Quote
	return s.updateValue(debugValue)
}

// controllers(Get, Edit, Delete, Update) use Get
func (s *Security) Get(session *Session) *Security {
	db := session.DB
	debugShares := false

	if s.ID > 0 {
		db.Preload("SecurityType").Preload("Company").
		   Preload("Account").First(&s)
	}
	// Verify we have access to Security
	if !s.HaveAccessPermission(session) {
		return nil
	}
	s.postQueryInit()

	if debugShares {
		log.Printf("[MODEL] GET SECURITY(%d:%s) SHARES(%f:%f)",
			   s.ID, s.Company.Symbol, s.Shares.InexactFloat64(),
			   s.computeShares(db).InexactFloat64())
	} else {
		log.Printf("[MODEL] GET SECURITY(%d:%s)", s.ID, s.Company.Symbol)
	}
	return s
}

func (s *Security) Delete(session *Session) error {
	db := session.DB
	// Verify we have access to Security
	s = s.Get(session)
	if s != nil {
		spewModel(s)
		db.Delete(s)
		return nil
	}
	return errors.New("Permission Denied")
}

func (s *Security) hasTaintedBuys() bool {
	return false
}

// Security access already verified with Get
func (s *Security) Update() error {
	db := getDbManager()
	updatedCompany := false

	s.sanitizeInputs()
	err := s.validateInputs()
	if err != nil {
		return err
	}
	spewModel(s)

	// for 'old' Securities in database
	if s.AccumulatedBasis.IsZero() {
		s.AccumulatedBasis = s.Basis
	}

	// verify safe to go AVGB -> FIFO
	if s.oldSecurityBasisTypeID != s.SecurityBasisTypeID &&
	   s.SecurityBasisTypeID == 1 && s.hasTaintedBuys() {
		s.SecurityBasisTypeID = s.oldSecurityBasisTypeID
	}

	s.Company.UserID = s.Account.UserID
	updatedCompany = s.Company.Update()
	// test if Company changed, and must update CompanyID
	if updatedCompany {
		s.CompanyID = s.Company.ID
	}

	result := db.Omit(clause.Associations).Save(s)
	err = result.Error
	if err != nil {
		return err
	}

	log.Printf("[MODEL] UPDATE SECURITY(%d:%s)", s.ID, s.Company.Symbol)
	return nil
}

func (s *Security) ChangeAccount(session *Session, name string) *Account {
	a := GetAccountByName(session, name)
	if a != nil {
		if !a.IsInvestment() {
			return nil
		}
		s.AccountID = a.ID

		db := session.DB
		db.Omit(clause.Associations).Model(s).
		   Update("account_id", a.ID)
		log.Printf("[MODEL] UPDATE SECURITY(%d:%s) MOVED TO(%d)",
			   s.ID, s.Company.Symbol, a.ID)

		// update cached Account.Balances to reflect Move
		if !s.Value.IsZero() {
			s.Account.Balance = s.Account.Balance.Sub(s.Value)
			a.User.cacheAccountBalance(&s.Account)
			a.Balance = a.Balance.Add(s.Value)
			a.User.cacheAccountBalance(a)
		}
	}
	return a
}

func (s *Security) fixupTrades(db *gorm.DB, entries []Trade) {
	fixSharesIn := false
	fixAdjustedBasis := false
	sharesSum := decimal.Zero

	for i := 0; i < len(entries); i++ {
		t := &entries[i]
		if t.IsSell() || t.IsSharesOut() {
			sharesSum = sharesSum.Sub(t.Shares)
		        s.BasisFromTrades = s.BasisFromTrades.Sub(t.Basis)
		} else if t.IsSplit() {
			sharesSum = sharesSum.Mul(t.Shares)
		} else {
			sharesSum = sharesSum.Add(t.Shares)
		        s.BasisFromTrades = s.BasisFromTrades.Add(t.Amount)
		}
		t.SharesSum = sharesSum
	}

	if !fixSharesIn && !fixAdjustedBasis {
		return
	}

	for i := 0; i < len(entries); i++ {
		t := &entries[i]

		if fixSharesIn && t.IsSharesIn() {
			db.Omit(clause.Associations).Model(t).
			   Update("trade_type_id", 5)
		}

		if fixAdjustedBasis && !t.Closed && t.IsBuy() &&
		    t.Basis.IsPositive() && t.AdjustedShares.IsZero() {
			t.Account.cloneVerified(&s.Account)
			gains,_ := t.ListGains(db)
			if len(gains) == 1 {
				tg:= gains[0]
				t.AdjustedShares = t.Shares.Sub(tg.Shares)
				db.Omit(clause.Associations).Model(t).
				   Update("adjusted_shares", t.AdjustedShares)
			}
		}
	}
}


// Debug routines -

// Find() for use with rails/ruby like REPL console (gomacro);
// controllers should not expose this as are no access controls
func (*Security) Find(ID uint) *Security {
	db := getDbManager()
	s := new(Security)
	db.First(&s, ID)
	s.Account.Verified = true
	s.postQueryInit()
	return s
}

func (s *Security) Print() {
	forceSpewModel(s.Model, 0)
	forceSpewModel(s, 1)
}

func (s *Security) PrintAll() {
	forceSpewModel(s, 0)
}
