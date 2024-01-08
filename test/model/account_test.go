/*
 * SPDX-FileCopyrightText: 2023 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model_test

import (
	"testing"
	"github.com/pacificbrian/go-bookkeeper/model"
	"gotest.tools/v3/assert"
)

func TestCreateAccount(t *testing.T) {
	a1 := new(model.Account)
	a1.Name = "Gopher Checking"
	err := a1.Create(defaultSession)
	assert.NilError(t, err)

	a2 := new(model.Account)
	a2.Name = "Gopher Credit"
	a2.AccountTypeID = model.AccountTypeCreditCard
	err = a2.Create(defaultSession)
	assert.NilError(t, err)

	a3 := new(model.Account)
	a3.Name = "Gopher Financial"
	a3.AccountTypeID = model.AccountTypeInvestment
	err = a3.Create(defaultSession)
	assert.NilError(t, err)

	a4 := new(model.Account)
	a4.Name = "Gopher Investments"
	a4.AccountTypeID = model.AccountTypeInvestment
	err = a4.Create(defaultSession)
	assert.NilError(t, err)
}
