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

func CreateMutualFund(t *testing.T, s *model.Security, name string) error {
	s.Company.Name = name
	s.SecurityBasisTypeID = model.BasisAverage
	s.SecurityTypeID = model.MutualFund

	return s.Create(defaultSession)
}

func TestCreateSecurity(t *testing.T) {
	a := model.GetAccountByName(defaultSession, "Gopher Financial")
	assert.Assert(t, a != nil)

	s := new(model.Security)
	s.AccountID = a.ID
	s.Company.Symbol = "GOAIX"
	err := CreateMutualFund(t, s, "Gopher AI Fund")
	assert.NilError(t, err)

	// should be able to fetch Security by Symbol only
	s1,_ := a.GetSecurityBySymbol(defaultSession, s.Company.Symbol)
	assert.Equal(t, s.ID, s1.ID)
}

func TestCreateSecurityNegative(t *testing.T) {
	a := model.GetAccountByName(defaultSession, "Gopher Checking")
	assert.Assert(t, a != nil)

	s := new(model.Security)
	s.AccountID = a.ID
	s.Company.Symbol = "GOFTX"
	err := CreateMutualFund(t, s, "Gopher Fintech Fund")
	assert.Assert(t, err != nil)
}

func TestCreateSecurityFromTrade(t *testing.T) {
	a := model.GetAccountByName(defaultSession, "Gopher Financial")
	assert.Assert(t, a != nil)

	tr := new(model.Trade)
	tr.AccountID = a.ID
	tr.Symbol = "BRK-B"
	tr.TradeTypeID = model.Buy
	tr.Date = time.Now()
	tr.Amount = decimal.NewFromInt32(1600)
	tr.Price = decimal.NewFromInt32(320)
	tr.Shares = decimal.NewFromInt32(5)

	err := tr.Create(defaultSession)
	assert.NilError(t, err)
}

func TestMoveSecurity(t *testing.T) {
	a := model.GetAccountByName(defaultSession, "Gopher Financial")
	assert.Assert(t, a != nil)
	aBalance := a.Balance

	b := model.GetAccountByName(defaultSession, "Gopher Investments")
	assert.Assert(t, b != nil)
	bBalance := b.Balance

	security,_ := a.GetSecurityBySymbol(defaultSession, "BRK-B")
	assert.Assert(t, security != nil)
	assert.Assert(t, security.ID != 0)

	security.ChangeAccount(defaultSession, b.Name)
	security = security.Find(security.ID)
	assert.Equal(t, security.AccountID, b.ID)
	aBalance = aBalance.Sub(security.Value)
	bBalance = bBalance.Add(security.Value)

	// get Accounts again which will refresh Balance from UserCache
	a = model.GetAccountByName(defaultSession, "Gopher Financial")
	b = model.GetAccountByName(defaultSession, "Gopher Investments")
	assert.Assert(t, a.Balance.Equal(aBalance))
	assert.Assert(t, b.Balance.Equal(bBalance))
}
