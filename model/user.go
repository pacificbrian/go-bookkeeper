/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"log"
	"github.com/pacificbrian/go-bookkeeper/config"
	gormdb "github.com/pacificbrian/go-bookkeeper/db"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	DefaultCashFlowLimit = 200
)

type UserCache struct {
	AccountBalances map[uint]decimal.Decimal
	AccountNames map[uint]string
	CategoryNames map[uint]string
}

type UserSettings struct {
	Model
	UserID uint
	CashFlowLimit int `gorm:"-:all"`
}

type User struct {
	gorm.Model
	Login string
	Email string
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

func (u *User) Cache() *UserCache {
	if u.Session == nil {
		return nil
	}
	return &u.Session.Cache
}

func (u *User) cacheAccountBalance(a *Account) {
	u.Cache().AccountBalances[a.ID] = a.Balance
}

func (u *User) updateAccountBalance(a *Account) {
	if !u.Cache().AccountBalances[a.ID].IsZero() {
		u.Cache().AccountBalances[a.ID] = a.Balance
	}
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

func (u *User) lookupAccountBalance(id uint) decimal.Decimal {
	return u.Cache().AccountBalances[id]
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
