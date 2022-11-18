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
	AccountTypeUndefined uint = iota
	AccountTypeCash
	AccountTypeDeposit
	AccountTypeCreditCard
	AccountTypeInvestment
	AccountTypeHealthCare
	AccountTypeLoan
	AccountTypeAsset
	AccountTypeCrypto
)

type AccountType struct {
	Model
	Name string `form:"account_type.Name"`
}

type CurrencyType struct {
	Model
	Name string `form:"currency_type.Name"`
}

func (a AccountType) GetAltText() string {
	return a.Name
}

func (a AccountType) GetIconPath() string {
	var path string
	switch a.ID {
	case AccountTypeCash:
		path = "images/icons/icn_small_deposit.png"
	case AccountTypeDeposit:
		path = "images/icons/icn_small_deposit.png"
	case AccountTypeCreditCard:
		path = "images/icons/icn_small_credit_card.gif"
	case AccountTypeInvestment:
		path = "images/icons/icn_investments.png"
	case AccountTypeLoan:
		path = "images/icons/icn_home.png"
	case AccountTypeHealthCare:
		path = "images/icons/icn_health.png"
	case AccountTypeAsset:
		path = "images/icons/icn_home.png"
	case AccountTypeCrypto:
		path = "images/icons/bitcoin.png"
	default:
		path = "images/icons/wrench.svg"
	}

	return path
}

func (a *AccountType) isInvestment() bool {
	return a.ID == AccountTypeInvestment
}

func (a *AccountType) isType(askedType string) bool {
	return a.Name == askedType
}

func (*AccountType) List(db *gorm.DB) []AccountType {
	entries := []AccountType{}
	db.Find(&entries)

	return entries
}

func (*CurrencyType) List(db *gorm.DB) []CurrencyType {
	entries := []CurrencyType{}
	db.Find(&entries)

	return entries
}
