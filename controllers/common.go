/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package controllers

import (
	"log"
	"strings"
	"time"
	"github.com/shopspring/decimal"
	"github.com/labstack/echo/v4"
)

const timeFormatPrint string = "2006-01-02 15:04:05"

type PutKeyValue struct {
	Key string `json:"key"`
	Value string `json:"value"`
}

func assert(assertion bool, panicString string) {
	if (!assertion) {
		log.Panic(panicString)
	}
}

func currency(value decimal.Decimal) string {
	return  "$" + value.StringFixedBank(2)
}

func getFormDate(c echo.Context) time.Time {
	dateStr := c.FormValue("date_month") + "/" +
		c.FormValue("date_day") + "/" +
		c.FormValue("date_year")
	// local TZ and add 8 hours for sanity
	date, _ := time.ParseInLocation("1/2/2006", dateStr, time.Local)
	return date.Add(time.Hour * 8)
}

// echo.Bind will try type.UnmarshalParam(), but we cannot
// define for lon-local types
func getFormDecimal(c echo.Context, name string) decimal.Decimal {
	valueStr := c.FormValue(name)
	valueStr = strings.Trim(valueStr, "$ ")
	valueStr = strings.Replace(valueStr, ",", "", -1)
	value,_ := decimal.NewFromString(valueStr)
	return value
}

func timeToString(dx time.Time) string {
	return dx.Format(timeFormatPrint)
}
