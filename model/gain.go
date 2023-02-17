/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"log"
	"time"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TradeGain struct {
	Model
	SellID uint
	BuyID uint
	DaysHeld int32
	Shares decimal.Decimal
	AdjustedShares decimal.Decimal // deprecated
	Basis decimal.Decimal
	BasisPS decimal.Decimal `gorm:"-:all"`
	Amount decimal.Decimal `gorm:"-:all"`
	Gain decimal.Decimal `gorm:"-:all"`
	GainPS decimal.Decimal `gorm:"-:all"`
	BuyDate time.Time `gorm:"-:all"`
	Sell Trade
	Buy Trade
}

func (t *Trade) ListGains(db *gorm.DB) []TradeGain {
	entries := []TradeGain{}

	if !t.Account.Verified || t.ID == 0 {
		return entries
	}

	if t.IsBuy() {
		db.Where(&TradeGain{BuyID: t.ID}).Find(&entries)
	} else if t.IsSell() {
		db.Where(&TradeGain{SellID: t.ID}).Find(&entries)
		for i := 0; i < len(entries); i++ {
			tg := &entries[i]
			tg.Amount = t.Amount.Div(t.Shares).Mul(tg.Shares).Round(2)
			tg.Gain = tg.Amount.Sub(tg.Basis)
			tg.GainPS = tg.Gain.Div(tg.Shares)
			tg.BasisPS = tg.Basis.Div(tg.Shares)
			if tg.DaysHeld > 0 {
				tg.BuyDate = t.Date.AddDate(0,0,int(-tg.DaysHeld))
			} else {
				buy := new(Trade)
				db.Select("date").First(&buy, tg.BuyID)
				tg.BuyDate = buy.Date
			}
		}
	}
	log.Printf("[MODEL] LIST ACCOUNT(%d) GAINS(%d:%d)",
		   t.AccountID, t.TradeTypeID, len(entries))

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


// Debug routines -

// Find() for use with rails/ruby like REPL console (gomacro);
// controllers should not expose this as are no access controls
func (*TradeGain) Find(ID uint) *TradeGain {
	db := getDbManager()
	tg := new(TradeGain)
	db.First(&tg, ID)
	return tg
}

func (tg *TradeGain) Print() {
	forceSpewModel(tg.Model, 0)
	forceSpewModel(tg, 1)
}

func (tg *TradeGain) PrintAll() {
	forceSpewModel(tg, 0)
}
