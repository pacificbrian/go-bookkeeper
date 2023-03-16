/*
 * SPDX-FileCopyrightText: 2023 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package controllers

import (
	"log"
	"net/http"
	"strconv"
	"github.com/labstack/echo/v4"
	"github.com/pacificbrian/go-bookkeeper/helpers"
	"github.com/pacificbrian/go-bookkeeper/model"
)

// For http.Status, see:
// https://go.dev/src/net/http/status.go

func ListTradeGains(c echo.Context) error {
	year, _ := strconv.Atoi(c.Param("year"))
	account_id, _ := strconv.Atoi(c.Param("account_id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Println("LIST GAINS (SELL TRADES)")
	get_json := false

	entry := new(model.Trade)
	entry.AccountID = uint(account_id)
	entry.Date = helpers.YearToDate(year)
	entries, totals := entry.ListByType(session, model.Sell, 0)

	if get_json {
		return c.JSON(http.StatusOK, entries)
	} else {
		data := map[string]any{ "account": &entry.Account,
					"trades": entries,
					"year": year,
					"total_gain": currency(totals[0]),
					"taxable_gain": currency(totals[1]) }
		return c.Render(http.StatusOK, "gains/index.html", data)
	}
}

func GetTradeGain(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("GET GAIN (SELL TRADE(%d))", id)
	get_json := false

	entry := new(model.Trade)
	entry.Model.ID = uint(id)
	entry = entry.Get(session)
	if entry == nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	if get_json {
		return c.JSON(http.StatusOK, entry)
	} else {
		var buys []model.TradeGain
		db := session.DB

		if entry != nil && entry.IsSell() {
			buys = entry.ListGains(db)
		}

		data := map[string]any{ "account": &entry.Account,
					"trade": entry,
					"gains": buys }
		return c.Render(http.StatusOK, "gains/show.html", data)
	}
}
