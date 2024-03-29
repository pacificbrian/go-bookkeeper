/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"errors"
	"log"
	"sync"
	"github.com/pacificbrian/go-bookkeeper/config"
	gormdb "github.com/pacificbrian/go-bookkeeper/db"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm/clause"
	"gorm.io/gorm"
)

const (
	DefaultCashFlowLimit = 200
)

type UserCache struct {
	AccountBalances map[uint]decimal.Decimal
	AccountNames map[uint]string
	CategoryNames map[uint]string
	mutex sync.Mutex
}

type UserSettings struct {
	Model
	UserID uint
	CashFlowLimit int `gorm:"-:all"`
}

type User struct {
	gorm.Model
	Login string `form:"user.Login"`
	Email string `form:"user.Email"`
	PasswordDigest string
	CashflowLimit int
	Session *Session `gorm:"-:all"`
	UserSettings UserSettings
	Accounts []Account
	Payees []Payee
}

type Session struct {
	User User
	Cache UserCache
	DB *gorm.DB
	DebugDB *gorm.DB
}

func (u *User) sanitizeInputs() {
	sanitizeString(&u.Login)
	sanitizeString(&u.Email)
}

func (u *User) Cache() *UserCache {
	if u.Session == nil {
		return nil
	}
	return &u.Session.Cache
}

func (u *User) cacheAccountBalance(a *Account) {
	uc := u.Cache()
	uc.mutex.Lock()
	uc.AccountBalances[a.ID] = a.Balance
	uc.mutex.Unlock()
}

func (u *User) writeAccountBalance(a *Account, update decimal.Decimal, force bool) decimal.Decimal {
	uc := u.Cache()

	uc.mutex.Lock()
	balance, valid := u.lookupAccountBalance(a.ID)
	if !valid && force {
		balance = a.Balance
	}
	if valid || force {
		balance = balance.Add(update)
		uc.AccountBalances[a.ID] = balance
	}
	uc.mutex.Unlock()
	return balance
}

func (u *User) insertAccountBalance(a *Account, update decimal.Decimal) decimal.Decimal {
	return u.writeAccountBalance(a, update, true)
}

func (u *User) updateAccountBalance(a *Account, update decimal.Decimal) decimal.Decimal {
	return u.writeAccountBalance(a, update, false)
}

func (u *User) cacheAccountName(a *Account) {
	u.Cache().AccountNames[a.ID] = a.Name
}

func (u *User) clearAccountName(a *Account) {
	uc := u.Cache()
	if uc != nil {
		delete(uc.AccountNames, a.ID)
	}
}

func (u *User) cacheCategoryName(c *Category) {
	u.Cache().CategoryNames[c.ID] = c.Name
	log.Printf("[CACHE] ADD CATEGORY(%d: %s)", c.ID, c.Name)
}

func (u *User) lookupAccountBalance(id uint) (decimal.Decimal, bool) {
	v, valid := u.Cache().AccountBalances[id]
	return v, valid
}

func (u *User) lookupAccountName(id uint) string {
	return u.Cache().AccountNames[id]
}

func (u *User) lookupCategoryName(id uint) string {
	return u.Cache().CategoryNames[id]
}

func getDbManager() *gorm.DB {
	return gormdb.DbManager()
}

func (uc *UserCache) init() {
	uc.AccountBalances = map[uint]decimal.Decimal{}
	uc.AccountNames = map[uint]string{}
	uc.CategoryNames = map[uint]string{}
	uc.mutex = sync.Mutex{}
}

func (u *User) initSettings() {
	globals := config.GlobalConfig()

	cashflowLimit := DefaultCashFlowLimit
	if u.CashflowLimit > 0 {
		cashflowLimit = u.CashflowLimit
	} else if globals.CashFlowLimit > 0 {
		cashflowLimit = globals.CashFlowLimit
	}
	u.UserSettings.CashFlowLimit = cashflowLimit
	u.UserSettings.UserID = u.ID
}

func (u *User) init(session *Session) {
	u.initSettings()
	u.Session = session
}

func (sn *Session) init() {
	sn.Cache.init()
	sn.DebugDB = gormdb.DebugDbManager()
	sn.DB = gormdb.DbManager()
}

func (session *Session) GetUser() *User {
	return &session.User
}

func (session *Session) CloseSession() {
	u := session.GetUser()
	log.Printf("[MODEL] CLOSE SESSION FOR USER(%d)", u.ID)

	// empty Session Caches
	a := new(Account)
	a.setSession(session)
	for k,v := range session.Cache.AccountBalances {
		a.ID = k
		a.Balance = v
		a.writeBalance()
	}
}

func (u *User) NewSession() *Session {
	newSession := new(Session)
	newSession.init()
	newSession.User.ID = u.ID
	newSession.User.init(newSession)

	log.Printf("[MODEL] NEW SESSION USER(%d) AUTH(%t)", u.ID, true)
	return newSession
}

func (u *User) setPassword(password string) error {
	hPassword, err := bcrypt.GenerateFromPassword([]byte(password),
						      bcrypt.DefaultCost)
	if err == nil {
		u.PasswordDigest = string(hPassword)
	}
	return err
}

func (u *User) Authenticate(password string) bool {
	// special case to allow empty password if user wants no security
	if u.PasswordDigest == "" && password == "" {
		return true
	}

	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordDigest),
					     []byte(password))
	return err == nil
}

func (u *User) useCount() uint {
	return u.countAccounts() + u.countPayees()
}

func (u *User) countAccounts() uint {
	var count int64 = 0

	db := getDbManager()
	db.Model(&Account{}).
	   Where("user_id = ?", u.ID).Count(&count)
	log.Printf("[MODEL] COUNT ACCOUNTS USER(%d:%d)", u.ID, count)

	return uint(count)
}

func (u *User) countPayees() uint {
	var count int64 = 0

	db := getDbManager()
	db.Model(&Payee{}).
	   Where("user_id = ?", u.ID).Count(&count)
	log.Printf("[MODEL] COUNT PAYEES USER(%d:%d)", u.ID, count)

	return uint(count)
}

func (u *User) GetByLogin(login string) *User {
	db := getDbManager()

	u.Login = login
	db.Where(&u).First(&u)
	if u.ID == 0 {
		return nil
	}

	log.Printf("[MODEL] GET USER(%d) BY LOGIN(%s)", u.ID, login)
	return u
}

func (u *User) Create(password [2]string) error {
	u.sanitizeInputs()
	avail := new(User).GetByLogin(u.Login) == nil
	if !avail {
		return errors.New("Duplicate User Login")
	}
	if password[0] != password[1] {
		return errors.New("Passwords don't match")
	}
	db := getDbManager()

	u.setPassword(password[0])
	spewModel(u)
	result := db.Omit(clause.Associations).Create(u)
	return result.Error
}

func (u *User) HaveAccessPermission(session *Session) *User {
	us := session.GetUser()
	if !(us == nil || us.ID != u.ID) {
		return us
	} else {
		return nil
	}
}

// Edit, Delete, Update use Get
func (u *User) Get(session *Session) *User {
	// Verify we have access to User
	return u.HaveAccessPermission(session)
}

func (u *User) Delete(session *Session) error {
	// Verify we have access to Payee
	u = u.Get(session)
	if u == nil {
		return errors.New("Permission Denied")
	}
	db := session.DB

	spewModel(u)
	count := u.useCount()
	log.Printf("[MODEL] DELETE USER(%d) IF COUNT(%d == 0)", u.ID, count)
	if count == 0 {
		db.Delete(u)
	}
	return nil
}
