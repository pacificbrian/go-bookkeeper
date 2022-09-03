/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"gorm.io/gorm"
)

type Payee struct {
	Model
	UserID uint `gorm:"not null"`
	User User
	CategoryID uint `form:"payee.category_id"`
	Category Category
	Name string `form:"payee.Name"`
	Address string
	SkipOnImport bool `form:"payee.SkipOnImport"`
}

func (*Payee) List(db *gorm.DB) []Payee {
	entries := []Payee{}
	db.Find(&entries)
	return entries
}
