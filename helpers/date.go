/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package helpers

import (
	"time"
)

// Index are 1-based
type DateHelper struct {
	Years [8]int
	Days [31]int
	Months []string
	YearIndex int
	MonthIndex int
	DayIndex int
}

func YearToDate(year int) time.Time {
	return time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
}

func (dh *DateHelper) Year() int {
	return dh.Years[dh.YearIndex - 1]
}

func (dh *DateHelper) SetYear(year int) {
	dh.YearIndex = 4
	for i := 0; i < 8; i++ {
		dh.Years[i] = year - (dh.YearIndex - 1) + i;
	}
}

func (dh *DateHelper) SetDate(date time.Time) {
	year,month,day := date.Date()

	dh.SetYear(year)
	dh.MonthIndex = int(month)
	dh.DayIndex = day
}

func (dh *DateHelper) Init() {
	dh.Months = []string{ "January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December" }

	for i := 0; i < 31; i++ {
		dh.Days[i] = i+1
	}

	dh.SetDate(time.Now())
}
