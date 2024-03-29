/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"errors"
	"log"
	"strconv"
	"time"
	"github.com/shopspring/decimal"
	"github.com/pacificbrian/go-bookkeeper/config"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Account struct {
	gorm.Model
	UserID uint `gorm:"not null"`
	AccountTypeID uint `form:"account.account_type_id"`
	CurrencyTypeID uint `form:"account.currency_type_id"`
	InstitutionID uint `form:"account.institution_id"`
	AverageBalance decimal.Decimal `gorm:"-:all"`
	Balance decimal.Decimal
	CashBalance decimal.Decimal
	Portfolio SecurityValue `gorm:"-:all"`
	Routing int `form:"account.Routing"`
	OfxIndex uint `form:"account.OfxIndex"`
	ClientUID string `form:"account.ClientUID"`
	Name string `form:"account.Name"`
	Number string `form:"account.Number"`
	HasScheduled bool
	Hidden bool `form:"account.Hidden"`
	Taxable bool `form:"account.Taxable"`
	Verified bool `gorm:"-:all"`
	Session *Session `gorm:"-:all"`
	AccountType AccountType
	CurrencyType CurrencyType
	Institution Institution
	User User
	CashFlows []CashFlow
	Securities []Security
}

func (a *Account) sanitizeInputs() {
	sanitizeString(&a.Name)
	sanitizeString(&a.Number)
}

func (Account) Currency(value decimal.Decimal) string {
	return "$" + value.StringFixedBank(2)
}

// for Bind() and setting from input/checkboxes */
func (a *Account) ClearBooleans() {
	a.Taxable = false
	a.Hidden = false
}

func (a *Account) GetRouting() string {
	if a.Routing == 0 {
		return ""
	}
	return strconv.Itoa(a.Routing)
}

func (a *Account) IsInvestment() bool {
	a.AccountType.ID = a.AccountTypeID
	return a.AccountType.isCrypto() ||
	       a.AccountType.isInvestment()
}

func (a *Account) SupportsDownload(requireID bool) bool {
	a.AccountType.ID = a.AccountTypeID
	return a.AccountType.supportsDownload() &&
	       (!requireID || a.InstitutionID > 0)
}

func (a Account) PortfolioTotalReturn() decimal.Decimal {
	if a.Portfolio.Basis.IsZero() {
		return decimal.Zero
	}
	simpleReturn := a.Portfolio.Value.Sub(a.Portfolio.Basis).
					  DivRound(a.Portfolio.Basis, 4)
	return decimalToPercentage(simpleReturn)
}

func (a *Account) TotalPortfolio(securities []Security) {
	for i := 0; i < len(securities); i++ {
		security := &securities[i]
		a.Portfolio.Basis = a.Portfolio.Basis.Add(security.Basis)
		a.Portfolio.Value = a.Portfolio.Value.Add(security.Value)
	}
}

// goroutine: store account.Names in userCache
func cacheAccountNames(u *User, accounts []Account) {
	for i := 0; i < len(accounts); i++ {
		u.cacheAccountName(&accounts[i])
	}
}

// goroutine: this checks and applies ScheduledCashFlows which are ready
func updateAccounts(accounts []Account, session *Session) {
	log.Printf("[MODEL] UPDATE ACCOUNTS(%d)", len(accounts))
	for i := 0; i < len(accounts); i++ {
		a := &accounts[i]
		a.updateAccount(session, true)
	}
}

func (a *Account) postQueryInit() {
	balance, valid := a.User.lookupAccountBalance(a.ID)
	if valid {
		a.Balance = balance
	}
}

func List(session *Session, all bool) []Account {
	db := session.DB
	u := session.GetUser()
	entries := []Account{}
	hidden_clause := ""
	if u == nil {
		return entries
	}

	if !all {
		hidden_clause = "(hidden = 0 OR hidden IS NULL)"
	}

	// Find Accounts for CurrentUser()
	db.Preload("AccountType").
	   Order("account_type_id").Order("Name").
	   Where(hidden_clause).
	   Where(&Account{UserID: u.ID}).Find(&entries)
	log.Printf("[MODEL] LIST ACCOUNTS(%d)", len(entries))

	for i := 0; i < len(entries); i++ {
		a := &entries[i]
		a.setSession(session)
		a.postQueryInit()
	}

	go cacheAccountNames(u, entries)
	return entries
}

func ListAccounts(session *Session, all bool) []Account {
	globals := config.GlobalConfig()

	entries := List(session, all)
	if entries != nil {
		if globals.UpdateAccountsOnLogin {
			go updateAccounts(entries, session)
		}
	}
	return entries
}

func (account *Account) ListImports(session *Session, limit int) []Import {
	db := session.DB
	entries := []Import{}

	if !account.Verified {
		account = account.Get(session, false)
		if account == nil {
			//errors.New("Permission Denied")
			return entries
		}
	}

	db.Order("created_on desc").Limit(limit).
	   Where(&Import{AccountID: account.ID}).Find(&entries)
	for i := 0; i < len(entries); i++ {
		imp := &entries[i]
		imp.Account.cloneVerified(account)
		imp.CountImported(session)
	}
	return entries
}

func (account *Account) ListScheduled(session *Session, canRecordOnly bool) []CashFlow {
	ignoreHasScheduled := false
	entries := []CashFlow{}
	if !account.Verified {
		account.Get(session, false)
	}

	if !account.Verified || (!account.HasScheduled && !ignoreHasScheduled) {
		return entries
	}

	db := session.DB
	query := map[string]interface{}{"account_id": account.ID,
					"type": "RCashFlow", "split": false}
	if canRecordOnly {
		db.Order("date asc").Preload("RepeatInterval.RepeatIntervalType").
				     Where("date <= ?", time.Now()).
				     Where("repeats_left > 0 OR repeats_left IS NULL").
				     Where("cash_flow_id > 0").
				     Where("repeat_interval_id > 0").
				     Joins("RepeatInterval").Find(&entries, query)
	} else {
		db.Order("date asc").Preload("RepeatInterval.RepeatIntervalType").
				     Preload("Payee").
				     Where("repeat_interval_id > 0").
				     Joins("RepeatInterval").Find(&entries, query)
	}

	for i := 0; i < len(entries); i++ {
		repeat := &entries[i]
		if canRecordOnly {
			repeat.Account.cloneVerified(account)
			repeat.PreloadRepeat(db)
		} else {
			// for Preload access to Account.User.Cache
			repeat.Account.cloneVerified(account)
			// for #Show
			repeat.Preload(db)
		}
	}
	log.Printf("[MODEL] LIST SCHEDULED ACCOUNT(%d:%d) (%t)",
		   account.ID, len(entries), canRecordOnly)

	if len(entries) > 0 && !account.HasScheduled {
		account.addScheduled(db)
	}
	return entries
}

func (u *User) getAccount(name string) *Account {
	a := new(Account)
	a.Name = name
	a.UserID = u.ID
	if a.Name != "" {
		// need Where because these are not primary keys
		db := getDbManager()
		db.Where(&a).First(&a)
	}

	if a.ID == 0 || !a.HaveAccessPermission(u.Session) {
		return nil
	}
	a.postQueryInit()
	return a
}

func GetAccountByName(session *Session, name string) *Account {
	u := session.GetUser()
	if u == nil {
		return nil
	}

	return u.getAccount(name)
}

// This will find a Company that must match both Company.Name and
// Company.Symbol. If such Company doesn't exist, one is created.
func (a *Account) GetSecurity(session *Session, company *Company) (*Security, error) {
	c := company.Create(false)
	if c == nil {
		return nil, errors.New("Invalid Request")
	}

	security := new(Security)
	security.CompanyID = c.ID
	security.AccountID = a.ID
	if a.ID > 0 {
		db := session.DB

		// need Where because these are not primary keys
		db.Preload("Account").
		   Where(&security).First(&security)
	}

	if security.ID > 0 {
		log.Printf("[MODEL] ACCOUNT(%d) COMPANY(%d) GET SECURITY(%d)",
			   a.ID, c.ID, security.ID)

		// verify Account
		if !security.HaveAccessPermission(session) {
			return nil, errors.New("Permission Denied")
		}
		security.Company = *c
		security.postQueryInit()
	}

	// return Security if not found, so CompanyID can be reused
	return security, nil
}

// This version will find a Company that matches Company.Symbol regardless of
// Company.Name. If no such Company, one is created with Company.Name = "".
func (a *Account) GetSecurityBySymbol(session *Session, symbol string) (*Security, error) {
	security := new(Security)
	security.AccountID = a.ID
	if a.ID > 0 {
		db := session.DB

		// need Where because these are not primary keys
		db.Preload("Account").
		   Where("symbol = ?", symbol).
		   Where(&Security{AccountID: a.ID}).
		   Joins("Company").First(&security)
	}

	if security.ID > 0 {
		log.Printf("[MODEL] ACCOUNT(%d) SYMBOL(%s) GET SECURITY(%d)",
			   a.ID, symbol, security.ID)

		// verify Account
		if !security.HaveAccessPermission(session) {
			return nil, errors.New("Permission Denied")
		}
		security.postQueryInit()
	} else {
		// create Company if doesn't exist
		c := new(Company)
		c.Symbol = symbol
		c = c.Create(false)
		if c == nil {
			return nil, errors.New("Invalid Request")
		}
		security.CompanyID = c.ID
	}

	// return Security if not found, so CompanyID can be reused
	return security, nil
}

func (a *Account) securityGetByImportName(session *Session, name string) *Security {
	security := new(Security)
	db := session.DB

	importName := security.sanitizeSecurityName(name)
	if a.ID > 0 && importName != "" {
		// need Where because these are not primary keys
		db.Preload("Account").
		   Where("import_name = ? OR name = ?", importName, importName).
		   Where(&Security{AccountID: a.ID}).
		   Joins("Company").First(&security)
	}
	log.Printf("[MODEL] ACCOUNT(%d) IMPORT GET SECURITY for (%s:%d)",
		   a.ID, importName, security.ID)

	if security.ID == 0 {
		return nil
	}

	// verify Account
	if !security.HaveAccessPermission(session) {
		return nil
	}
	return security
}

func (a *Account) Init() *Account {
	a.Taxable = true
	// a.UserID unset (not needed for New)
	return a
}

// Average Balance for last 30 days prior to end date; uses/requires a.Balance.
// If Account is less than 30 days old, this will add Zeros for those days which
// gives the correct behavior for Interest calculations.
func (a *Account) averageDailyBalance(db *gorm.DB, endDate time.Time) decimal.Decimal {
	var daysLeft int32 = 30
	var days int32
	var validEntries uint
	total := decimal.Zero

	latestBalance := a.CashBalance
	lastBalance := a.CashBalance
	lastTime := endDate
	thirtyDaysAgo := lastTime.AddDate(0, 0, int(-daysLeft))

	// entries will be descending order
	entries := new(CashFlow).ListByDate(db, a, &thirtyDaysAgo)

	for i := 0; i < len(entries); i++ {
		if daysLeft <= 0 {
			break
		}

		cf := &entries[i]
		lastBalance = cf.Balance
		if !(cf.Date.After(endDate)) {
			days = durationDays(lastTime.Sub(cf.Date))
			if days > 0 {
				if days > daysLeft {
					days = daysLeft
				}
				total = total.Add(lastBalance.Mul(decimal.NewFromInt32(days)))
				daysLeft -= days
			}
			lastTime = cf.Date
			validEntries += 1
		}
		// if no more entries[], but still days_left, this is correct
		// Balance that was in account for remaining days left, which
		// is handled outside loop (below)
		lastBalance = lastBalance.Sub(cf.Amount)
	}

	if daysLeft > 0 {
		balanceBeforeFirstCashFlow := lastBalance.Mul(decimal.NewFromInt32(daysLeft))
		log.Printf("[MODEL] ACCOUNT(%d) BALANCE ACTIVE($%f) IDLE($%f DAYS:%d)",
			   a.ID, total.InexactFloat64(),
			   balanceBeforeFirstCashFlow.InexactFloat64(), daysLeft)
		total = total.Add(balanceBeforeFirstCashFlow)
	}

	balance := total.DivRound(decimal.NewFromInt32(30), 2)
	log.Printf("[MODEL] ACCOUNT(%d) BALANCE($%f) 30-DAY AVG($%f from %d/%d entries)",
		   a.ID, latestBalance.InexactFloat64(),
		   balance.InexactFloat64(), validEntries, len(entries))
	return balance
}

func (a Account) HasAverageDailyBalance() bool {
	return !a.AverageBalance.IsZero()
}

func (a *Account) SetAverageDailyBalance(session *Session) {
	a.AverageBalance = a.averageDailyBalance(session.DB, time.Now())
}

func (a *Account) addScheduled(db *gorm.DB) {
	db.Omit(clause.Associations).Model(a).Update("HasScheduled", 1)
}

func (a *Account) lastCashFlow(imported bool) *CashFlow {
	var entry *CashFlow
	db := getDbManager()
	query := map[string]interface{}{"account_id": a.ID, "type": nil,
					"repeat_interval_id": 0,
					"transfer": false}
	if (imported) {
		db.Order("date desc").Where("import_id > 0").
		   First(&entry, query)
	} else {
		db.Order("date desc").First(&entry, query)
	}
	return entry
}

func (a *Account) updateBalance(c *CashFlow) {
	db := getDbManager()

	// catastrophic if we end up here without a.Verified
	assert(a.Verified, "Unexpected: Account.Verified Unset!")
	assert(a.ID > 0, "Unexpected: Account.ID Unset!")
	if !c.mustUpdateBalance() {
		return
	}

	adjustAmount := c.Amount.Sub(c.oldAmount)
	if adjustAmount.IsZero() {
		return
	}

	oldCashBalance := a.CashBalance
	// Update object Balance fields just in case used in caller;
	// If we didn't have accurate Balance, these will be unused in caller.
	a.Balance = a.Balance.Add(adjustAmount)
	a.CashBalance = a.CashBalance.Add(adjustAmount)
	a.User.updateAccountBalance(a, adjustAmount)
	// TODO: should be fine to discard cached Balance as written below

	if c.oldAmount.IsZero() || oldCashBalance.IsZero() {
		// This case intended to handle when we don't know if we have
		// accurate Account Balances, and so just use +delta.
		// (Such as updates for Transfer/Pair).
		// But will fall into this case when Balance/oldAmount is 0.
		log.Printf("[MODEL] UPDATE CASH BALANCE ACCOUNT(%d:%d): +%f",
			   a.ID, c.ID, adjustAmount.InexactFloat64())
		db.Omit(clause.Associations).Model(a).
		   Updates(map[string]interface{}{
			   "cash_balance": gorm.Expr("cash_balance + ?", adjustAmount),
			   "balance": gorm.Expr("balance + ?", adjustAmount)})
	} else {
		// Could delete this and always use above. Need performance data.
		log.Printf("[MODEL] UPDATE CASH BALANCE ACCOUNT(%d:%d): %f -> %f",
			   a.ID, c.ID, oldCashBalance.InexactFloat64(),
			   a.CashBalance.InexactFloat64())
		db.Omit(clause.Associations).Model(a).
		   Updates(Account{CashBalance: a.CashBalance,
				   Balance: a.Balance})
	}
}

func (a *Account) writeBalance() {
	if !a.Verified {
		return
	}
	db := getDbManager()

	db.Omit(clause.Associations).Model(a).
	   Update("Balance", a.Balance)
	log.Printf("[MODEL] ACCOUNT(%d) WRITE CACHED BALANCE(%f)",
		   a.ID, a.Balance.InexactFloat64())
}

func (a *Account) Create(session *Session) error {
	db := session.DB
	u := session.GetUser()
	if u == nil {
		return errors.New("Permission Denied")
	}

	a.sanitizeInputs()
	// Account.User is set to CurrentUser()
	a.UserID = u.ID
	spewModel(a)
	result := db.Omit(clause.Associations).Create(a)
	return result.Error
}

func (a *Account) cloneVerifiedFrom(src *Account) {
	a.User.ID = src.User.ID
	a.User.Session = src.User.Session
	a.User.UserSettings = src.User.UserSettings
	a.Session = src.Session
	a.Verified = src.Verified
	assert(a.Session == a.User.Session, "Account/User Corrupt!")
}

func (a *Account) cloneVerified(src *Account) {
	a.cloneVerifiedFrom(src)
	a.ID = src.ID
	a.AccountTypeID = src.AccountTypeID
	a.Balance = src.Balance
	a.CashBalance = src.CashBalance
}

// Account.User is populated from Session
func (a *Account) setSession(session *Session) bool {
	u := session.GetUser()
	a.Verified = (u != nil)
	if a.Verified {
		a.UserID = u.ID
		a.User = *u
		a.Session = session
	}
	return a.Verified
}

func (a *Account) HaveAccessPermission(session *Session) bool {
	u := session.GetUser()
	// store in a.Verified if this Account is trusted
	a.Verified = !(u == nil || a.ID == 0 || u.ID != a.UserID)
	if a.Verified {
		a.User = *u
		a.Session = session
	}
	return a.Verified
}

func (a *Account) updateAccountScheduled(session *Session) {
	var repeat *CashFlow
	enableScheduledCashFlow := true

	// test if any ScheduledCashFlows need to post
	scheduled := a.ListScheduled(session, true)
	if !a.Verified || len(scheduled) == 0 {
		return
	}

	log.Printf("[MODEL] UPDATE ACCOUNT(%d) HAVE SCHEDULED(%d)",
		   a.ID, len(scheduled))
	if !enableScheduledCashFlow {
		return
	}

	for i := 0; i < len(scheduled); i++ {
		repeat = &scheduled[i]
		total, err := repeat.tryInsertRepeatCashFlow()
		if err == nil {
			// update inside loop so cloneVerified carries Balance
			// changes forward to next repeat
			a.Balance = a.Balance.Add(total)
			a.CashBalance = a.CashBalance.Add(total)
		}
	}
}

func (a *Account) updateAccount(session *Session, async bool) {
	if !a.HaveAccessPermission(session) {
		return
	}

	a.postQueryInit()
	a.updateAccountScheduled(session)

	// invoked only when run from a goroutine, should
	// avoid extra database access from Account.Get
	if async {
		// updates any Security.Values and Account.Balance, we do
		// !async because if aready async then we don't want nested
		// goroutines to complete and not also run in background
		a.getOpenSecurities(!async)
	}
}

// update Account.Balance from Securities.Value
func (a *Account) updateValue(debugValue bool) {
	if !a.Verified || !a.IsInvestment() || len(a.Securities) == 0 {
		return
	}

	oldBalance := a.Balance
	a.Balance = a.CashBalance
	for i := 0; i < len(a.Securities); i++ {
		s :=  &a.Securities[i]
		a.Balance = a.Balance.Add(s.Value)
	}
	a.User.cacheAccountBalance(a)

	if debugValue {
		log.Printf("[MODEL] ACCOUNT(%d:%d) REFRESH BALANCE(%f -> %f)",
			   a.ID, len(a.Securities),
			   oldBalance.InexactFloat64(),
			   a.Balance.InexactFloat64())
	}
}

// Show, Edit, Delete, Update use Get
// a.UserID unset, need to load
func (a *Account) Get(session *Session, preload bool) *Account {
	db := session.DB

	// Load and Verify we have access to Account
	if a.ID > 0 {
		if preload {
			// Get (Show)
			db.Preload("AccountType").First(&a)
		} else {
			// Edit, Delete, Update
			db.First(&a)
		}
	}
	if !a.HaveAccessPermission(session) {
		return nil
	}

	if preload {
		a.updateAccount(session, false)
		spewModel(a)
	}
	return a
}

func (a *Account) Delete(session *Session) error {
	// Verify we have access to Account
	a = a.Get(session, false)
	if a == nil {
		return errors.New("Permission Denied")
	}
	db := session.DB

	spewModel(a)
	count := new(CashFlow).Count(a) +
		 new(Trade).Count(a)
	log.Printf("[MODEL] DELETE ACCOUNT(%d) IF COUNT(%d == 0)", a.ID, count)
	if count == 0 {
		db.Delete(a)
	} else if !a.Hidden {
		// if cannot delete, then make Hidden
		a.Hidden = true
		db.Omit(clause.Associations).Save(a)
	}

	return nil
}

// Account access already verified with Get
func (a *Account) Update() error {
	db := getDbManager()
	a.sanitizeInputs()
	spewModel(a)
	result := db.Omit(clause.Associations).Save(a)
	a.User.clearAccountName(a)
	return result.Error
}


// Find() for use with rails/ruby like REPL console (gomacro);
// controllers should not expose this as are no access controls
func (*Account) Find(ID uint) *Account {
	db := getDbManager()
	a := new(Account)
	db.First(&a, ID)
	a.Verified = true
	return a
}

func (a *Account) AddToBalance(fAmount float64) {
	amount := decimal.NewFromFloat(fAmount)
	a.Balance = a.Balance.Add(amount)
	a.CashBalance = a.CashBalance.Add(amount)
}

func (a *Account) Print() {
	forceSpewModel(a.Model, 0)
	forceSpewModel(a, 1)
}

func (a *Account) PrintAll() {
	forceSpewModel(a, 0)
}
