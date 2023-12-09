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
	"strings"
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

const timeFormatCompare string = "2006-01-02 15:04:05 -0700"
const timeFormatPrint string = "2006-01-02 15:04:05"
const dateFormatPrint string = "2006-01-02"

var useSpew bool = false

func forceSpewModel(data any, depth int) {
	spew.Config.MaxDepth = depth
	spew.Dump(data)
	spew.Config.MaxDepth = 0
}

func spewModel(data any) {
	if useSpew {
		forceSpewModel(data, 0)
	}
}

func assert(assertion bool, panicString string) {
	if (!assertion) {
		log.Panic(panicString)
	}
}

// Trim all spaces
func sanitizeString(input *string) {
	*input = strings.Join(strings.Fields(*input), " ")
}

func compareDates(dx *time.Time, dy *time.Time) bool {
	return dx.Format(timeFormatCompare) == dy.Format(timeFormatCompare)
}

func currency(value decimal.Decimal) string {
	return  "$" + value.StringFixedBank(2)
}

func dateFirst(a *time.Time, b *time.Time, descending bool) bool {
	if descending {
		return a.After(*b)
	}
	return b.After(*a)
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
			   days, timeToString(to), timeToString(from))
	}
	return days
}

func daysInMonth(dx *time.Time) uint {
	t := time.Date(dx.Year(), dx.Month(), 32, 0, 0, 0, 0, time.Local)
	return uint(32 - t.Day())
}

func decimalToPercentage(num decimal.Decimal) decimal.Decimal {
	return num.Mul(decimal.NewFromInt32(100))
}

func durationDays(d time.Duration) int32 {
	return int32(d.Hours()) / 24
}

func timeToString(dx *time.Time) string {
	return dx.Format(timeFormatPrint)
}

func dateToString(dx *time.Time) string {
	return dx.Format(dateFormatPrint)
}

func yearToDate(year int) time.Time {
	return time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
}
