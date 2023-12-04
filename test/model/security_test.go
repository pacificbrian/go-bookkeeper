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

func TestCreateSecurity(t *testing.T) {
	s := new(model.Security)
	s.AccountID = 1
	s.Company.Name = "Gopher AI Fund"
	s.Company.Symbol = "GOAIX"
	s.SecurityBasisTypeID = model.BasisAverage
	s.SecurityTypeID = model.MutualFund

	err := s.Create(defaultSession)
	assert.NilError(t, err)
}

func TestCreateSecurityFromTrade(t *testing.T) {
	tr := new(model.Trade)
	tr.AccountID = 1
	tr.Symbol = "BRK-B"
	tr.TradeTypeID = model.Buy
	tr.Date = time.Now()
	tr.Amount = decimal.NewFromInt32(1600)
	tr.Price = decimal.NewFromInt32(320)
	tr.Shares = decimal.NewFromInt32(5)

	err := tr.Create(defaultSession)
	assert.NilError(t, err)
}