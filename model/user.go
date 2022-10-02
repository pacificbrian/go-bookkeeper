/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"log"
	"gorm.io/gorm"
)

const (
	DefaultCashFlowLimit = 200
)

type UserSettings struct {
	Model
	UserID uint
	CashFlowLimit int `gorm:"-:all"`
}

type User struct {
	gorm.Model
	Login string
	Email string
	//Password string `gorm:"->:false;<-"`
	CashflowLimit int
	Categories []Category
	Payees []Payee
	UserSettings UserSettings
}

var currentUser *User

func (u *User) initSettings() {
	u.CashflowLimit = DefaultCashFlowLimit
	u.UserSettings.CashFlowLimit = u.CashflowLimit
	u.UserSettings.UserID = u.ID
}

func init() {
	// replace when adding User login
	currentUser = new(User)
	currentUser.ID = 1
	currentUser.initSettings()
	log.Printf("SET CURRENT USER(%d)", currentUser.ID)
}

func GetCurrentUser() *User {
	// replace with Sessions
	return currentUser
}
