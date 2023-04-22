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
)

// For http.Status, see:
// https://go.dev/src/net/http/status.go

func ListPayees(c echo.Context) error {
	usage, _ := strconv.Atoi(c.QueryParam("usage"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("LIST PAYEES USAGE(%d)", usage)
	get_json := false

	entries := new(model.Payee).List(session)

	if get_json {
		return c.JSON(http.StatusOK, entries)
	} else {
		data := map[string]any{ "payees": entries,
					"account": nil,
					"show_use_count": usage > 0,
					"categories": new(model.Category).List(session.DB) }
		return c.Render(http.StatusOK, "payees/index.html", data)
	}
}

func CreatePayee(c echo.Context) error {
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Println("CREATE PAYEE")

	entry := new(model.Payee)
	c.Bind(entry)
	entry.Create(session)

	// http.StatusCreated
	return c.Redirect(http.StatusSeeOther, "/payees")
}

func DeletePayee(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("DELETE PAYEE(%d)", id)

	entry := new(model.Payee)
	entry.ID = uint(id)
	if entry.Delete(session) != nil {
		return c.NoContent(http.StatusUnauthorized)
	} else {
		return c.NoContent(http.StatusAccepted)
	}
}

func GetPayee(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("GET PAYEE(%d)", id)
	get_json := false

	entry := new(model.Payee)
	entry.ID = uint(id)
	entry = entry.Get(session)
	if entry == nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	if get_json {
		return c.JSON(http.StatusOK, entry)
	} else {
		var cash_flows []model.CashFlow

		cash_flows = entry.ListCashFlows()
		data := map[string]any{ "payee": entry,
					"account": nil,
					"disallow_cashflow_delete": true,
					"no_cashflow_balance": true,
					"with_cashflow_account": true,
					"cash_flows": cash_flows,
					"categories": new(model.Category).List(session.DB) }
		return c.Render(http.StatusOK, "payees/show.html", data)
	}
}

func UpdatePayee(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("UPDATE PAYEE(%d)", id)

	entry := new(model.Payee)
	entry.ID = uint(id)
	entry = entry.Get(session)
	if entry == nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	c.Bind(entry)
	entry.Update()
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/payees/%d", id))
}

func EditPayee(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("EDIT PAYEE(%d)", id)

	entry := new(model.Payee)
	entry.ID = uint(id)
	entry = entry.Get(session)

	data := map[string]any{ "payee": entry,
				"account": nil,
				"categories": new(model.Category).List(session.DB) }
	return c.Render(http.StatusOK, "payees/edit.html", data)
}
