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

func ListSecurities(c echo.Context) error {
	all, _ := strconv.Atoi(c.QueryParam("all"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	get_json := false

	account := new(model.Account)
	entries := new(model.Security).List(session, all == 0)
	log.Printf("LIST SECURITIES(%d) ALL(%d)", len(entries), all)

	if get_json {
		return c.JSON(http.StatusOK, entries)
	} else {
		var cashflows []model.CashFlow

		account.TotalPortfolio(entries)

		dh := new(helpers.DateHelper)
		dh.Init()

		data := map[string]any{ "securities": entries,
					"account": account,
					"date_helper": dh,
					"cash_flows": cashflows,
					"allSecurities": all > 0 }
		return c.Render(http.StatusOK, "securities/index.html", data)
	}
}

func CreateSecurity(c echo.Context) error {
	account_id, _ := strconv.Atoi(c.Param("account_id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}

	entry := new(model.Security)
	c.Bind(entry)
	c.Bind(&entry.Company)
	log.Printf("CREATE SECURITY NAME(%s) SYMBOL(%s)",
		   entry.Company.Name, entry.Company.Symbol)
	entry.AccountID = uint(account_id)
	err := entry.Create(session)
	if err != nil {
		log.Printf("CREATE SECURITY ACCOUNT(%d) FAILED: %v",
			   account_id, err)
		return c.NoContent(http.StatusUnauthorized)
	}
	return c.Redirect(http.StatusSeeOther,
			  fmt.Sprintf("/accounts/%d", account_id))
}

func DeleteSecurity(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("DELETE SECURITY(%d)", id)

	entry := new(model.Security)
	entry.ID = uint(id)
	if entry.Delete(session) != nil {
		return c.NoContent(http.StatusUnauthorized)
	} else {
		return c.NoContent(http.StatusAccepted)
	}
}

func GetSecurity(c echo.Context) error {
	debugParam, _ := strconv.Atoi(c.QueryParam("debug"))
	accountID, _ := strconv.Atoi(c.Param("account_id"))
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("GET ACCOUNT(%d) SECURITY(%d)", accountID, id)
	get_json := false

	entry := new(model.Security)
	entry.ID = uint(id)
	entry = entry.Get(session)
	if entry == nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	if get_json {
		return c.JSON(http.StatusOK, entry)
	} else {
		var trades []model.Trade
		var account *model.Account
		db := session.DB

		if entry != nil {
			account = &entry.Account
			trades = entry.ListTrades()
		}

		dh := new(helpers.DateHelper)
		dh.Init()

		data := map[string]any{ "security": entry,
					"account": account,
					"date_helper": dh,
					"trades": trades,
					"trade_types": new(model.TradeType).List(db),
					"debug_shares": debugParam > 0 }
		return c.Render(http.StatusOK, "securities/show.html", data)
	}
}

func NewSecurity(c echo.Context) error {
	account_id, _ := strconv.Atoi(c.Param("account_id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("NEW SECURITY ACCOUNT(%d)", account_id)

	entry := new(model.Security)
	entry.AccountID = uint(account_id)

	account := &entry.Account
	account.Model.ID = uint(account_id)
	account.Get(session, false)

	data := map[string]any{ "security": entry,
				"security_basis_types": new(model.SecurityBasisType).List(session.DB),
				"security_types": new(model.SecurityType).List(session.DB) }
	return c.Render(http.StatusOK, "securities/new.html", data)
}

func UpdateSecurity(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("UPDATE SECURITY(%d)", id)

	entry := new(model.Security)
	entry.ID = uint(id)
	entry = entry.Get(session)
	if entry == nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	err := c.Bind(entry)
	assert(err == nil, "UPDATE SECURITY BIND FAILED")
	c.Bind(&entry.Company)
	entry.Basis = getFormDecimal(c, "security.Basis")
	entry.Update()
	a_id := entry.AccountID
	return c.Redirect(http.StatusSeeOther,
			  fmt.Sprintf("/accounts/%d/securities/%d", a_id, id))
}

func EditSecurity(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("EDIT SECURITY(%d)", id)

	entry := new(model.Security)
	entry.ID = uint(id)
	entry = entry.Get(session)

	data := map[string]any{ "security": entry,
				"security_basis_types": new(model.SecurityBasisType).List(session.DB),
				"security_types": new(model.SecurityType).List(session.DB) }
	return c.Render(http.StatusOK, "securities/edit.html", data)
}

func MoveSecurity(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	newAccountName := ""

	isPost := (c.Request().Method == "POST")
	if isPost {
		newAccountName = c.FormValue("account.Name")
		log.Printf("MOVE SECURITY(%d) GET POST(%s)", id, newAccountName)
	} else {
		log.Printf("MOVE SECURITY(%d) GET", id)
	}

	entry := new(model.Security)
	entry.ID = uint(id)
	entry = entry.Get(session)
	if isPost && entry == nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	if !isPost {
		// GET
		data := map[string]any{ "security": entry }
		return c.Render(http.StatusOK, "securities/move.html", data)
	}

	// POST
	a := entry.ChangeAccount(session, newAccountName)
	if a == nil {
		return c.Redirect(http.StatusSeeOther,
				  fmt.Sprintf("/securities/%d/move", id))
	} else {
		return c.Redirect(http.StatusSeeOther,
				  fmt.Sprintf("/accounts/%d/securities/%d", a.ID, id))
	}
}
