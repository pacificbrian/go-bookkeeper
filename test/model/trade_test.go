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

func makeTrade(tr *model.Trade, symbol string, daysAgo int, price int32, shares int32) {
	tr.Symbol = symbol
	tr.TradeTypeID = model.Buy
	tr.Date = time.Now()
	tr.Date = tr.Date.AddDate(0, 0, daysAgo)
	tr.Price = decimal.NewFromInt32(price)
	tr.Shares = decimal.NewFromInt32(shares)
	tr.Amount = tr.Shares.Mul(tr.Price)
}

func TestSellTrade(t *testing.T) {
	basis := decimal.Zero
	tr := new(model.Trade)
	tr.AccountID = 1

	// store Account balance before Trade
	a := new(model.Account)
	a = a.Find(tr.AccountID)
	balance := a.CashBalance

	makeTrade(tr, "GOOGL", -7, 125, 10)
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
}

func TestSellTradeAverageBasis(t *testing.T) {
	sharesHeld := decimal.Zero
	basis := decimal.Zero

	s := new(model.Security)
	s.AccountID = 1
	CreateMutualFund(t, s, "Gopher Growth Fund")

	tr := new(model.Trade)
	tr.AccountID = 1
	tr.SecurityID = s.ID

	// store Account balance before Trade
	a := new(model.Account)
	a = a.Find(tr.AccountID)
	balance := a.CashBalance

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
