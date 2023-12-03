/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"github.com/pacificbrian/qif"
	"gorm.io/gorm"
)

const (
	UndefinedTradeType uint = iota
	Buy
	Sell
	// below are not Trades, but CashFlow Credits
	Dividend
	Distribution
	// below are effectively Buy types, but no CashFlow Debit
	ReinvestedDividend
	ReinvestedDistribution
	// below only for moving Shares between accounts, careful some 401k
	// exported data will incorrectly encode ReinvestedDividend as SharesIn
	SharesIn
	SharesOut
	// trade.Shares is split ratio (specified negative for reverse split)
	Split
)

type TradeType struct {
	Model
	Name string `form:"trade_type.Name"`
}

// SQL query string for Buy types
var listBuyTypes = "id = 1 OR id = 5 OR id = 6"
// SQL query string for all Buy, Sell types for Trades
var TradeTypeQueries = [9]string{"",
				 "trade_type_id = 1 OR trade_type_id = 5 OR trade_type_id = 6",
				 "trade_type_id = 2",
				 "trade_type_id = 3 OR trade_type_id = 5",
				 "trade_type_id = 4 OR trade_type_id = 6",
				 "", // use Buy or Dividend
				 "", // use Buy or Distribution
				 "trade_type_id = 7",
				 "trade_type_id = 8"}
var TradeTypeCashFlowsQuery string = "trade_type_id <= 6"

var TradeTypeQueryDesc = [9]string{"",
				   "",
				   "Shares Sold",
				   "Dividend",
				   "Distribution",
				   "", // use Buy or Dividend
				   "", // use Buy or Distribution
				   "",
				   ""}

func TradeTypeIsValid(TradeTypeID uint) bool {
	return TradeTypeID > 0 && TradeTypeID <= Split
}

func TradeTypeIsBuy(TradeTypeID uint) bool {
	return (TradeTypeID == Buy)
}

func TradeTypeIsDividend(TradeTypeID uint) bool {
	return (TradeTypeID == ReinvestedDividend ||
		TradeTypeID == Dividend)
}

func TradeTypeIsDistribution(TradeTypeID uint) bool {
	return (TradeTypeID == ReinvestedDistribution ||
		TradeTypeID == Distribution)
}

func TradeTypeIsReinvest(TradeTypeID uint) bool {
	return (TradeTypeID == ReinvestedDividend ||
		TradeTypeID == ReinvestedDistribution)
}

func TradeTypeIsSell(TradeTypeID uint) bool {
	return (TradeTypeID == Sell)
}

func TradeTypeIsSharesIn(TradeTypeID uint) bool {
	return (TradeTypeID == SharesIn)
}

func TradeTypeIsSharesOut(TradeTypeID uint) bool {
	return (TradeTypeID == SharesOut)
}

func TradeTypeIsSplit(TradeTypeID uint) bool {
	return (TradeTypeID == Split)
}

func TradeTypeToCashFlowType(TradeTypeID uint) uint {
	var cType uint

	if TradeTypeIsReinvest(TradeTypeID) {
		cType = Credit
	} else {
		switch TradeTypeID {
		case Buy:
			cType = Debit
		case Sell:
			fallthrough
		case Dividend:
			fallthrough
		case Distribution:
			cType = Credit
		}
	}

	return cType
}

func actionToTradeType(action qif.InvestmentAction) uint {
	switch action {
		case qif.ActionBuy:
			return Buy
		case qif.ActionSell:
			return Sell
		case qif.ActionIntInc:
			fallthrough
		case qif.ActionDiv:
			return Dividend
		case qif.ActionCGLong:
			fallthrough
		case qif.ActionCGMid:
			fallthrough
		case qif.ActionCGShort:
			return Distribution
		case qif.ActionReInvInt:
			fallthrough
		case qif.ActionReInvDiv:
			return ReinvestedDividend
		case qif.ActionReInvLg:
			fallthrough
		case qif.ActionReInvMd:
			fallthrough
		case qif.ActionReInvSh:
			return ReinvestedDistribution
		case qif.ActionStockSplit:
			return Split
		case qif.ActionSharesOut:
			return SharesOut
		case qif.ActionSharesIn:
			return SharesIn
	}

	return UndefinedTradeType
}

func (*TradeType) List(db *gorm.DB) []TradeType {
	// need userCache lookup
	entries := []TradeType{}
	db.Find(&entries)

	return entries
}

func (*TradeType) ListBuys(db *gorm.DB) []TradeType {
	// need userCache lookup
	entries := []TradeType{}
	db.Where(listBuyTypes).Find(&entries)

	return entries
}
