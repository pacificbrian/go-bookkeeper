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
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Println("LIST SECURITIES")
	get_json := false

	// need flag for all or open
	openOnly := false

	var entries []model.Security
	entry := new(model.Security)
	entry.AccountID = uint(id)
	entries = entry.List(session, nil, openOnly)

	if get_json {
		return c.JSON(http.StatusOK, entries)
	} else {
		data := map[string]any{ "securities": entries,
					"account": &entry.Account }
		return c.Render(http.StatusOK, "securities/index.html", data)
	}
}

func CreateSecurity(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Println("CREATE SECURITY")

	entry := new(model.Security)
	c.Bind(entry)
	entry.AccountID = uint(id)
	entry.Create(session)

	// http.StatusCreated
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/accounts/%d", id))
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
	// set status based on if Get failed

	if get_json {
		return c.JSON(http.StatusOK, entry)
	} else {
		var trades []model.Trade
		var account *model.Account
		db := session.DB

		if entry != nil {
			account = &entry.Account
			trades = entry.ListTrades(db, false)
		}

		dh := new(helpers.DateHelper)
		dh.Init()

		data := map[string]any{ "security": entry,
					"account": account,
					"date_helper": dh,
					"trades": trades,
					"trade_types": new(model.TradeType).List(db) }
		return c.Render(http.StatusOK, "securities/show.html", data)
	}
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

	c.Bind(entry)
	c.Bind(&entry.Company)
	entry.Update(session)
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
