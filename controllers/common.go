/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package controllers

import (
	"time"
	"github.com/labstack/echo/v4"
)

const timeFormatPrint string = "2006-01-02 08:00:00"

type PutKeyValue struct {
	Key string `json:"key"`
	Value string `json:"value"`
}

func getFormDate(c echo.Context) time.Time {
	dateStr := c.FormValue("date_month") + "/" +
		c.FormValue("date_day") + "/" +
		c.FormValue("date_year")
	// local TZ and add 8 hours for sanity
	date, _ := time.ParseInLocation("1/2/2006", dateStr, time.Local)
	return date.Add(time.Hour * 8)
}

func timeToString(dx time.Time) string {
	return dx.Format(timeFormatPrint)
}
