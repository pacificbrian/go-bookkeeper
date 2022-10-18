/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"gorm.io/gorm"
)

const (
	UndefinedType uint = iota
	Debit
	Credit
	DebitTransfer
	CreditTransfer
)

type CashFlowType struct {
	Model
	Name string `form:"cash_flow_type.Name"`
}

func CashFlowTypeIsCredit(CashFlowType uint) bool {
	return (CashFlowType == Credit || CashFlowType == CreditTransfer)
}

func CashFlowTypeIsDebit(CashFlowType uint) bool {
	return (CashFlowType == Debit || CashFlowType == DebitTransfer)
}

func (*CashFlowType) List(db *gorm.DB) []CashFlowType {
	// need userCache lookup
	entries := []CashFlowType{}
	db.Find(&entries)

	return entries
}
