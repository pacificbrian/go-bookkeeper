/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"log"
	"golang.org/x/crypto/bcrypt"
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
	//Password string `gorm:"->:false;<-"`
	CashflowLimit int
	Cache *UserCache `gorm:"-:all"`
	UserSettings UserSettings
	Categories []Category
	Payees []Payee
}

type Session struct {
	User User
	Cache UserCache
}

var currentSession *Session

func (u *User) cacheAccount(a *Account) {
	u.Cache.AccountNames[a.ID] = a.Name
}

func (u *User) cacheCategory(c *Category) {
	u.Cache.CategoryNames[c.ID] = c.Name
}

func (u *User) lookupAccount(id uint) string {
	return u.Cache.AccountNames[id]
}

func (u *User) lookupCategory(id uint) string {
	return u.Cache.CategoryNames[id]
}

func (uc *UserCache) init() {
	uc.AccountNames = map[uint]string{}
	uc.CategoryNames = map[uint]string{}
}

func (u *User) initSettings() {
	u.CashflowLimit = DefaultCashFlowLimit
	u.UserSettings.CashFlowLimit = u.CashflowLimit
	u.UserSettings.UserID = u.ID
}

func (u *User) init(userCache *UserCache) {
	u.ID = 1
	u.initSettings()
	u.Cache = userCache
}

func init() {
	// replace when adding User login
	currentSession = new(Session)
	currentSession.Cache.init()
	currentSession.User.init(&currentSession.Cache)

	// example of using bcrypt for passwords
	// overwrite u.PasswordDigest just for testing
	password := "Gopher"
	GetCurrentUser().setPassword(password)
	validPassword := GetCurrentUser().Authenticate(password)

	log.Printf("SET CURRENT USER(%d) AUTH(%t)", GetCurrentUser().ID,
		   validPassword)
}

func GetCurrentUser() *User {
	return &currentSession.User
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
