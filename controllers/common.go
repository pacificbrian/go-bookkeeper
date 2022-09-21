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


func getFormDate(c echo.Context) time.Time {
	dateStr := c.FormValue("date_month") + "/" +
		c.FormValue("date_day") + "/" +
		c.FormValue("date_year")
	date, _ := time.Parse("1/2/2006", dateStr)
	return date
}
