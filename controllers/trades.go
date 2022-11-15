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

func CreateTrade(c echo.Context) error {
	account_id, _ := strconv.Atoi(c.Param("account_id"))
	security_id, _ := strconv.Atoi(c.Param("security_id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("CREATE TRADE ACCOUNT(%d) SECURITY(%d)",
		   account_id, security_id)

	entry := new(model.Trade)
	c.Bind(entry)
	entry.AccountID = uint(account_id)
	entry.SecurityID = uint(security_id)
	entry.Date = getFormDate(c)
	err := entry.Create(session)
	account_id = int(entry.AccountID)
	if err != nil {
		log.Printf("CREATE TRADE ACCOUNT(%d) SECURITY(%d) FAILED: %v",
			   account_id, security_id, err)
	}

	if security_id > 0 {
		return c.Redirect(http.StatusSeeOther,
				  fmt.Sprintf("/accounts/%d/securities/%d",
				  account_id, security_id))
	} else if account_id > 0 {
		return c.Redirect(http.StatusSeeOther,
				  fmt.Sprintf("/accounts/%d", account_id))
	} else {
		return c.NoContent(http.StatusAccepted)
	}
}

func DeleteTrade(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("DELETE TRADE(%d)", id)

	entry := new(model.Trade)
	entry.ID = uint(id)
	if entry.Delete(session) != nil {
		return c.NoContent(http.StatusUnauthorized)
	} else {
		return c.NoContent(http.StatusAccepted)
	}
}

func UpdateTrade(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	security_id, _ := strconv.Atoi(c.Param("security_id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("UPDATE TRADE(%d)", id)

	entry := new(model.Trade)
	entry.ID = uint(id)
	entry = entry.Get(session)
	if entry == nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	c.Bind(entry)
	entry.Date = getFormDate(c)
	entry.Update(session)
	a_id := entry.AccountID
	s_id := entry.SecurityID
	if security_id > 0 {
		return c.Redirect(http.StatusSeeOther,
				  fmt.Sprintf("/accounts/%d/securities/%d", a_id, s_id))
	} else {
		return c.Redirect(http.StatusSeeOther,
				  fmt.Sprintf("/accounts/%d", a_id))
	}
}

func EditTrade(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("EDIT TRADE(%d)", id)

	entry := new(model.Trade)
	entry.ID = uint(id)
	entry = entry.Get(session)

	dh := new(helpers.DateHelper)
	dh.Init()
	dh.SetDate(entry.Date)

	data := map[string]any{ "trade": entry,
				"date_helper": dh,
				"trade_types": new(model.TradeType).List(session.DB) }
	return c.Render(http.StatusOK, "trades/edit.html", data)
}
