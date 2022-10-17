/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"log"
	"time"
	"github.com/davecgh/go-spew/spew"
	"github.com/shopspring/decimal"
)

type Model struct {
        ID        uint `gorm:"primaryKey"`
}

var useSpew bool = false

func spewModel(data any) {
	if useSpew {
		spew.Dump(data)
	}
}

func assert(assertion bool, panicString string) {
	if (!assertion) {
		log.Panic(panicString)
	}
}

func currency(value decimal.Decimal) string {
	return  "$" + value.StringFixedBank(2)
}

func durationDays(d time.Duration) int32 {
	return int32(d.Hours()) / 24
}
