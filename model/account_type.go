/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"gorm.io/gorm"
)

type AccountType struct {
	gorm.Model
	Name string `form:"account_type.Name"`
}

func (*AccountType) List(db *gorm.DB) []AccountType {
	entries := []AccountType{}
	db.Find(&entries)

	return entries
}
