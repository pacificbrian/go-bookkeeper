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

func TestSellTrade(t *testing.T) {
	tr := new(model.Trade)
	tr.AccountID = 1
	tr.Symbol = "GOOGL"
	tr.TradeTypeID = model.Buy
	tr.Date = time.Now()
	tr.Date = tr.Date.AddDate(0, 0, -7)
	tr.Amount = decimal.NewFromInt32(1250)
	basis := tr.Amount
	tr.Price = decimal.NewFromInt32(125)
	tr.Shares = decimal.NewFromInt32(10)

	// store Account balance before Trade
	a := new(model.Account)
	a = a.Find(tr.AccountID)
	balance := a.CashBalance

	err := tr.Create(defaultSession)
	assert.NilError(t, err)

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
	assert.Assert(t, gain.Equal(s.RetainedEarnings))
	assert.Assert(t, basis.Equal(s.AccumulatedBasis))
}
