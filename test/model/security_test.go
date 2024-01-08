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
