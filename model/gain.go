/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TradeGain struct {
	Model
	SellID uint
	BuyID uint
	Sell Trade
	Buy Trade
	DaysHeld int32
	Shares decimal.Decimal
	AdjustedShares decimal.Decimal
	Basis decimal.Decimal
}

func (*TradeGain) List(db *gorm.DB) []TradeGain {
	entries := []TradeGain{}
	db.Find(&entries)

	return entries
}

func (tg *TradeGain) recordGain(db *gorm.DB, sell *Trade, buy *Trade,
				shares decimal.Decimal) {
	tg.SellID = sell.ID
	tg.BuyID = buy.ID
	tg.DaysHeld = durationDays(sell.Date.Sub(buy.Date))
	tg.Shares = sell.Shares
	tg.Basis = buy.Amount.Sub(buy.Basis)
	// [sell,buy].Basis is updated in caller

	db.Omit(clause.Associations).Create(tg)
}
