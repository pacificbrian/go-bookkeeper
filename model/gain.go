/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"errors"
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
	BasisFIFO decimal.Decimal
	BasisPS decimal.Decimal `gorm:"-:all"`
	Amount decimal.Decimal `gorm:"-:all"`
	Gain decimal.Decimal `gorm:"-:all"`
	GainPS decimal.Decimal `gorm:"-:all"`
	BuyDate time.Time `gorm:"-:all"`
}

func (t *Trade) ListGains(db *gorm.DB) ([]TradeGain, []decimal.Decimal) {
	entries := []TradeGain{}
	var totals []decimal.Decimal

	if !t.Account.Verified || t.ID == 0 {
		return entries, totals
	}

	if t.IsBuy() {
		db.Where(&TradeGain{BuyID: t.ID}).Find(&entries)
	} else if t.IsSell() {
		db.Where(&TradeGain{SellID: t.ID}).Find(&entries)
		totals = make([]decimal.Decimal, 3)
		for i := 0; i < len(entries); i++ {
			tg := &entries[i]
			tg.postQueryInit(t)
			if tg.DaysHeld > 0 {
				tg.BuyDate = t.Date.AddDate(0,0,int(-tg.DaysHeld))
			} else {
				buy := new(Trade)
				db.Select("date").First(&buy, tg.BuyID)
				tg.BuyDate = buy.Date
			}
			totals[0] = totals[0].Add(tg.Amount)
			totals[1] = totals[1].Add(tg.Basis)
			totals[2] = totals[2].Add(tg.Gain)
		}
	}
	log.Printf("[MODEL] LIST ACCOUNT(%d) GAINS(%d:%d)",
		   t.AccountID, t.TradeTypeID, len(entries))

	return entries, totals
}

func (tg *TradeGain) postQueryInit(sold *Trade) {
	tg.Amount = sold.Amount.Div(sold.Shares).Mul(tg.Shares).Round(2)
	tg.Gain = tg.Amount.Sub(tg.Basis)
	tg.GainPS = tg.Gain.Div(tg.Shares)
	tg.BasisPS = tg.Basis.Div(tg.Shares)
}

func (tg *TradeGain) recordGain(sell *Trade, buy *Trade,
				maxShares decimal.Decimal,
				updateDB bool) {
	buyRemain := buy.SharesRemaining()
	tg.SellID = sell.ID
	tg.BuyID = buy.ID
	tg.DaysHeld = durationDays(sell.Date.Sub(buy.Date))
	tg.Shares = decimal.Min(maxShares, buyRemain)
	tg.Basis = buy.gainBasis(tg.Shares)
	tg.BasisFIFO = buy.gainBasisFIFO(tg.Shares)
	tg.postQueryInit(sell)
	// [sell,buy].Basis is updated in caller

	if updateDB {
		db := getDbManager()
		db.Omit(clause.Associations).Create(tg)
	}
}

func (tg *TradeGain) Delete(session *Session) error {
	db := getDbManager()

	buy := &Trade{}
	buy.ID = tg.BuyID
	buy = buy.Get(session)
	if buy == nil {
		return errors.New("Permission Denied")
	}
	if tg.BasisFIFO.IsZero() {
		tg.BasisFIFO = tg.Basis
	}
	buy.revertBasis(tg.BasisFIFO, tg.Shares)

	db.Delete(tg)
	log.Printf("[MODEL] DELETE GAIN(%d) FOR BUY(%d)", tg.ID, buy.ID)
	return nil
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

func (*TradeGain) FindForSale(ID uint) []TradeGain {
	db := getDbManager()
	entries := []TradeGain{}
	db.Where(&TradeGain{SellID: ID}).Find(&entries)
	return entries
}

func (tg *TradeGain) Save() error {
	db := getDbManager()
	result := db.Omit(clause.Associations).Save(tg)
	return result.Error
}

func (tg *TradeGain) Print() {
	forceSpewModel(tg.Model, 0)
	forceSpewModel(tg, 1)
}

func (tg *TradeGain) PrintAll() {
	forceSpewModel(tg, 0)
}
