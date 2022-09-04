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

type User struct {
	gorm.Model
	Login string
	Email string
	Password string `gorm:"->:false;<-"`
	Categories []Category
	Payees []Payee
}

var currentUser *User

func init() {
	// replace when adding User login
	currentUser = new(User)
	currentUser.ID = 1
	log.Printf("SET CURRENT USER(%d)", currentUser.ID)
}

func GetCurrentUser() *User {
	// replace with Sessions
	return currentUser
}
