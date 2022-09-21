/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"gorm.io/gorm"
)

type TradeType struct {
	Model
	Name string `form:"trade_type.Name"`
}

func (*TradeType) List(db *gorm.DB) []TradeType {
	// need userCache lookup
	entries := []TradeType{}
	db.Find(&entries)

	return entries
}
