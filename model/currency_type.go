/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"gorm.io/gorm"
)

type CurrencyType struct {
	Model
	Name string `form:"currency_type.Name"`
}

func (*CurrencyType) List(db *gorm.DB) []CurrencyType {
	entries := []CurrencyType{}
	db.Find(&entries)

	return entries
}
