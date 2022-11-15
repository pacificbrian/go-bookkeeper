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

// SQL query string for all Buy, Sell types
var TradeTypeQueries = [3]string{"",
				 "trade_type_id = 1 OR trade_type_id = 5 OR trade_type_id = 6",
				 "trade_type_id = 2"}
var TradeTypeCashFlowsQuery string = "trade_type_id <= 4"

func TradeTypeIsBuy(TradeTypeID uint) bool {
	return (TradeTypeID == Buy || TradeTypeID == ReinvestedDividend ||
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

func (*TradeType) List(db *gorm.DB) []TradeType {
	// need userCache lookup
	entries := []TradeType{}
	db.Find(&entries)

	return entries
}
