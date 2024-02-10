/*
 * SPDX-FileCopyrightText: 2023 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model_test

import (
	"testing"
	"time"
	"github.com/shopspring/decimal"
	"github.com/pacificbrian/go-bookkeeper/model"
	"gotest.tools/v3/assert"
)

func makeTrade(tr *model.Trade, symbol string, daysAdd int, price int32, shares int32) {
	tr.Symbol = symbol
	tr.TradeTypeID = model.Buy
	tr.Date = time.Now()
	tr.Date = tr.Date.AddDate(0, 0, daysAdd)
	tr.Price = decimal.NewFromInt32(price)
	tr.Shares = decimal.NewFromInt32(shares)
	tr.Amount = tr.Shares.Mul(tr.Price)
}

func TestSellTrade(t *testing.T) {
	a := model.GetAccountByName(defaultSession, "Gopher Investments")
	assert.Assert(t, a != nil)
	// store Account balance before Trade
	balance := a.CashBalance
	basis := decimal.Zero

	tr := new(model.Trade)
	tr.AccountID = a.ID

	var held int = 7
	makeTrade(tr, "GOOGL", -held, 125, 10)
	err := tr.Create(defaultSession)
	assert.NilError(t, err)
	basis = basis.Add(tr.Amount)

	// check Account balance after Trade
	balance = balance.Sub(tr.Amount)
	a = a.Find(tr.AccountID)
	assert.Assert(t, balance.Equal(a.CashBalance))

	tr.TradeTypeID = model.Sell
	tr.Date = time.Now()
	tr.Amount = decimal.NewFromInt32(1350)
	tr.Price = decimal.NewFromInt32(135)

	err = tr.Create(defaultSession)
	assert.NilError(t, err)

	// check Account balance after Trade
	balance = balance.Add(tr.Amount)
	a = a.Find(tr.AccountID)
	assert.Assert(t, balance.Equal(a.CashBalance))

	// is gain correct?
	gain := tr.Amount.Sub(basis)
	assert.Assert(t, gain.Equal(tr.Gain))

	// check s.RetainedEarnings
	s := new(model.Security)
	s = s.Find(tr.SecurityID)
	assert.Assert(t, s.Basis.IsZero())
	assert.Assert(t, gain.Equal(s.RetainedEarnings))
	assert.Assert(t, basis.Equal(s.AccumulatedBasis))

	// test can update Sell Trade Date
	sell := tr.Find(tr.ID)
	sell.Date = sell.Date.AddDate(0, 0, 2)
	err = sell.Update()
	assert.NilError(t, err)
	held += 2

	tg := new(model.TradeGain)
	gains := tg.FindForSale(sell.ID)
	assert.Assert(t, len(gains) == 1)
	assert.Equal(t, int(gains[0].DaysHeld), held)
}

func TestSellTradeAverageBasis(t *testing.T) {
	a := model.GetAccountByName(defaultSession, "Gopher Financial")
	assert.Assert(t, a != nil)
	// store Account balance before Trade
	balance := a.CashBalance
	sharesHeld := decimal.Zero
	basis := decimal.Zero

	s := new(model.Security)
	s.AccountID = a.ID
	CreateMutualFund(t, s, "Gopher Growth Fund")

	tr := new(model.Trade)
	tr.AccountID = a.ID
	tr.SecurityID = s.ID

	makeTrade(tr, "", -14, 120, 10)
	err := tr.Create(defaultSession)
	assert.NilError(t, err)
	basis = basis.Add(tr.Amount)
	sharesHeld = sharesHeld.Add(tr.Shares)

	makeTrade(tr, "", -7, 140, 10)
	err = tr.Create(defaultSession)
	assert.NilError(t, err)
	basis = basis.Add(tr.Amount)
	sharesHeld = sharesHeld.Add(tr.Shares)

	// check Account balance after Trade
	balance = balance.Sub(basis)
	a = a.Find(tr.AccountID)
	assert.Assert(t, balance.Equal(a.CashBalance))

	tr.TradeTypeID = model.Sell
	tr.Date = time.Now()
	tr.Amount = decimal.NewFromInt32(1350)
	tr.Price = decimal.NewFromInt32(135)
	tr.Shares = decimal.NewFromInt32(10)

	err = tr.Create(defaultSession)
	assert.NilError(t, err)

	// check Account balance after Trade
	balance = balance.Add(tr.Amount)
	a = a.Find(tr.AccountID)
	assert.Assert(t, balance.Equal(a.CashBalance))

	// is gain correct?
	gainBasis := basis.Div(sharesHeld).Mul(tr.Shares).Round(2)
	gain := tr.Amount.Sub(gainBasis)
	assert.Assert(t, gain.Equal(tr.Gain))

	// check s.RetainedEarnings
	s = s.Find(tr.SecurityID)
	assert.Assert(t, basis.Sub(gainBasis).Equal(s.Basis))
	assert.Assert(t, gain.Equal(s.RetainedEarnings))
	assert.Assert(t, basis.Equal(s.AccumulatedBasis))
}
