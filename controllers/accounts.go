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
	gormdb "go-bookkeeper/db"
	"go-bookkeeper/model"
	"go-bookkeeper/helpers"
)

// For http.Status, see:
// https://go.dev/src/net/http/status.go

// Need Access Controls
// Test with InPlace POST (AJAX is now FetchAPI ?)
// Use Controller objects
// Need Input Validation after Bind

func ListAccounts(c echo.Context) error {
	log.Println("LIST ACCOUNTS")
	db := gormdb.DbManager()
	get_json := false

	entries := model.ListAccounts(db)

	if get_json {
		return c.JSON(http.StatusOK, entries)
	} else {
		// Test if performance diff w/ map vs pongo2.context
		//data := pongo2.Context{ "accounts":entries }
		data := map[string]any{ "accounts":entries }
		return c.Render(http.StatusOK, "accounts/index.html", data)
	}
}

func CreateAccount(c echo.Context) error {
	log.Println("CREATE ACCOUNT")
	db := gormdb.DbManager()

	entry := new(model.Account)
	c.Bind(entry)
	entry.Create(db)
	// set status based on if Create failed

	return c.Redirect(http.StatusSeeOther, "/accounts")
}

func DeleteAccount(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("DELETE ACCOUNT(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.Account)
	entry.Model.ID = uint(id)
	if entry.Delete(db) != nil {
		return c.NoContent(http.StatusUnauthorized)
	} else {
		return c.NoContent(http.StatusAccepted)
	}
}

func GetAccount(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("GET ACCOUNT(%d)", id)
	db := gormdb.DebugDbManager()
	get_json := false

	// should be in Model
	entry := new(model.Account)
	entry.Model.ID = uint(id)
	entry = entry.Get(db, true)
	// set status based on if Get failed

	if get_json {
		return c.JSON(http.StatusOK, entry)
	} else {
		var cash_flows []model.CashFlow

		if entry != nil {
			// order by date
			cash_flows = new(model.CashFlow).List(db, entry)
		}

		dh := new(helpers.DateHelper)
		dh.Init()

		data := map[string]any{ "account":entry,
					"date_helper": dh,
					"button_text": "Add",
					"cash_flows": cash_flows,
					"total_amount": nil,
					"cash_flow_types": new(model.CashFlowType).List(db),
					"categories": new(model.Category).List(db) }
		return c.Render(http.StatusOK, "accounts/show.html", data)
	}
}

func UpdateAccount(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("UPDATE ACCOUNT(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.Account)
	entry.Model.ID = uint(id)
	entry = entry.Get(db, false)
	if entry == nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	c.Bind(entry)
	entry.Update(db)
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/accounts/%d", id))
}

func EditAccount(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("EDIT ACCOUNT(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.Account)
	entry.Model.ID = uint(id)
	entry = entry.Get(db, false)
	// handle no access (entry == nil)

	data := map[string]any{ "account": entry,
				"is_edit": true,
				"button_text": "Update Account",
				"account_types": new(model.AccountType).List(db),
				"currency_types": new(model.CurrencyType).List(db) }
	return c.Render(http.StatusOK, "accounts/edit.html", data)
}

func NewAccount(c echo.Context) error {
	log.Println("NEW ACCOUNT")
	db := gormdb.DbManager()

	data := map[string]any{ "account": new(model.Account).Init(),
				"button_text": "Create Account",
				"account_types": new(model.AccountType).List(db),
				"currency_types": new(model.CurrencyType).List(db) }
	return c.Render(http.StatusOK, "accounts/new.html", data)
}
