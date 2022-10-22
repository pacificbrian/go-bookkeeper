/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package model

import (
	"log"
	"mime/multipart"
	"time"
	"github.com/davecgh/go-spew/spew"
	"github.com/shopspring/decimal"
)

type HttpFile struct {
	FileName string
	FileData multipart.File
}

type Model struct {
        ID        uint `gorm:"primaryKey"`
}

var useSpew bool = false

func forceSpewModel(data any) {
	spew.Dump(data)
}

func spewModel(data any) {
	if useSpew {
		forceSpewModel(data)
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
