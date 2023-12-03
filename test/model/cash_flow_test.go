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

// for controller testing, consider to make echo.Context
//e := echo.New()
//c := e.NewContext(
//	httptest.NewRequest(echo.GET, "/", nil)
//	httptest.NewRecorder())

func compareCashFlows(a *model.CashFlow, b *model.CashFlow, haveNames bool) bool {
	return !(a == nil || b == nil ||
		 (haveNames && a.PayeeName != b.PayeeName) ||
		 (haveNames && a.CategoryName != b.CategoryName) ||
		 !a.Amount.Equal(b.Amount))
}

func TestCreateCashFlow(t *testing.T) {
	c := new(model.CashFlow)
	c.AccountID = 1
	c.Date = time.Now()
	c.PayeeName = "Gopher Construction"
	c.CashFlowTypeID = model.Debit
	c.Amount = decimal.NewFromInt32(35)
	c.CategoryName = "Home:Improvement"
	c.CategoryID = model.CategoryGetByName(c.CategoryName).ID

	err := c.Create(defaultSession)
	assert.NilError(t, err)
	id := c.ID

	verify := new(model.CashFlow)
	verify.Model.ID = id
	verify = verify.Get(defaultSession, true)
	c.Amount = c.Amount.Abs() // because of edit=true above
	assert.Assert(t, compareCashFlows(c, verify, true))
}

func TestUpdateCashFlow(t *testing.T) {
	var id uint = 1

	edit := new(model.CashFlow)
	edit.Model.ID = id
	edit = edit.Get(defaultSession, false)
	edit.Amount = decimal.NewFromInt32(45)

	err := edit.Update()
	assert.NilError(t, err)

	verify := new(model.CashFlow)
	verify.Model.ID = uint(id)
	verify = verify.Get(defaultSession, false)
	assert.Assert(t, compareCashFlows(edit, verify, false))
}
