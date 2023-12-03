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
	a := new(model.Account)
	a.Name = "Gopher Checking"
	err := a.Create(defaultSession)
	assert.NilError(t, err)
}
