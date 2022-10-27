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

const timeFormatCompare string = "2006-01-02 08:00:00 -0700"
const timeFormatPrint string = "2006-01-02 08:00:00"

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

func compareDates(dx *time.Time, dy *time.Time) bool {
	return dx.Format(timeFormatCompare) == dy.Format(timeFormatCompare)
}

func currency(value decimal.Decimal) string {
	return  "$" + value.StringFixedBank(2)
}

func daysBetweenDates(from *time.Time, to *time.Time, onlyBusinessDays bool) int32 {
	days := durationDays(to.Sub(*from))
	if !onlyBusinessDays {
		return int32(days)
	}

	debugEnabled := false
	numWholeWeeks :=  days / 7
	numWholeWeekdays := numWholeWeeks * 5

	extraWeekdays := int32(to.Weekday() - from.Weekday())
	if extraWeekdays > 0 {
		if to.Weekday() == time.Saturday {
		    extraWeekdays-- // we added 1 extra day
		}
	} else if extraWeekdays < 0 {
		extraWeekdays += 5
		if from.Weekday() == time.Saturday {
		    extraWeekdays++ // we subtracted 1 extra day
		}
	}

	days = int32(numWholeWeekdays+extraWeekdays)
	if debugEnabled {
		log.Printf("[HELPER] DAYS(%d) INBETWEEN: (%s) - (%s)",
			   days, to.Format(timeFormatPrint), from.Format(timeFormatPrint))
	}
	return days
}

func decimalToPercentage(num decimal.Decimal) decimal.Decimal {
	return num.Mul(decimal.NewFromInt32(100))
}

func durationDays(d time.Duration) int32 {
	return int32(d.Hours()) / 24
}
