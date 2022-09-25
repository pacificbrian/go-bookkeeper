/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ModelWithDelete struct {
        ID        uint `gorm:"primaryKey"`
        DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Model struct {
        ID        uint `gorm:"primaryKey"`
        //DeletedAt gorm.DeletedAt `gorm:"index"`
}

var useSpew bool = false

func spewModel(data any) {
	if useSpew {
		spew.Dump(data)
	}
}

func currency(value decimal.Decimal) string {
	return  "$" + value.StringFixedBank(2)
}
