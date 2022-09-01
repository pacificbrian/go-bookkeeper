/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"gorm.io/gorm"
)

type CashFlowType struct {
	gorm.Model
	Name string `form:"cash_flow_type.Name"`
}

func (*CashFlowType) List(db *gorm.DB) []CashFlowType {
	entries := []CashFlowType{}
	db.Find(&entries)

	return entries
}
