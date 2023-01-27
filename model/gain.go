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
	AdjustedShares decimal.Decimal // deprecated
	Basis decimal.Decimal
}

func (*TradeGain) List(db *gorm.DB) []TradeGain {
	entries := []TradeGain{}
	db.Find(&entries)

	return entries
}

func (t *Trade) ListGains(db *gorm.DB) []TradeGain {
	entries := []TradeGain{}
	db.Where(&TradeGain{BuyID: t.ID}).Find(&entries)
	return entries
}

func (tg *TradeGain) recordGain(db *gorm.DB, sell *Trade, buy *Trade,
				maxShares decimal.Decimal,
				updateDB bool) {
	buyRemain := buy.SharesRemaining()
	tg.SellID = sell.ID
	tg.BuyID = buy.ID
	tg.DaysHeld = durationDays(sell.Date.Sub(buy.Date))
	tg.Shares = decimal.Min(maxShares, buyRemain)
	tg.Basis = buy.Amount.Sub(buy.Basis)
	if !buyRemain.Equal(tg.Shares) {
		// must calculate using Basis per share
		tg.Basis = tg.Basis.Div(buyRemain).Mul(tg.Shares).Round(2)
	}
	// [sell,buy].Basis is updated in caller

	if updateDB {
		db.Omit(clause.Associations).Create(tg)
	}
}
