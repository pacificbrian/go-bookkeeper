/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"log"
	"golang.org/x/crypto/bcrypt"
	"github.com/pacificbrian/go-bookkeeper/config"
	gormdb "github.com/pacificbrian/go-bookkeeper/db"
	"gorm.io/gorm"
)

const (
	DefaultCashFlowLimit = 200
)

type UserCache struct {
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
	return &u.Session.Cache
}

func (u *User) cacheAccount(a *Account) {
	u.Cache().AccountNames[a.ID] = a.Name
}

func (u *User) cacheCategory(c *Category) {
	u.Cache().CategoryNames[c.ID] = c.Name
	log.Printf("[CACHE] ADD CATEGORY(%d: %s)", c.ID, c.Name)
}

func (u *User) lookupAccount(id uint) string {
	return u.Cache().AccountNames[id]
}

func (u *User) lookupCategory(id uint) string {
	return u.Cache().CategoryNames[id]
}

func (uc *UserCache) init() {
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

func (session *Session) GetCurrentUser() *User {
	return &session.User
}

func (u *User) NewSession() *Session {
	newSession := new(Session)
	newSession.init()
	newSession.User.ID = u.ID
	newSession.User.init(newSession)

	// example of using bcrypt for passwords
	// overwrite u.PasswordDigest just for testing
	password := "Gopher"
	u.setPassword(password)
	validPassword := u.Authenticate(password)

	log.Printf("[MODEL] NEW SESSION USER(%d) AUTH(%t)", u.ID,
		   validPassword)
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
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordDigest),
					     []byte(password))
	return err == nil
}
