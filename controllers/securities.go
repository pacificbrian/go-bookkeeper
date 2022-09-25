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

func ListSecurities(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Println("LIST SECURITIES")
	db := gormdb.DbManager()
	get_json := false

	var entries []model.Security
	entry := new(model.Security)
	entry.AccountID = uint(id)
	entries = entry.List(db)

	if get_json {
		return c.JSON(http.StatusOK, entries)
	} else {
		data := map[string]any{ "securities":entries,
					"account":&entry.Account }
		return c.Render(http.StatusOK, "securities/index.html", data)
	}
}

func CreateSecurity(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Println("CREATE SECURITY")
	db := gormdb.DbManager()

	entry := new(model.Security)
	c.Bind(entry)
	entry.AccountID = uint(id)
	entry.Create(db)

	// http.StatusCreated
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/accounts/%d", id))
}

func DeleteSecurity(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("DELETE SECURITY(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.Security)
	entry.ID = uint(id)
	if entry.Delete(db) != nil {
		return c.NoContent(http.StatusUnauthorized)
	} else {
		return c.NoContent(http.StatusAccepted)
	}
}

func GetSecurity(c echo.Context) error {
	accountID, _ := strconv.Atoi(c.Param("account_id"))
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("GET ACCOUNT(%d) SECURITY(%d)", accountID, id)
	db := gormdb.DebugDbManager()
	get_json := false

	entry := new(model.Security)
	entry.ID = uint(id)
	entry = entry.Get(db)
	// set status based on if Get failed

	if get_json {
		return c.JSON(http.StatusOK, entry)
	} else {
		var trades []model.Trade
		var account *model.Account

		if entry != nil {
			account = &entry.Account
			trades = new(model.Trade).List(db, account)
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
	log.Printf("UPDATE SECURITY(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.Security)
	entry.ID = uint(id)
	entry = entry.Get(db)
	if entry == nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	c.Bind(entry)
	entry.Update(db)
	a_id := entry.AccountID
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/accounts/%d", a_id))
	//return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/security/%d", id))
}

func EditSecurity(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("EDIT SECURITY(%d)", id)
	db := gormdb.DbManager()

	entry := new(model.Security)
	entry.ID = uint(id)
	entry = entry.Get(db)

	data := map[string]any{ "security": entry,
				"security_types": new(model.SecurityType).List(db) }
	return c.Render(http.StatusOK, "security/edit.html", data)
}
