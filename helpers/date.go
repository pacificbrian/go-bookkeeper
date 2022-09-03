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

func (dh *DateHelper) Init() {
	year,month,day := time.Now().Date()

	dh.YearIndex = 4
	for i := 0; i < 8; i++ {
		dh.Years[i] = year - (dh.YearIndex - 1) + i;
	}

	dh.Months = []string{ "January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December" }
	dh.MonthIndex = int(month)

	for i := 0; i < 31; i++ {
		dh.Days[i] = i+1
	}
	dh.DayIndex = day
}
