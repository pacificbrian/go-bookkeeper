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
	"strings"
	"github.com/labstack/echo/v4"
	"github.com/pacificbrian/go-bookkeeper/model"
)

func getAccount(session *model.Session, id uint) *model.Account {
	account := new(model.Account)
	if (id > 0) {
		account.Model.ID = id
		account = account.Get(session, false)
	}
	return account
}

func redirectToPayee(c echo.Context, id int, account_id int) error {
	if account_id > 0 {
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/accounts/%d/payees/%d",
								   account_id, id))
	} else {
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/payees/%d", id))
	}
}

// For http.Status, see:
// https://go.dev/src/net/http/status.go

func ListPayees(c echo.Context) error {
	usage, _ := strconv.Atoi(c.QueryParam("usage"))
	account_id, _ := strconv.Atoi(c.Param("account_id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	get_json := false

	var entries []model.Payee
	log.Printf("LIST ACCOUNT(%d) PAYEES, USAGE(%d)", account_id, usage)

	account := getAccount(session, uint(account_id))
	// account nil if access denied
	if account_id == 0 || account != nil {
		entries = new(model.Payee).List(session, account)
	}

	if get_json {
		return c.JSON(http.StatusOK, entries)
	} else {
		data := map[string]any{ "payees": entries,
					"account": account,
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
	account_id, _ := strconv.Atoi(c.Param("account_id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	log.Printf("GET ACCOUNT(%d) PAYEE(%d)", account_id, id)
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

		account := getAccount(session, uint(account_id))
		// account nil if access denied
		if account_id == 0 || account != nil {
			cash_flows = entry.ListCashFlows(account)
		}
		duplicate_payees := entry.GetDuplicates()
		data := map[string]any{ "payee": entry,
					"account": account,
					"disallow_cashflow_delete": true,
					"no_cashflow_balance": true,
					"with_cashflow_account": true,
					"cash_flows": cash_flows,
					"duplicate_payees": duplicate_payees,
					"categories": new(model.Category).List(session.DB) }
		return c.Render(http.StatusOK, "payees/show.html", data)
	}
}

func UpdatePayee(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	account_id, _ := strconv.Atoi(c.Param("account_id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	if id == 0 {
		return redirectToPayee(c, id, account_id)
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
	return redirectToPayee(c, id, account_id)
}

func MergePayee(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	account_id, _ := strconv.Atoi(c.Param("account_id"))
	merge_id, _ := strconv.Atoi(c.FormValue("payee.merge_id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	if id == 0 || merge_id == 0 {
		return redirectToPayee(c, id, account_id)
	}

	action := c.FormValue("submit")
	hasAll := strings.Contains(action, "All")
	log.Printf("MERGE PAYEE(%d) WITH(%d) ALL(%t)", id, merge_id, hasAll)

	entry := new(model.Payee)
	entry.ID = uint(id)
	entry = entry.Get(session)
	merge := new(model.Payee)
	if !hasAll {
		merge.ID = uint(merge_id)
		merge = merge.Get(session)
	}

	account := getAccount(session, uint(account_id))
	if entry == nil || merge == nil || account == nil {
		return c.NoContent(http.StatusUnauthorized)
	}
	entry.Merge(merge, account)

	return redirectToPayee(c, id, account_id)
}

func PayeeSetCategory(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	account_id, _ := strconv.Atoi(c.Param("account_id"))
	category_id, _ := strconv.Atoi(c.FormValue("payee.category_id"))
	session := getSession(c)
	if session == nil {
		return redirectToLogin(c)
	}
	if id == 0 || category_id == 0 {
		return redirectToPayee(c, id, account_id)
	}

	action := c.FormValue("submit")
	toAll := strings.Contains(action, "All")
	log.Printf("PAYEE(%d) SET_CATEGORY(%d) ALL(%t)", id, category_id, toAll)

	entry := new(model.Payee)
	entry.ID = uint(id)
	entry = entry.Get(session)
	account := getAccount(session, uint(account_id))
	if entry == nil || account == nil {
		return c.NoContent(http.StatusUnauthorized)
	}
	entry.SetCategory(account, uint(category_id), toAll)

	return redirectToPayee(c, id, account_id)
}

func EditPayee(c echo.Context) error {
	// TODO account_id
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
