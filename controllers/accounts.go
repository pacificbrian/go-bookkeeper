/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"github.com/labstack/echo/v4"
	"github.com/pacificbrian/go-bookkeeper/model"
	"github.com/pacificbrian/go-bookkeeper/helpers"
)

// For http.Status, see:
// https://go.dev/src/net/http/status.go

// Need Access Controls
// Test with InPlace POST (AJAX is now FetchAPI ?)
// Use Controller objects
// Need Input Validation after Bind

func ListAccounts(c echo.Context) error {
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Println("LIST ACCOUNTS")
	get_json := false

	entries := model.ListAccounts(session, false)

	dh := new(helpers.DateHelper)
	dh.Init()

	if get_json {
		return c.JSON(http.StatusOK, entries)
	} else {
		// Test if performance diff w/ map vs pongo2.context
		//data := pongo2.Context{ "accounts":entries }
		data := map[string]any{ "accounts": entries,
					"date_helper": dh }
		return c.Render(http.StatusOK, "accounts/index.html", data)
	}
}

func CreateAccount(c echo.Context) error {
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Println("CREATE ACCOUNT")

	entry := new(model.Account)
	c.Bind(entry)
	entry.Create(session)
	// set status based on if Create failed

	return c.Redirect(http.StatusSeeOther, "/accounts")
}

func DeleteAccount(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("DELETE ACCOUNT(%d)", id)

	entry := new(model.Account)
	entry.Model.ID = uint(id)
	if entry.Delete(session) != nil {
		return c.NoContent(http.StatusUnauthorized)
	} else {
		return c.NoContent(http.StatusAccepted)
	}
}

func GetAccount(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	all, _ := strconv.Atoi(c.QueryParam("all"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("GET ACCOUNT(%d) ALL(%d)", id, all)
	get_json := false
	debugAB := false

	// should be in Model
	entry := new(model.Account)
	entry.Model.ID = uint(id)
	entry = entry.Get(session, true)
	if entry == nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	if get_json {
		return c.JSON(http.StatusOK, entry)
	} else {
		var cashflows []model.CashFlow
		var securities []model.Security
		var tradeTypes []model.TradeType
		db := session.DB

		if entry != nil {
			// List will order returned results
			if entry.IsInvestment() {
				securities = new(model.Security).List(session, entry, all == 0)
				cashflows = new(model.Trade).ListCashFlows(db, entry)
				tradeTypes = new(model.TradeType).List(db)
				entry.TotalPortfolio(securities)
			}
			cashflows = new(model.CashFlow).ListMerge(db, entry, cashflows)
		}

		if debugAB {
			entry.SetAverageDailyBalance(session)
		}

		dh := new(helpers.DateHelper)
		dh.Init()

		data := map[string]any{ "account": entry,
					"date_helper": dh,
					"button_text": "Add CashFlow",
					"cash_flows": cashflows,
					"securities": securities,
					"allSecurities": all > 0,
					"total_amount": nil,
					"cash_flow_types": new(model.CashFlowType).List(db),
					"categories": new(model.Category).List(db),
					"trade_types": tradeTypes }
		return c.Render(http.StatusOK, "accounts/show.html", data)
	}
}

func UpdateAccount(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("UPDATE ACCOUNT(%d)", id)

	entry := new(model.Account)
	entry.Model.ID = uint(id)
	entry = entry.Get(session, false)
	if entry == nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	entry.ClearBooleans()
	c.Bind(entry)
	entry.Update()
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/accounts/%d", id))
}

func EditAccount(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("EDIT ACCOUNT(%d)", id)
	db := session.DB

	entry := new(model.Account)
	entry.Model.ID = uint(id)
	entry = entry.Get(session, false)
	// handle no access (entry == nil)

	data := map[string]any{ "account": entry,
				"is_edit": true,
				"button_text": "Update Account",
				"account_types": new(model.AccountType).List(db),
				"currency_types": new(model.CurrencyType).List(db) }
	return c.Render(http.StatusOK, "accounts/edit.html", data)
}

func NewAccount(c echo.Context) error {
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Println("NEW ACCOUNT")
	db := session.DB

	data := map[string]any{ "account": new(model.Account).Init(),
				"button_text": "Create Account",
				"account_types": new(model.AccountType).List(db),
				"currency_types": new(model.CurrencyType).List(db) }
	return c.Render(http.StatusOK, "accounts/new.html", data)
}
